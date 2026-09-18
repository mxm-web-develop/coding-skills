package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentPlansIgnoreReplacedAndUnreleasedRelease(t *testing.T) {
	data := boardData{Plans: []boardPlan{
		{ID: "old", Status: "superseded", Milestones: []boardMilestone{{Title: "OLD", TargetRelease: "v1.0.0"}}},
		{ID: "new", Status: "accepted", Milestones: []boardMilestone{{Title: "NEW", TargetRelease: "v1.0.0"}}},
	}, Releases: []boardRelease{{Version: "v1.0.0", Status: "planned"}}}
	files := expectedPlanFiles(data)
	if !strings.Contains(files["plans/v1.0.0.md"], "NEW") || strings.Contains(renderPlanIndex(data), "| 已发布 |") {
		t.Fatal("current plan was replaced by history or an unshipped release")
	}
}

func TestBoardShowsOnlyLatestTrustedTestAttempt(t *testing.T) {
	data := boardData{Evidence: []Evidence{
		{ID: "a", WorkItemID: "w", TestID: "test", Result: "failed", Trust: "verified-local", CreatedAt: "2026-01-01T00:00:00Z"},
		{ID: "b", WorkItemID: "w", TestID: "test", Result: "passed", Trust: "verified-local", CreatedAt: "2026-01-02T00:00:00Z"},
	}}
	p, f, _ := evidenceCountsForWork(data, "w")
	if p != 1 || f != 0 {
		t.Fatalf("latest pass: %d/%d", p, f)
	}
	data.Evidence[1].Trust = "unverified"
	p, _, other := evidenceCountsForWork(data, "w")
	if p != 0 || other != 1 {
		t.Fatal("unverified result presented as a pass")
	}
}

func TestBoardWriteRejectsInternalState(t *testing.T) {
	if err := writeBoardFile(filepath.Join(t.TempDir(), "board.md"), "任务 WI-20260918-abcdef12 状态: in_progress"); err == nil {
		t.Fatal("internal wording leaked")
	}
}

func TestGeneratedPlanReconciliationPreservesHistory(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "board")
	if err := os.MkdirAll(filepath.Join(dir, "plans"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := reconcileBoardFiles(root, map[string]string{"PLANS.md": "index", "plans/v1.0.0.md": "old plan"}); err != nil {
		t.Fatal(err)
	}
	if err := reconcileBoardFiles(root, map[string]string{"STATUS.md": "current"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "plans/v1.0.0.md")); !os.IsNotExist(err) {
		t.Fatal("stale plan left active")
	}
	files, _ := filepath.Glob(filepath.Join(root, ".ai-flow/archive/generated/*/plans/v1.0.0.md"))
	if len(files) != 1 {
		t.Fatal("old generated content was not preserved")
	}
}

func TestBoardBoundsLongRunningTaskList(t *testing.T) {
	data := boardData{Status: projectStatus{CurrentVersion: "v1.0.0"}}
	for i := 0; i < 60; i++ {
		data.WorkItems = append(data.WorkItems, WorkItem{ID: fmt.Sprint(i), Title: fmt.Sprintf("历史任务%d", i), Status: "done"})
	}
	for i := 0; i < 30; i++ {
		data.WorkItems = append(data.WorkItems, WorkItem{ID: fmt.Sprint(i + 60), Title: fmt.Sprintf("开发任务%d", i), Status: "in_progress"})
	}
	board := renderStatusBoard(data)
	if strings.Contains(board, "历史任务") || strings.Count(board, "<!-- ai-flow-trace:work=") != 12 {
		t.Fatal("history or unbounded active list leaked")
	}
}

func TestLatestEvidenceFractionalTimestampOrder(t *testing.T) {
	all := []Evidence{{ID: "a", WorkItemID: "w", TestID: "t", CreatedAt: "2026-01-01T00:00:00Z"}, {ID: "b", WorkItemID: "w", TestID: "t", CreatedAt: "2026-01-01T00:00:00.1Z"}}
	if got := latestEvidence(all); len(got) != 1 || got[0].ID != "b" {
		t.Fatal("fractional timestamp sorted lexically")
	}
}
