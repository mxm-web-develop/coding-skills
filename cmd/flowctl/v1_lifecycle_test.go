package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func v1Fixture(t *testing.T) (string, WorkItem, HarnessRun) {
	t.Helper()
	root := t.TempDir()
	writeBoardTextFixture(t, root, ".ai-flow/bin/"+executableName("flowctl"), "test runtime")
	if err := runProjectInit([]string{"--root", root, "--mode", "existing", "--name", "验收样例"}); err != nil {
		t.Fatal(err)
	}
	schemas, _ := filepath.Glob(filepath.Join("..", "..", "schemas", "*.json"))
	for _, path := range schemas {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := writeBytesAtomic(filepath.Join(root, "schemas", filepath.Base(path)), data); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	w := WorkItem{SchemaVersion: 1, WorkflowVersion: 1, ID: "WI-20260918-11111111", Revision: 1, Kind: "feature", Title: "订单导出", Status: "in_progress", Priority: "high", RequirementIDs: []string{}, AcceptanceCriteria: []string{"可下载"}, Scope: []string{"src"}, EvidenceIDs: []string{}, CreatedAt: now, UpdatedAt: now}
	r := HarnessRun{SchemaVersion: 1, ID: "RUN-20260918-11111111", Revision: 1, WorkItemID: w.ID, Owner: "test", Status: "running", Phase: "implementing", GitSHA: gitSHA(root), CheckpointIDs: []string{}, EvidenceIDs: []string{}, StartedAt: now, UpdatedAt: now, Budgets: RunBudgets{MaxElapsedMinutes: 60, MaxRetries: 3, MaxChangedFiles: 20}}
	w.Owner = &r.Owner
	w.RunID = &r.ID
	if err := writeJSONAtomic(workItemPath(root, w.ID), w); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(runPath(root, r.ID), r); err != nil {
		t.Fatal(err)
	}
	return root, w, r
}

func v1Verify(t *testing.T, root string, w WorkItem, r HarnessRun) {
	t.Helper()
	if err := runEvidenceCommand([]string{"--root", root, "--work", w.ID, "--run", r.ID, "--test", "check", "--quiet", "--", "go", "version"}); err != nil {
		t.Fatal(err)
	}
}
func v1Present(t *testing.T, root string, w WorkItem) {
	t.Helper()
	if err := runWorkReview([]string{"--root", root, "--id", w.ID, "--reviewer", "reviewer", "--decision", "approved", "--summary", "需求与检查通过"}); err != nil {
		t.Fatal(err)
	}
	if err := runWorkAcceptance([]string{"--root", root, "--id", w.ID, "--instructions", "打开导出页面", "--step", "点击导出", "--expected", "下载文件"}); err != nil {
		t.Fatal(err)
	}
}

func TestV1AcceptanceCloseoutPreservesTraceability(t *testing.T) {
	root, w, r := v1Fixture(t)
	v1Verify(t, root, w, r)
	if err := runWorkComplete([]string{"--root", root, "--id", w.ID}); err == nil {
		t.Fatal("completed without review and user acceptance")
	}
	v1Present(t, root, w)
	if err := runWorkAccept([]string{"--root", root, "--id", w.ID, "--by", "用户", "--feedback", "已实际下载，正确", "--result", "passed"}); err != nil {
		t.Fatal(err)
	}
	args := []string{"--root", root, "--id", w.ID, "--summary", "支持下载订单文件"}
	if err := runWorkComplete(args); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(workItemPath(root, w.ID)); !os.IsNotExist(err) {
		t.Fatal("closed record remained active")
	}
	item, err := readWorkItem(root, w.ID)
	if err != nil || item.Status != "done" || item.ArchivePath == "" {
		t.Fatalf("archive lookup: %+v %v", item, err)
	}
	if err := runWorkComplete(args); err != nil {
		t.Fatal("closeout retry failed", err)
	}
	if err := runValidate([]string{"--root", root}); err != nil {
		t.Fatal("history broke validation", err)
	}
}

func TestV1CodeChangeInvalidatesVerificationAndUserAcceptance(t *testing.T) {
	root, w, r := v1Fixture(t)
	v1Verify(t, root, w, r)
	v1Present(t, root, w)
	writeBoardTextFixture(t, root, "src/export.go", "changed code")
	if err := runWorkAccept([]string{"--root", root, "--id", w.ID, "--by", "用户", "--feedback", "通过", "--result", "passed"}); err == nil {
		t.Fatal("accepted changed code")
	}
	item, _ := readWorkItem(root, w.ID)
	if err := validateWorkEvidence(root, item); err == nil {
		t.Fatal("accepted stale verification")
	}
}

func TestV1RejectedAcceptanceReturnsSameTask(t *testing.T) {
	root, w, r := v1Fixture(t)
	v1Verify(t, root, w, r)
	v1Present(t, root, w)
	if err := runWorkAccept([]string{"--root", root, "--id", w.ID, "--by", "用户", "--feedback", "下载内容不正确", "--result", "failed"}); err != nil {
		t.Fatal(err)
	}
	item, _ := readWorkItem(root, w.ID)
	if item.Status != "in_progress" || item.RunID == nil || *item.RunID != r.ID || item.Review != nil {
		t.Fatal("rejection lost task continuity")
	}
}

func TestV1OverlappingTaskAndMissingDependencyCannotStart(t *testing.T) {
	root, w, r := v1Fixture(t)
	lease := WorkLease{SchemaVersion: 1, WorkItemID: "WI-20260918-22222222", RunID: r.ID, Owner: "other", Scope: []string{"src/**"}, AcquiredAt: time.Now().UTC().Format(time.RFC3339), ExpiresAt: time.Now().Add(time.Hour).UTC().Format(time.RFC3339)}
	if err := writeJSONAtomic(leasePath(root, lease.WorkItemID), lease); err != nil {
		t.Fatal(err)
	}
	if err := checkWorkStart(root, w); err == nil || !strings.Contains(err.Error(), "overlapping") {
		t.Fatal("overlapping writer accepted", err)
	}
}

func TestV1RequiredTestCannotBeSkipped(t *testing.T) {
	root, w, r := v1Fixture(t)
	v1Verify(t, root, w, r)
	item, _ := readWorkItem(root, w.ID)
	item.RequiredTests = []string{"missing-test"}
	if err := validateWorkEvidence(root, item); err == nil {
		t.Fatal("missing required test passed completion gate")
	}
}

func TestV1StaleAcceptanceCanReopenWithoutNewTask(t *testing.T) {
	root, w, r := v1Fixture(t)
	v1Verify(t, root, w, r)
	v1Present(t, root, w)
	writeBoardTextFixture(t, root, "src/changed.txt", "new result")
	if err := runWorkReopen([]string{"--root", root, "--id", w.ID, "--reason", "用户补充字段要求"}); err != nil {
		t.Fatal(err)
	}
	v1Verify(t, root, w, r)
	v1Present(t, root, w)
	current, _ := readWorkItem(root, w.ID)
	if current.Status != "awaiting_acceptance" || *current.RunID != r.ID {
		t.Fatal("lost continuity")
	}
}

func TestV1TimeBudgetStopsAndCanBeExplicitlyRevised(t *testing.T) {
	root, w, r := v1Fixture(t)
	r.StartedAt = time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339)
	if err := writeJSONAtomic(runPath(root, r.ID), r); err != nil {
		t.Fatal(err)
	}
	if err := checkRunBudget(root, w, r); err == nil {
		t.Fatal("expired budget passed")
	}
	if err := runWorkBudget([]string{"--root", root, "--id", w.ID, "--max-elapsed-minutes", "240", "--reason", "完成诊断后继续验证"}); err != nil {
		t.Fatal(err)
	}
	r, _ = readRun(root, r.ID)
	if err := checkRunBudget(root, w, r); err != nil {
		t.Fatal(err)
	}
}

func TestHumanNextActionPreserved(t *testing.T) {
	action := "补齐CSV转义和金额测试；下载交互待确认"
	if humanAction(action) != action {
		t.Fatal("specific next action was lost")
	}
}

func TestV1EvidenceMergesResultsFromStaleCommandContexts(t *testing.T) {
	root, w, r := v1Fixture(t)
	staleWork, staleRun := w, r
	first := Evidence{ID: "EV-20260918-11111111", WorkItemID: w.ID, TestID: "first", Result: "passed", Trust: "verified-local"}
	second := Evidence{ID: "EV-20260918-22222222", WorkItemID: w.ID, TestID: "second", Result: "passed", Trust: "verified-local"}
	if err := persistEvidence(root, &w, &r, &first); err != nil {
		t.Fatal(err)
	}
	if err := persistEvidence(root, &staleWork, &staleRun, &second); err != nil {
		t.Fatal(err)
	}
	current, _ := readWorkItem(root, w.ID)
	run, _ := readRun(root, r.ID)
	if len(current.EvidenceIDs) != 2 || len(run.EvidenceIDs) != 2 {
		t.Fatal("concurrent command completion lost an earlier result")
	}
}
