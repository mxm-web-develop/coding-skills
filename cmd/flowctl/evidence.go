package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

func runEvidence(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: flowctl evidence <run|record|list|show|verify>")
	}
	switch args[0] {
	case "run":
		return runEvidenceCommand(args[1:])
	case "record":
		return runEvidenceRecord(args[1:])
	case "list":
		return runEvidenceList(args[1:])
	case "show":
		return runEvidenceShow(args[1:])
	case "verify":
		return runEvidenceVerify(args[1:])
	default:
		return fmt.Errorf("unknown evidence command %q", args[0])
	}
}

func runEvidenceCommand(args []string) error {
	fs := flag.NewFlagSet("evidence run", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	workID := fs.String("work", "", "Work Item ID")
	runID := fs.String("run", "", "Run ID")
	testID := fs.String("test", "", "Test ID or stable test name")
	quiet := fs.Bool("quiet", false, "write output only to the evidence log")
	if err := fs.Parse(args); err != nil {
		return err
	}
	commandArgs := fs.Args()
	if len(commandArgs) == 0 {
		return errors.New("command is required after --")
	}
	if strings.TrimSpace(*testID) == "" {
		return errors.New("--test is required")
	}
	root, item, run, err := loadEvidenceContext(*rootArg, *workID, *runID)
	if err != nil {
		return err
	}

	evidenceID, err := newObjectID("EV")
	if err != nil {
		return err
	}
	logDir := filepath.Join(root, ".ai-flow", "evidence", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	logPath := filepath.Join(logDir, evidenceID+".log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	command := exec.Command(commandArgs[0], commandArgs[1:]...)
	command.Dir = root
	var output io.Writer = logFile
	if !*quiet {
		output = io.MultiWriter(os.Stdout, logFile)
	}
	command.Stdout = output
	command.Stderr = output
	started := time.Now().UTC()
	runErr := command.Run()
	ended := time.Now().UTC()
	if closeErr := logFile.Close(); closeErr != nil && runErr == nil {
		runErr = closeErr
	}
	exitCode := 0
	result := "passed"
	if runErr != nil {
		result = "failed"
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(runErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	digest, err := sha256File(logPath)
	if err != nil {
		return err
	}
	relLog, _ := filepath.Rel(root, logPath)
	evidence := Evidence{
		SchemaVersion: 1,
		ID:            evidenceID,
		WorkItemID:    item.ID,
		RunID:         evidenceRunID(run),
		TestID:        strings.TrimSpace(*testID),
		Source:        "local",
		Trust:         "verified-local",
		Result:        result,
		Command:       commandArgs,
		ExitCode:      exitCode,
		GitSHA:        gitSHA(root),
		Environment: map[string]string{
			"os":   runtime.GOOS,
			"arch": runtime.GOARCH,
		},
		StartedAt:   started.Format(time.RFC3339),
		EndedAt:     ended.Format(time.RFC3339),
		LogPath:     filepath.ToSlash(relLog),
		LogSHA256:   digest,
		ExternalURI: nil,
		CreatedAt:   ended.Format(time.RFC3339),
	}
	if err := persistEvidence(root, &item, &run, &evidence); err != nil {
		return err
	}
	fmt.Println(evidence.ID)
	if runErr != nil {
		return fmt.Errorf("evidence command failed with exit code %d; evidence recorded as %s", exitCode, evidence.ID)
	}
	return nil
}

func runEvidenceRecord(args []string) error {
	fs := flag.NewFlagSet("evidence record", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	workID := fs.String("work", "", "Work Item ID")
	runID := fs.String("run", "", "Run ID (optional when source=agent-claim or mode=external)")
	testID := fs.String("test", "", "Test ID or stable test name")
	source := fs.String("source", "external", "external, ci, or agent-claim")
	mode := fs.String("mode", "run", "run ties this record to a harness run; external records a standalone evidence without a run")
	uri := fs.String("uri", "", "external evidence URI")
	description := fs.String("description", "", "claim or external evidence description")
	if err := fs.Parse(args); err != nil {
		return err
	}
	// source=local is produced by `evidence run`, not by `evidence record`.
	// Check this first so the error message is precise.
	if *source == "local" {
		return errors.New("--source=local is reserved for `evidence run`; use external, ci, or agent-claim here")
	}
	if !contains([]string{"external", "ci", "agent-claim"}, *source) {
		return errors.New("--source must be external, ci, or agent-claim")
	}
	if !contains([]string{"run", "external"}, *mode) {
		return errors.New("--mode must be run or external")
	}
	if strings.TrimSpace(*testID) == "" || strings.TrimSpace(*description) == "" {
		return errors.New("--test and --description are required")
	}
	// --run is required only when the source and mode both demand a harness run.
	requiresRun := *mode == "run" && *source != "agent-claim"
	if requiresRun && strings.TrimSpace(*runID) == "" {
		return errors.New("--run is required for this --source/--mode combination (use --mode=external or --source=agent-claim to record standalone evidence)")
	}
	if !requiresRun && strings.TrimSpace(*runID) != "" {
		return errors.New("--run is ignored in this --source/--mode combination; omit it")
	}
	root, item, run, err := loadEvidenceContext(*rootArg, *workID, *runID)
	if err != nil {
		return err
	}
	id, err := newObjectID("EV")
	if err != nil {
		return err
	}
	logDir := filepath.Join(root, ".ai-flow", "evidence", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}
	logPath := filepath.Join(logDir, id+".log")
	descriptor := fmt.Sprintf("source: %s\nuri: %s\ndescription: %s\n", *source, *uri, *description)
	if err := os.WriteFile(logPath, []byte(descriptor), 0o644); err != nil {
		return err
	}
	digest, err := sha256File(logPath)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	relLog, _ := filepath.Rel(root, logPath)
	var externalURI *string
	if strings.TrimSpace(*uri) != "" {
		externalURI = pointer(strings.TrimSpace(*uri))
	}
	evidence := Evidence{
		SchemaVersion: 1,
		ID:            id,
		WorkItemID:    item.ID,
		RunID:         evidenceRunID(run),
		TestID:        strings.TrimSpace(*testID),
		Source:        *source,
		Trust:         "unverified",
		Result:        "unverified",
		Command:       []string{"external-record"},
		ExitCode:      -1,
		GitSHA:        gitSHA(root),
		Environment:   map[string]string{"recorded_by": "flowctl"},
		StartedAt:     now,
		EndedAt:       now,
		LogPath:       filepath.ToSlash(relLog),
		LogSHA256:     digest,
		ExternalURI:   externalURI,
		CreatedAt:     now,
	}
	if err := persistEvidence(root, &item, &run, &evidence); err != nil {
		return err
	}
	fmt.Println(evidence.ID)
	return nil
}

func runEvidenceList(args []string) error {
	fs := flag.NewFlagSet("evidence list", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	workID := fs.String("work", "", "filter by Work Item ID")
	runID := fs.String("run", "", "filter by Run ID")
	result := fs.String("result", "", "filter by result")
	jsonOutput := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	files, err := listJSONFiles(filepath.Join(root, ".ai-flow", "evidence"))
	if err != nil {
		return err
	}
	items := []Evidence{}
	for _, path := range files {
		var evidence Evidence
		if err := readJSON(path, &evidence); err != nil {
			return err
		}
		evRunID := ""
		if evidence.RunID != nil {
			evRunID = *evidence.RunID
		}
		if (*workID == "" || evidence.WorkItemID == *workID) && (*runID == "" || evRunID == *runID) && (*result == "" || evidence.Result == *result) {
			items = append(items, evidence)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt < items[j].CreatedAt })
	if *jsonOutput {
		return printJSON(items)
	}
	for _, evidence := range items {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", evidence.ID, evidence.Result, evidence.Trust, evidence.TestID, evidence.GitSHA)
	}
	return nil
}

func runEvidenceShow(args []string) error {
	fs := flag.NewFlagSet("evidence show", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	id := fs.String("id", "", "Evidence ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	evidence, err := readEvidence(root, *id)
	if err != nil {
		return err
	}
	return printJSON(evidence)
}

func runEvidenceVerify(args []string) error {
	fs := flag.NewFlagSet("evidence verify", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	id := fs.String("id", "", "Evidence ID")
	requireCurrent := fs.Bool("require-current-git", false, "fail when evidence Git SHA is not current")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	evidence, err := readEvidence(root, *id)
	if err != nil {
		return err
	}
	logPath := filepath.Join(root, filepath.FromSlash(evidence.LogPath))
	digest, err := sha256File(logPath)
	if err != nil {
		return err
	}
	if digest != evidence.LogSHA256 {
		return fmt.Errorf("evidence log hash mismatch: expected %s, got %s", evidence.LogSHA256, digest)
	}
	currentSHA := gitSHA(root)
	gitCurrent := currentSHA == evidence.GitSHA
	if *requireCurrent && !gitCurrent {
		return fmt.Errorf("evidence Git SHA is stale: evidence=%s current=%s", evidence.GitSHA, currentSHA)
	}
	return printJSON(map[string]any{
		"id":               evidence.ID,
		"integrity":        "valid",
		"trust":            evidence.Trust,
		"result":           evidence.Result,
		"git_sha_current":  gitCurrent,
		"evidence_git_sha": evidence.GitSHA,
		"current_git_sha":  currentSHA,
	})
}

func loadEvidenceContext(rootArg, workID, runID string) (string, WorkItem, HarnessRun, error) {
	root, err := resolveRoot(rootArg, true)
	if err != nil {
		return "", WorkItem{}, HarnessRun{}, err
	}
	item, err := readWorkItem(root, workID)
	if err != nil {
		return "", item, HarnessRun{}, err
	}
	if strings.TrimSpace(runID) == "" {
		// Standalone evidence (agent-claim or --mode=external): no harness run.
		return root, item, HarnessRun{}, nil
	}
	run, err := readRun(root, runID)
	if err != nil {
		return "", item, run, err
	}
	if run.WorkItemID != item.ID {
		return "", item, run, fmt.Errorf("run %s belongs to %s, not %s", run.ID, run.WorkItemID, item.ID)
	}
	if contains([]string{"completed", "cancelled"}, run.Status) {
		return "", item, run, fmt.Errorf("cannot add evidence to run in status %s", run.Status)
	}
	return root, item, run, nil
}

func persistEvidence(root string, item *WorkItem, run *HarnessRun, evidence *Evidence) error {
	if err := writeJSONAtomic(evidencePath(root, evidence.ID), evidence); err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	item.EvidenceIDs = uniqueAppend(item.EvidenceIDs, evidence.ID)
	item.Revision++
	item.UpdatedAt = now
	if err := writeJSONAtomic(workItemPath(root, item.ID), item); err != nil {
		return err
	}
	// When the evidence is recorded without a harness run (agent-claim or
	// --mode=external), only the development task gets the back-reference and
	// no run state is touched.
	eventTarget := ""
	if run != nil && run.ID != "" {
		run.EvidenceIDs = uniqueAppend(run.EvidenceIDs, evidence.ID)
		run.Status = "verifying"
		run.Phase = "verifying"
		run.Revision++
		run.UpdatedAt = now
		if err := writeJSONAtomic(runPath(root, run.ID), run); err != nil {
			return err
		}
		eventTarget = run.ID
	}
	eventExtras := map[string]any{"result": evidence.Result, "trust": evidence.Trust, "work_item_id": item.ID}
	if run == nil || run.ID == "" {
		eventExtras["standalone"] = true
	}
	return appendEvent(root, "evidence.recorded", "evidence", evidence.ID, eventTarget, 1, eventExtras)
}


// evidenceRunID returns a pointer to the harness run's ID for evidence.RunID,
// or nil when the run is empty (standalone evidence: agent-claim or
// --mode=external). It is intentionally a thin wrapper so the nil branch is
// expressed at every callsite.
func evidenceRunID(run HarnessRun) *string {
	if run.ID == "" {
		return nil
	}
	id := run.ID
	return &id
}
