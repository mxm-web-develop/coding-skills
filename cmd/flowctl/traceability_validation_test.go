package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	traceGoalID        = "GOAL-20260817-11111111"
	traceRequirementID = "REQ-20260817-22222222"
	tracePlanID        = "PLAN-20260817-33333333"
	traceWorkID        = "WI-20260817-44444444"
	traceDecisionID    = "ADR-20260817-55555555"
	traceTestID        = "TEST-20260817-66666666"
	traceRunID         = "RUN-20260817-77777777"
	traceCheckpointID  = "CP-20260817-88888888"
	traceEvidenceID    = "EV-20260817-99999999"
	traceReleaseID     = "REL-20260817-aaaaaaaa"
	traceTime          = "2026-08-17T12:00:00Z"
)

func TestCompleteDeliveryTraceabilityGraph(t *testing.T) {
	root := createTraceabilityFixture(t)
	compiled, err := compileSchemas(filepath.Join("..", "..", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	targets, err := collectValidationTargets(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range targets {
		values, readErr := readValidationValues(target)
		if readErr != nil {
			t.Fatal(readErr)
		}
		for _, value := range values {
			if validateErr := compiled[target.Schema].Validate(value); validateErr != nil {
				t.Fatalf("%s does not match %s: %v", target.Path, target.Schema, validateErr)
			}
		}
	}
	if issues := validateTraceability(root); len(issues) != 0 {
		t.Fatalf("complete delivery graph produced issues: %#v", issues)
	}
}

func TestTraceabilityFindsBrokenBacklinksAndReleaseCoverage(t *testing.T) {
	root := createTraceabilityFixture(t)
	requirementPath := filepath.Join(root, ".ai-flow", "requirements", traceRequirementID+".json")
	var requirement map[string]any
	readFixtureJSON(t, requirementPath, &requirement)
	requirement["test_ids"] = []string{}
	if err := writeJSONAtomic(requirementPath, requirement); err != nil {
		t.Fatal(err)
	}

	evidencePath := filepath.Join(root, ".ai-flow", "evidence", traceEvidenceID+".json")
	var evidence map[string]any
	readFixtureJSON(t, evidencePath, &evidence)
	evidence["result"] = "failed"
	if err := writeJSONAtomic(evidencePath, evidence); err != nil {
		t.Fatal(err)
	}

	issues := validateTraceability(root)
	assertTraceIssue(t, issues, "linked requirement does not point back to test")
	assertTraceIssue(t, issues, "release task has a requirement without a linked test")
}

func TestTraceabilityRequiresTrustedPassingReleaseVerification(t *testing.T) {
	root := createTraceabilityFixture(t)
	evidencePath := filepath.Join(root, ".ai-flow", "evidence", traceEvidenceID+".json")
	var evidence map[string]any
	readFixtureJSON(t, evidencePath, &evidence)
	evidence["result"] = "failed"
	if err := writeJSONAtomic(evidencePath, evidence); err != nil {
		t.Fatal(err)
	}
	assertTraceIssue(t, validateTraceability(root), "release requirement lacks trusted passing verification")
}

func TestTraceabilityRejectsVerificationLogTraversal(t *testing.T) {
	root := createTraceabilityFixture(t)
	evidencePath := filepath.Join(root, ".ai-flow", "evidence", traceEvidenceID+".json")
	var evidence map[string]any
	readFixtureJSON(t, evidencePath, &evidence)
	evidence["log_path"] = ".ai-flow/evidence/logs/../../project.yaml"
	if err := writeJSONAtomic(evidencePath, evidence); err != nil {
		t.Fatal(err)
	}
	assertTraceIssue(t, validateTraceability(root), "path traversal is not allowed")
}

func TestTraceabilityRejectsNonCanonicalObjectPath(t *testing.T) {
	root := createTraceabilityFixture(t)
	original := filepath.Join(root, ".ai-flow", "requirements", traceRequirementID+".json")
	misplaced := filepath.Join(root, ".ai-flow", "requirements", "REQ-20260817-bbbbbbbb.json")
	if err := os.Rename(original, misplaced); err != nil {
		t.Fatal(err)
	}
	assertTraceIssue(t, validateTraceability(root), "object is not stored at its canonical path")
}

func createTraceabilityFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, directory := range []string{"state", "goals", "requirements", "plans", "work-items", "decisions", "tests", "evidence", "releases", "runs/" + traceRunID + "/checkpoints"} {
		if err := os.MkdirAll(filepath.Join(root, ".ai-flow", filepath.FromSlash(directory)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, ".ai-flow", "state", "current.yaml"), []byte("active_goal: "+traceGoalID+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeTraceFixture(t, root, "goals", traceGoalID, map[string]any{
		"schema_version": 1, "id": traceGoalID, "revision": 1, "status": "accepted", "title": "卖家报告",
		"problem": "报告生成会阻塞请求", "outcome": "用户可以异步取得报告", "actors": []string{"卖家"},
		"in_scope": []string{"异步报告"}, "non_goals": []string{"实时流处理"}, "acceptance_criteria": []string{"报告最终可下载"},
		"created_at": traceTime, "updated_at": traceTime,
	})
	writeTraceFixture(t, root, "requirements", traceRequirementID, map[string]any{
		"schema_version": 1, "id": traceRequirementID, "revision": 1, "goal_id": traceGoalID, "status": "accepted",
		"statement": "报告在后台可靠生成", "acceptance_criteria": []string{"失败可以重试"}, "test_ids": []string{traceTestID},
		"created_at": traceTime, "updated_at": traceTime,
	})
	writeTraceFixture(t, root, "plans", tracePlanID, map[string]any{
		"schema_version": 1, "id": tracePlanID, "revision": 1, "goal_id": traceGoalID, "status": "accepted", "title": "报告交付安排",
		"milestones":    []map[string]any{{"id": "MS-report", "title": "可靠生成", "outcome": "可以恢复失败任务", "requirement_ids": []string{traceRequirementID}, "exit_gates": []string{"测试通过"}}},
		"work_item_ids": []string{traceWorkID}, "created_at": traceTime, "updated_at": traceTime,
	})
	goalID := traceGoalID
	runID := traceRunID
	writeTraceFixture(t, root, "work-items", traceWorkID, WorkItem{
		SchemaVersion: 1, ID: traceWorkID, Revision: 3, Kind: "feature", Title: "实现异步报告", Status: "done", Priority: "high",
		GoalID: &goalID, RequirementIDs: []string{traceRequirementID}, AcceptanceCriteria: []string{"报告最终可下载"}, Scope: []string{"src/reports/**"},
		RunID: &runID, EvidenceIDs: []string{traceEvidenceID}, CreatedAt: traceTime, UpdatedAt: traceTime,
	})
	writeTraceFixture(t, root, "decisions", traceDecisionID, map[string]any{
		"schema_version": 1, "id": traceDecisionID, "revision": 1, "status": "accepted", "title": "后台任务实现",
		"context": "当前团队只维护数据库", "goal_id": traceGoalID, "requirement_ids": []string{traceRequirementID}, "work_item_ids": []string{traceWorkID},
		"options":  []map[string]any{{"name": "数据库任务表", "tradeoffs": []string{"需要控制并发"}}},
		"decision": "采用数据库任务表", "consequences": []string{"无需新增服务"}, "created_at": traceTime, "updated_at": traceTime,
	})
	writeTraceFixture(t, root, "tests", traceTestID, map[string]any{
		"schema_version": 1, "id": traceTestID, "revision": 1, "requirement_ids": []string{traceRequirementID}, "work_item_id": traceWorkID,
		"level": "integration", "purpose": "验证失败重试", "command": []string{"go", "test", "./..."}, "expected_result": "任务恢复并完成",
		"status": "green", "evidence_ids": []string{traceEvidenceID}, "created_at": traceTime, "updated_at": traceTime,
	})
	run := HarnessRun{
		SchemaVersion: 1, ID: traceRunID, Revision: 3, WorkItemID: traceWorkID, Owner: "codex-agent", Status: "completed", Phase: "reviewing",
		GitSHA: "abcdef1234567", CheckpointIDs: []string{traceCheckpointID}, EvidenceIDs: []string{traceEvidenceID},
		StartedAt: traceTime, UpdatedAt: traceTime, CompletedAt: pointer(traceTime),
	}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", "runs", traceRunID, "run.json"), run); err != nil {
		t.Fatal(err)
	}
	checkpoint := Checkpoint{
		SchemaVersion: 1, ID: traceCheckpointID, RunID: traceRunID, WorkItemID: traceWorkID, Sequence: 1, Phase: "implementing",
		Summary: "完成任务领取", NextAction: "运行测试", GitSHA: "abcdef1234567", CompletedSteps: []string{"实现任务领取"},
		ChangedFiles: []string{}, OpenQuestions: []string{}, CreatedAt: traceTime,
	}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", "runs", traceRunID, "checkpoints", traceCheckpointID+".json"), checkpoint); err != nil {
		t.Fatal(err)
	}
	logRelative := ".ai-flow/evidence/logs/" + traceEvidenceID + ".log"
	logAbsolute := filepath.Join(root, filepath.FromSlash(logRelative))
	if err := os.MkdirAll(filepath.Dir(logAbsolute), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logAbsolute, []byte("passed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeTraceFixture(t, root, "evidence", traceEvidenceID, Evidence{
		SchemaVersion: 1, ID: traceEvidenceID, WorkItemID: traceWorkID, RunID: traceRunID, TestID: traceTestID,
		Source: "local", Trust: "verified-local", Result: "passed", Command: []string{"go", "test", "./..."}, ExitCode: 0,
		GitSHA: "abcdef1234567", Environment: map[string]string{"os": "test"}, StartedAt: traceTime, EndedAt: traceTime,
		LogPath: logRelative, LogSHA256: strings.Repeat("a", 64), CreatedAt: traceTime,
	})
	writeTraceFixture(t, root, "releases", traceReleaseID, map[string]any{
		"schema_version": 1, "id": traceReleaseID, "revision": 1, "status": "ready", "previous_version": "v0.2.0", "version": "v0.3.0",
		"work_item_ids": []string{traceWorkID}, "commit_shas": []string{"abcdef1234567"}, "evidence_ids": []string{traceEvidenceID},
		"summary": "增加异步报告", "known_issues": []string{}, "rollback": "关闭异步报告入口", "created_at": traceTime, "updated_at": traceTime,
	})
	return root
}

func writeTraceFixture(t *testing.T, root, directory, id string, value any) {
	t.Helper()
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", directory, id+".json"), value); err != nil {
		t.Fatal(err)
	}
}

func readFixtureJSON(t *testing.T, path string, value any) {
	t.Helper()
	if err := readSemanticJSON(path, value); err != nil {
		t.Fatal(err)
	}
}

func assertTraceIssue(t *testing.T, issues []validationIssue, expected string) {
	t.Helper()
	for _, issue := range issues {
		if strings.Contains(issue.Message, expected) {
			return
		}
	}
	t.Fatalf("traceability issues do not contain %q: %#v", expected, issues)
}
