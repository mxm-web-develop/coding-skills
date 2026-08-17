package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckpointResumeRequiresExplicitOwnershipHandoff(t *testing.T) {
	root := t.TempDir()
	workID := "WI-20260817-11223344"
	runID := "RUN-20260817-22334455"
	checkpointID := "CP-20260817-33445566"
	for _, directory := range []string{"work-items", "runs/" + runID + "/checkpoints", "locks", "events", "bin"} {
		if err := os.MkdirAll(filepath.Join(root, ".ai-flow", filepath.FromSlash(directory)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".ai-flow", "bin", executableName("flowctl")), []byte("test runtime"), 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	oldOwner := "agent:claude"
	work := WorkItem{
		SchemaVersion: 1, ID: workID, Revision: 2, Kind: "feature", Title: "Continue across IDEs", Status: "blocked", Priority: "medium",
		RequirementIDs: []string{}, AcceptanceCriteria: []string{"the same run resumes"}, Scope: []string{"src/**"}, Owner: &oldOwner, RunID: &runID,
		EvidenceIDs: []string{}, BlockedReason: pointer("waiting for editor switch"), CreatedAt: now, UpdatedAt: now,
	}
	run := HarnessRun{
		SchemaVersion: 1, ID: runID, Revision: 2, WorkItemID: workID, Owner: oldOwner, Status: "checkpointed", Phase: "solution_design",
		GitSHA: gitSHA(root), CheckpointIDs: []string{checkpointID}, EvidenceIDs: []string{}, StartedAt: now, UpdatedAt: now,
	}
	checkpoint := Checkpoint{
		SchemaVersion: 1, ID: checkpointID, RunID: runID, WorkItemID: workID, Sequence: 1, Phase: "solution_design",
		Summary: "Saved before editor switch", NextAction: "Continue the same task", GitSHA: gitSHA(root), CompletedSteps: []string{}, ChangedFiles: []string{}, OpenQuestions: []string{}, CreatedAt: now,
	}
	if err := writeJSONAtomic(workItemPath(root, workID), &work); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(runPath(root, runID), &run); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(checkpointPath(root, runID, checkpointID), &checkpoint); err != nil {
		t.Fatal(err)
	}

	err := runCheckpointResume([]string{"--root", root, "--run", runID, "--owner", "agent:cursor"})
	if err == nil || !strings.Contains(err.Error(), "explicit transfer") {
		t.Fatalf("owner changed without explicit handoff: %v", err)
	}
	if err := runCheckpointResume([]string{"--root", root, "--run", runID, "--owner", "agent:cursor", "--handoff-from", oldOwner}); err != nil {
		t.Fatal(err)
	}

	updatedRun, err := readRun(root, runID)
	if err != nil {
		t.Fatal(err)
	}
	updatedWork, err := readWorkItem(root, workID)
	if err != nil {
		t.Fatal(err)
	}
	var lease WorkLease
	if err := readJSON(leasePath(root, workID), &lease); err != nil {
		t.Fatal(err)
	}
	if updatedRun.Owner != "agent:cursor" || updatedRun.Status != "running" {
		t.Fatalf("run was not transferred in place: %#v", updatedRun)
	}
	if updatedWork.Owner == nil || *updatedWork.Owner != "agent:cursor" || updatedWork.RunID == nil || *updatedWork.RunID != runID {
		t.Fatalf("work no longer points at transferred run: %#v", updatedWork)
	}
	if lease.Owner != "agent:cursor" || lease.RunID != runID {
		t.Fatalf("lease was not transferred: %#v", lease)
	}
}
