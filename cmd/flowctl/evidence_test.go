package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bootstrapEvidenceProject creates a minimal .ai-flow project root with one
// Work Item and an optional Harness Run. Returned IDs are stable so tests can
// cross-check back-references.
func bootstrapEvidenceProject(t *testing.T, withRun bool) (root string, workID string, runID string) {
	t.Helper()
	root = t.TempDir()
	for _, d := range []string{
		"state", "goals", "requirements", "plans", "work-items",
		"decisions", "tests", "evidence", "releases", "runs", "bin",
	} {
		if err := os.MkdirAll(filepath.Join(root, ".ai-flow", d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Stub runtime so flowctl's resolveRoot check passes. Tests do not shell
	// out to it; the file just needs to exist.
	stub := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(filepath.Join(root, ".ai-flow", "bin", "flowctl"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	workID = "WI-20260818-aaaabbbb"
	runID = "RUN-20260818-cccccccc"
	writeBoardTextFixture(t, root, ".ai-flow/state/project.json",
		`{"schema_version":1,"project_name":"smoke","current_version":"v0.1.0","phase":"draft","status":"draft","updated_at":"2026-08-18T00:00:00Z"}`)
	writeBoardTextFixture(t, root, ".ai-flow/work-items/"+workID+".json",
		`{"schema_version":1,"id":"`+workID+`","revision":1,"kind":"feature","title":"smoke","status":"active","priority":"P1","evidence_ids":[],"created_at":"2026-08-18T00:00:00Z","updated_at":"2026-08-18T00:00:00Z"}`)
	if withRun {
		if err := os.MkdirAll(filepath.Join(root, ".ai-flow", "runs", runID), 0o755); err != nil {
			t.Fatal(err)
		}
		writeBoardTextFixture(t, root, ".ai-flow/runs/"+runID+"/run.json",
			`{"schema_version":1,"id":"`+runID+`","revision":1,"work_item_id":"`+workID+`","owner":"codex","status":"active","phase":"implementing","evidence_ids":[],"started_at":"2026-08-18T00:00:00Z","updated_at":"2026-08-18T00:00:00Z"}`)
	}
	return root, workID, runID
}

func runEvidenceRecordForTest(t *testing.T, root string, args []string) string {
	t.Helper()
	full := append([]string{"--root", root}, args...)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStdout := os.Stdout
	os.Stdout = w
	runErr := runEvidenceRecord(full)
	w.Close()
	os.Stdout = origStdout
	if runErr != nil {
		t.Fatalf("evidence record returned error: %v", runErr)
	}
	buf := make([]byte, 128)
	n, _ := r.Read(buf)
	return strings.TrimSpace(string(buf[:n]))
}

func runEvidenceRecordCapture(t *testing.T, root string, args []string) error {
	t.Helper()
	full := append([]string{"--root", root}, args...)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	origStdout := os.Stdout
	os.Stdout = w
	runErr := runEvidenceRecord(full)
	w.Close()
	os.Stdout = origStdout
	_, _ = r.Read(make([]byte, 256))
	return runErr
}

func readEvidenceForTest(t *testing.T, root, id string) Evidence {
	t.Helper()
	ev, err := readEvidence(root, id)
	if err != nil {
		t.Fatalf("readEvidence(%s): %v", id, err)
	}
	return ev
}

func TestEvidenceRecordAgentClaimWithoutRunSucceeds(t *testing.T) {
	root, workID, _ := bootstrapEvidenceProject(t, false)
	id := runEvidenceRecordForTest(t, root, []string{
		"--work", workID, "--test", "TEST-20260818-eeeeeeee",
		"--source", "agent-claim", "--description", "no harness needed",
	})
	if !strings.HasPrefix(id, "EV-") {
		t.Fatalf("expected a new evidence id, got %q", id)
	}
	ev := readEvidenceForTest(t, root, id)
	if ev.RunID != nil {
		t.Fatalf("agent-claim evidence must not carry a run id, got %v", *ev.RunID)
	}
	if ev.Source != "agent-claim" {
		t.Fatalf("source mismatch: %s", ev.Source)
	}
}

func TestEvidenceRecordExternalModeWithoutRunSucceeds(t *testing.T) {
	root, workID, _ := bootstrapEvidenceProject(t, false)
	id := runEvidenceRecordForTest(t, root, []string{
		"--work", workID, "--test", "TEST-20260818-eeeeeeee",
		"--source", "external", "--mode", "external",
		"--description", "external ci run summary",
	})
	ev := readEvidenceForTest(t, root, id)
	if ev.RunID != nil {
		t.Fatalf("mode=external evidence must not carry a run id, got %v", *ev.RunID)
	}
}

func TestEvidenceRecordRequiresRunForRunModeExternalSource(t *testing.T) {
	root, workID, _ := bootstrapEvidenceProject(t, true)
	err := runEvidenceRecordCapture(t, root, []string{
		"--work", workID, "--test", "TEST-20260818-eeeeeeee",
		"--source", "external", "--description", "needs run",
	})
	if err == nil {
		t.Fatal("expected error when --run is missing for source=external + mode=run")
	}
	if !strings.Contains(err.Error(), "--run is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvidenceRecordRejectsRunWhenStandaloneIsImplied(t *testing.T) {
	root, workID, runID := bootstrapEvidenceProject(t, true)
	cases := [][]string{
		{"--work", workID, "--run", runID, "--test", "TEST-20260818-eeeeeeee", "--source", "agent-claim", "--description", "x"},
		{"--work", workID, "--run", runID, "--test", "TEST-20260818-eeeeeeee", "--source", "external", "--mode", "external", "--description", "x"},
	}
	for i, args := range cases {
		err := runEvidenceRecordCapture(t, root, args)
		if err == nil {
			t.Fatalf("case %d: expected error when --run is set on a standalone combination", i)
		}
		if !strings.Contains(err.Error(), "--run is ignored") {
			t.Fatalf("case %d: unexpected error: %v", i, err)
		}
	}
}

func TestEvidenceRecordRejectsSourceLocal(t *testing.T) {
	root, workID, runID := bootstrapEvidenceProject(t, true)
	err := runEvidenceRecordCapture(t, root, []string{
		"--work", workID, "--run", runID, "--test", "TEST-20260818-eeeeeeee",
		"--source", "local", "--description", "x",
	})
	if err == nil {
		t.Fatal("expected error rejecting --source=local on the record subcommand")
	}
	if !strings.Contains(err.Error(), "reserved for `evidence run`") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvidenceRecordBackwardCompatibleWithRun(t *testing.T) {
	root, workID, runID := bootstrapEvidenceProject(t, true)
	id := runEvidenceRecordForTest(t, root, []string{
		"--work", workID, "--run", runID, "--test", "TEST-20260818-eeeeeeee",
		"--source", "external", "--description", "tied to run",
	})
	ev := readEvidenceForTest(t, root, id)
	if ev.RunID == nil || *ev.RunID != runID {
		t.Fatalf("expected run id %s, got %v", runID, ev.RunID)
	}
}

func TestTraceabilityValidationAllowsStandaloneEvidence(t *testing.T) {
	root, workID, _ := bootstrapEvidenceProject(t, false)
	id := runEvidenceRecordForTest(t, root, []string{
		"--work", workID, "--test", "TEST-20260818-eeeeeeee",
		"--source", "agent-claim", "--description", "standalone claim",
	})
	if id == "" {
		t.Fatal("agent-claim record returned no id")
	}
	if _, err := readEvidence(root, id); err != nil {
		t.Fatalf("standalone evidence file should exist: %v", err)
	}
}

func TestTraceabilityValidationRejectsLocalEvidenceWithoutRun(t *testing.T) {
	root, workID, _ := bootstrapEvidenceProject(t, false)
	// Hand-write a local evidence record that points at no run. The validator
	// must reject this combination because local execution must be tied to a
	// harness run.
	const evidID = "EV-20260818-bbbbbbbb"
	ev := Evidence{
		SchemaVersion: 1, ID: evidID, WorkItemID: workID, RunID: nil,
		TestID: "TEST-20260818-eeeeeeee", Source: "local", Trust: "verified-local",
		Result: "passed", Command: []string{"go", "test", "./..."}, ExitCode: 0,
		GitSHA: "abcdef1234567",
		StartedAt: "2026-08-18T00:00:00Z", EndedAt: "2026-08-18T00:00:01Z",
		LogPath: ".ai-flow/evidence/logs/" + evidID + ".log",
		LogSHA256: strings.Repeat("a", 64),
		CreatedAt: "2026-08-18T00:00:01Z",
	}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", "evidence", evidID+".json"), ev); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".ai-flow", "evidence", "logs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".ai-flow", "evidence", "logs", evidID+".log"), []byte("log\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	issues := validateTraceability(root)
	matched := false
	for _, issue := range issues {
		if issue.Schema == "semantic-links" && strings.Contains(issue.Message, "must be linked to a development run") {
			matched = true
			break
		}
	}
	if !matched {
		t.Fatalf("expected validator to reject local evidence without a run; got issues: %+v", issues)
	}
}
