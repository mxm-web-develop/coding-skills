package main

import (
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func runCheckpoint(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: flowctl checkpoint <save|list|show|latest|resume>")
	}
	switch args[0] {
	case "save":
		return runCheckpointSave(args[1:])
	case "list":
		return runCheckpointList(args[1:])
	case "show":
		return runCheckpointShow(args[1:])
	case "latest":
		return runCheckpointLatest(args[1:])
	case "resume":
		return runCheckpointResume(args[1:])
	default:
		return fmt.Errorf("unknown checkpoint command %q", args[0])
	}
}

func runCheckpointSave(args []string) error {
	fs := flag.NewFlagSet("checkpoint save", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	runID := fs.String("run", "", "Run ID")
	phase := fs.String("phase", "", "current workflow phase")
	summary := fs.String("summary", "", "completed work summary")
	next := fs.String("next", "", "next resumable action")
	expected := fs.Int("expect-revision", 0, "optimistic run revision")
	var completed stringListFlag
	var changed stringListFlag
	var questions stringListFlag
	fs.Var(&completed, "completed", "completed step; repeatable")
	fs.Var(&changed, "changed-file", "changed file; repeatable")
	fs.Var(&questions, "question", "open question; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*phase) == "" || strings.TrimSpace(*summary) == "" || strings.TrimSpace(*next) == "" {
		return errors.New("--phase, --summary, and --next are required")
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	run, err := readRun(root, *runID)
	if err != nil {
		return err
	}
	if err := checkExpectedRevision(run.Revision, *expected); err != nil {
		return err
	}
	if contains([]string{"completed", "cancelled"}, run.Status) {
		return fmt.Errorf("cannot checkpoint run in status %s", run.Status)
	}
	id, err := newObjectID("CP")
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	checkpoint := Checkpoint{
		SchemaVersion:  1,
		ID:             id,
		RunID:          run.ID,
		WorkItemID:     run.WorkItemID,
		Sequence:       len(run.CheckpointIDs) + 1,
		Phase:          strings.TrimSpace(*phase),
		Summary:        strings.TrimSpace(*summary),
		NextAction:     strings.TrimSpace(*next),
		GitSHA:         gitSHA(root),
		CompletedSteps: nonNil(completed),
		ChangedFiles:   nonNil(changed),
		OpenQuestions:  nonNil(questions),
		CreatedAt:      now,
	}
	if err := writeJSONAtomic(checkpointPath(root, run.ID, checkpoint.ID), &checkpoint); err != nil {
		return err
	}
	run.CheckpointIDs = uniqueAppend(run.CheckpointIDs, checkpoint.ID)
	run.Status = "checkpointed"
	run.Phase = checkpoint.Phase
	run.Revision++
	run.UpdatedAt = now
	if err := writeJSONAtomic(runPath(root, run.ID), &run); err != nil {
		return err
	}
	if err := appendEvent(root, "checkpoint.saved", "checkpoint", checkpoint.ID, run.ID, run.Revision, map[string]any{"work_item_id": run.WorkItemID, "sequence": checkpoint.Sequence}); err != nil {
		return err
	}
	fmt.Println(checkpoint.ID)
	return nil
}

func runCheckpointList(args []string) error {
	fs := flag.NewFlagSet("checkpoint list", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	runID := fs.String("run", "", "Run ID")
	jsonOutput := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	checkpoints, err := listCheckpoints(root, *runID)
	if err != nil {
		return err
	}
	if *jsonOutput {
		return printJSON(checkpoints)
	}
	for _, checkpoint := range checkpoints {
		fmt.Printf("%s\t%s\t%d\t%s\t%s\n", checkpoint.ID, checkpoint.RunID, checkpoint.Sequence, checkpoint.Phase, checkpoint.NextAction)
	}
	return nil
}

func runCheckpointShow(args []string) error {
	fs := flag.NewFlagSet("checkpoint show", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	runID := fs.String("run", "", "Run ID")
	id := fs.String("id", "", "Checkpoint ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := requireObjectID(*runID, "RUN"); err != nil {
		return err
	}
	if err := requireObjectID(*id, "CP"); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	var checkpoint Checkpoint
	if err := readJSON(checkpointPath(root, *runID, *id), &checkpoint); err != nil {
		return err
	}
	return printJSON(checkpoint)
}

func runCheckpointLatest(args []string) error {
	fs := flag.NewFlagSet("checkpoint latest", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	runID := fs.String("run", "", "Run ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	checkpoint, err := latestCheckpoint(root, *runID)
	if err != nil {
		return err
	}
	return printJSON(checkpoint)
}

func runCheckpointResume(args []string) error {
	fs := flag.NewFlagSet("checkpoint resume", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	runID := fs.String("run", "", "Run ID")
	owner := fs.String("owner", "", "writer identity")
	handoffFrom := fs.String("handoff-from", "", "current owner identity when explicitly transferring this run")
	allowDrift := fs.Bool("allow-git-drift", false, "allow resuming from a different Git revision")
	ttl := fs.Duration("ttl", 60*time.Minute, "renewed lease duration")
	expected := fs.Int("expect-revision", 0, "optimistic run revision")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*owner) == "" {
		return errors.New("--owner is required")
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	run, err := readRun(root, *runID)
	if err != nil {
		return err
	}
	if err := checkExpectedRevision(run.Revision, *expected); err != nil {
		return err
	}
	newOwner := strings.TrimSpace(*owner)
	previousOwner := run.Owner
	if previousOwner != newOwner {
		if strings.TrimSpace(*handoffFrom) == "" {
			return fmt.Errorf("run is owned by %s; use --handoff-from %q for an explicit transfer", previousOwner, previousOwner)
		}
		if strings.TrimSpace(*handoffFrom) != previousOwner {
			return fmt.Errorf("handoff owner mismatch: run is owned by %s", previousOwner)
		}
		if !contains([]string{"checkpointed", "blocked"}, run.Status) {
			return fmt.Errorf("cannot transfer run in status %s; save a checkpoint first", run.Status)
		}
	}
	checkpoint, err := latestCheckpoint(root, run.ID)
	if err != nil {
		return err
	}
	currentSHA := gitSHA(root)
	if !*allowDrift && checkpoint.GitSHA != currentSHA {
		return fmt.Errorf("Git revision changed since checkpoint: checkpoint=%s current=%s", checkpoint.GitSHA, currentSHA)
	}
	item, err := readWorkItem(root, run.WorkItemID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	lease := WorkLease{
		SchemaVersion: 1,
		WorkItemID:    item.ID,
		RunID:         run.ID,
		Owner:         newOwner,
		Scope:         item.Scope,
		AcquiredAt:    now.Format(time.RFC3339),
		ExpiresAt:     now.Add(*ttl).Format(time.RFC3339),
	}
	if err := writeJSONAtomic(leasePath(root, item.ID), &lease); err != nil {
		return err
	}
	run.Status = "running"
	run.Owner = newOwner
	run.Revision++
	run.UpdatedAt = now.Format(time.RFC3339)
	if err := writeJSONAtomic(runPath(root, run.ID), &run); err != nil {
		return err
	}
	item.Status = "in_progress"
	item.BlockedReason = nil
	item.Owner = &newOwner
	item.RunID = &run.ID
	item.Revision++
	item.UpdatedAt = now.Format(time.RFC3339)
	if err := writeJSONAtomic(workItemPath(root, item.ID), &item); err != nil {
		return err
	}
	eventData := map[string]any{"checkpoint_id": checkpoint.ID}
	if previousOwner != newOwner {
		eventData["handoff_from"] = previousOwner
		eventData["handoff_to"] = newOwner
	}
	if err := appendEvent(root, "checkpoint.resumed", "run", run.ID, run.ID, run.Revision, eventData); err != nil {
		return err
	}
	return printJSON(checkpoint)
}

func listCheckpoints(root, runID string) ([]Checkpoint, error) {
	if err := requireObjectID(runID, "RUN"); err != nil {
		return nil, err
	}
	files, err := listJSONFiles(filepath.Join(root, ".ai-flow", "runs", runID, "checkpoints"))
	if err != nil {
		return nil, err
	}
	checkpoints := []Checkpoint{}
	for _, path := range files {
		var checkpoint Checkpoint
		if err := readJSON(path, &checkpoint); err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, checkpoint)
	}
	return checkpoints, nil
}

func latestCheckpoint(root, runID string) (Checkpoint, error) {
	checkpoints, err := listCheckpoints(root, runID)
	if err != nil {
		return Checkpoint{}, err
	}
	if len(checkpoints) == 0 {
		return Checkpoint{}, fmt.Errorf("run %s has no checkpoints", runID)
	}
	latest := checkpoints[0]
	for _, checkpoint := range checkpoints[1:] {
		if checkpoint.Sequence > latest.Sequence {
			latest = checkpoint
		}
	}
	return latest, nil
}
