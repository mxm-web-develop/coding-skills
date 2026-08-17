package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func runWork(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: flowctl work <create|list|show|ready|start|block|review-ready|complete|cancel>")
	}
	switch args[0] {
	case "create":
		return runWorkCreate(args[1:])
	case "list":
		return runWorkList(args[1:])
	case "show":
		return runWorkShow(args[1:])
	case "ready":
		return runWorkReady(args[1:])
	case "start":
		return runWorkStart(args[1:])
	case "block":
		return runWorkBlock(args[1:])
	case "review-ready":
		return runWorkReviewReady(args[1:])
	case "complete":
		return runWorkComplete(args[1:])
	case "cancel":
		return runWorkCancel(args[1:])
	default:
		return fmt.Errorf("unknown work command %q", args[0])
	}
}

func runWorkCreate(args []string) error {
	fs := flag.NewFlagSet("work create", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	title := fs.String("title", "", "work item title")
	kind := fs.String("kind", "feature", "feature, bug, refactor, test, docs, research, or chore")
	priority := fs.String("priority", "medium", "critical, high, medium, or low")
	goalID := fs.String("goal", "", "linked Goal ID")
	status := fs.String("status", "ready", "draft or ready")
	var requirements stringListFlag
	var acceptance stringListFlag
	var scope stringListFlag
	fs.Var(&requirements, "requirement", "linked Requirement ID; repeatable")
	fs.Var(&acceptance, "acceptance", "acceptance criterion; repeatable")
	fs.Var(&scope, "scope", "allowed path or component; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*title) == "" {
		return errors.New("--title is required")
	}
	if !contains([]string{"feature", "bug", "refactor", "test", "docs", "research", "chore"}, *kind) {
		return fmt.Errorf("invalid --kind: %s", *kind)
	}
	if !contains([]string{"critical", "high", "medium", "low"}, *priority) {
		return fmt.Errorf("invalid --priority: %s", *priority)
	}
	if *status != "draft" && *status != "ready" {
		return errors.New("--status must be draft or ready")
	}
	if len(acceptance) == 0 {
		return errors.New("at least one --acceptance is required")
	}
	if len(scope) == 0 {
		return errors.New("at least one --scope is required")
	}
	for _, id := range requirements {
		if err := requireObjectID(id, "REQ"); err != nil {
			return err
		}
	}
	var goal *string
	if *goalID != "" {
		if err := requireObjectID(*goalID, "GOAL"); err != nil {
			return err
		}
		goal = goalID
	}

	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	id, err := newObjectID("WI")
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item := WorkItem{
		SchemaVersion:      1,
		ID:                 id,
		Revision:           1,
		Kind:               *kind,
		Title:              strings.TrimSpace(*title),
		Status:             *status,
		Priority:           *priority,
		GoalID:             goal,
		RequirementIDs:     nonNil(requirements),
		AcceptanceCriteria: nonNil(acceptance),
		Scope:              nonNil(scope),
		Owner:              nil,
		RunID:              nil,
		EvidenceIDs:        []string{},
		BlockedReason:      nil,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := writeJSONAtomic(workItemPath(root, id), &item); err != nil {
		return err
	}
	if err := appendEvent(root, "work.created", "work-item", id, "", item.Revision, map[string]any{"status": item.Status}); err != nil {
		return err
	}
	fmt.Println(id)
	return nil
}

func runWorkList(args []string) error {
	fs := flag.NewFlagSet("work list", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	status := fs.String("status", "", "filter by status")
	jsonOutput := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	files, err := listJSONFiles(filepath.Join(root, ".ai-flow", "work-items"))
	if err != nil {
		return err
	}
	items := []WorkItem{}
	for _, path := range files {
		var item WorkItem
		if err := readJSON(path, &item); err != nil {
			return err
		}
		if *status == "" || item.Status == *status {
			items = append(items, item)
		}
	}
	if *jsonOutput {
		return printJSON(items)
	}
	for _, item := range items {
		fmt.Printf("%s\t%s\t%s\t%s\n", item.ID, item.Status, item.Priority, item.Title)
	}
	return nil
}

func runWorkShow(args []string) error {
	fs := flag.NewFlagSet("work show", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	id := fs.String("id", "", "Work Item ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	item, err := readWorkItem(root, *id)
	if err != nil {
		return err
	}
	return printJSON(item)
}

func runWorkReady(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work ready")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.Status != "draft" && item.Status != "blocked" {
		return fmt.Errorf("cannot mark work ready from status %s", item.Status)
	}
	item.Status = "ready"
	item.BlockedReason = nil
	return saveWorkMutation(root, &item, "work.ready", map[string]any{"status": item.Status})
}

func runWorkStart(args []string) error {
	fs := flag.NewFlagSet("work start", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	id := fs.String("id", "", "Work Item ID")
	owner := fs.String("owner", "", "writer identity")
	ttl := fs.Duration("ttl", 60*time.Minute, "lease duration")
	expected := fs.Int("expect-revision", 0, "optimistic revision")
	maxElapsed := fs.Int("max-elapsed-minutes", 120, "run elapsed-time budget")
	maxRetries := fs.Int("max-retries", 3, "run retry budget")
	maxFiles := fs.Int("max-changed-files", 50, "run changed-file budget")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*owner) == "" {
		return errors.New("--owner is required")
	}
	if *ttl <= 0 {
		return errors.New("--ttl must be positive")
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.Status != "ready" {
		return fmt.Errorf("cannot start work from status %s", item.Status)
	}
	if err := ensureLeaseAvailable(root, item.ID); err != nil {
		return err
	}
	runID, err := newObjectID("RUN")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	run := HarnessRun{
		SchemaVersion: 1,
		ID:            runID,
		Revision:      1,
		WorkItemID:    item.ID,
		Owner:         strings.TrimSpace(*owner),
		Status:        "running",
		Phase:         "implementing",
		GitSHA:        gitSHA(root),
		CheckpointIDs: []string{},
		EvidenceIDs:   []string{},
		Budgets: RunBudgets{
			MaxElapsedMinutes: *maxElapsed,
			MaxRetries:        *maxRetries,
			MaxChangedFiles:   *maxFiles,
		},
		StartedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
		CompletedAt: nil,
	}
	lease := WorkLease{
		SchemaVersion: 1,
		WorkItemID:    item.ID,
		RunID:         runID,
		Owner:         run.Owner,
		Scope:         item.Scope,
		AcquiredAt:    now.Format(time.RFC3339),
		ExpiresAt:     now.Add(*ttl).Format(time.RFC3339),
	}
	if err := writeJSONAtomic(runPath(root, runID), &run); err != nil {
		return err
	}
	if err := writeJSONAtomic(leasePath(root, item.ID), &lease); err != nil {
		return err
	}
	item.Status = "in_progress"
	item.Owner = &run.Owner
	item.RunID = &run.ID
	item.Revision++
	item.UpdatedAt = now.Format(time.RFC3339)
	if err := writeJSONAtomic(workItemPath(root, item.ID), &item); err != nil {
		return err
	}
	if err := appendEvent(root, "work.started", "work-item", item.ID, run.ID, item.Revision, map[string]any{"owner": run.Owner}); err != nil {
		return err
	}
	fmt.Println(runID)
	return nil
}

func runWorkBlock(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work block")
	reason := fs.String("reason", "", "blocking reason")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*reason) == "" {
		return errors.New("--reason is required")
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.Status == "done" || item.Status == "cancelled" {
		return fmt.Errorf("cannot block work from status %s", item.Status)
	}
	item.Status = "blocked"
	item.BlockedReason = pointer(strings.TrimSpace(*reason))
	if item.RunID != nil {
		run, readErr := readRun(root, *item.RunID)
		if readErr == nil {
			run.Status = "blocked"
			run.Revision++
			run.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
			if err := writeJSONAtomic(runPath(root, run.ID), &run); err != nil {
				return err
			}
		}
	}
	return saveWorkMutation(root, &item, "work.blocked", map[string]any{"reason": *reason})
}

func runWorkReviewReady(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work review-ready")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.Status != "in_progress" {
		return fmt.Errorf("cannot mark review-ready from status %s", item.Status)
	}
	item.Status = "ready_for_review"
	if item.RunID != nil {
		run, readErr := readRun(root, *item.RunID)
		if readErr != nil {
			return readErr
		}
		run.Status = "reviewing"
		run.Phase = "reviewing"
		run.Revision++
		run.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		if err := writeJSONAtomic(runPath(root, run.ID), &run); err != nil {
			return err
		}
	}
	return saveWorkMutation(root, &item, "work.review_ready", map[string]any{"status": item.Status})
}

func runWorkComplete(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work complete")
	var evidenceIDs stringListFlag
	fs.Var(&evidenceIDs, "evidence", "verified Evidence ID; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.Status != "in_progress" && item.Status != "ready_for_review" {
		return fmt.Errorf("cannot complete work from status %s", item.Status)
	}
	if len(evidenceIDs) == 0 {
		evidenceIDs = append(evidenceIDs, item.EvidenceIDs...)
	}
	if len(evidenceIDs) == 0 {
		return errors.New("at least one verified --evidence is required")
	}
	for _, evidenceID := range evidenceIDs {
		evidence, readErr := readEvidence(root, evidenceID)
		if readErr != nil {
			return readErr
		}
		if evidence.WorkItemID != item.ID {
			return fmt.Errorf("evidence %s belongs to %s", evidence.ID, evidence.WorkItemID)
		}
		if evidence.Result != "passed" || evidence.Trust == "unverified" {
			return fmt.Errorf("evidence %s is not a trusted pass", evidence.ID)
		}
		logPath := filepath.Join(root, filepath.FromSlash(evidence.LogPath))
		digest, hashErr := sha256File(logPath)
		if hashErr != nil {
			return fmt.Errorf("verify evidence %s: %w", evidence.ID, hashErr)
		}
		if digest != evidence.LogSHA256 {
			return fmt.Errorf("evidence %s log hash mismatch", evidence.ID)
		}
		item.EvidenceIDs = uniqueAppend(item.EvidenceIDs, evidence.ID)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item.Status = "done"
	if item.RunID != nil {
		run, readErr := readRun(root, *item.RunID)
		if readErr != nil {
			return readErr
		}
		run.Status = "completed"
		run.Phase = "completed"
		run.Revision++
		run.UpdatedAt = now
		run.CompletedAt = &now
		if err := writeJSONAtomic(runPath(root, run.ID), &run); err != nil {
			return err
		}
	}
	if err := os.Remove(leasePath(root, item.ID)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return saveWorkMutation(root, &item, "work.completed", map[string]any{"evidence_ids": item.EvidenceIDs})
}

func runWorkCancel(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work cancel")
	reason := fs.String("reason", "", "cancellation reason")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*reason) == "" {
		return errors.New("--reason is required")
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.Status == "done" {
		return errors.New("completed work cannot be cancelled")
	}
	item.Status = "cancelled"
	item.BlockedReason = pointer(strings.TrimSpace(*reason))
	if err := os.Remove(leasePath(root, item.ID)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return saveWorkMutation(root, &item, "work.cancelled", map[string]any{"reason": *reason})
}

func workMutationFlags(name string) (*flag.FlagSet, *string, *string, *int) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	id := fs.String("id", "", "Work Item ID")
	expected := fs.Int("expect-revision", 0, "optimistic revision")
	return fs, rootArg, id, expected
}

func loadWorkMutation(rootArg, id string, expected int) (string, WorkItem, error) {
	root, err := resolveRoot(rootArg, true)
	if err != nil {
		return "", WorkItem{}, err
	}
	item, err := readWorkItem(root, id)
	if err != nil {
		return "", item, err
	}
	if err := checkExpectedRevision(item.Revision, expected); err != nil {
		return "", item, err
	}
	return root, item, nil
}

func saveWorkMutation(root string, item *WorkItem, eventType string, data map[string]any) error {
	item.Revision++
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := writeJSONAtomic(workItemPath(root, item.ID), item); err != nil {
		return err
	}
	runID := ""
	if item.RunID != nil {
		runID = *item.RunID
	}
	if err := appendEvent(root, eventType, "work-item", item.ID, runID, item.Revision, data); err != nil {
		return err
	}
	return printJSON(item)
}

func ensureLeaseAvailable(root, workID string) error {
	path := leasePath(root, workID)
	var lease WorkLease
	if err := readJSON(path, &lease); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	expires, err := time.Parse(time.RFC3339, lease.ExpiresAt)
	if err != nil {
		return fmt.Errorf("invalid existing lease: %w", err)
	}
	if time.Now().UTC().Before(expires) {
		return fmt.Errorf("work item is leased by %s in %s until %s", lease.Owner, lease.RunID, lease.ExpiresAt)
	}
	return nil
}

func printJSON(value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func pointer(value string) *string { return &value }
