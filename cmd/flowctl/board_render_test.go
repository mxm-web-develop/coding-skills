package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintBoardFileDetectsForbiddenSectionRefs(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		mustHit    []string
		mustMiss   []string
	}{
		{
			name:    "bare §N",
			body:    "按需求文档 §42 字面布局建 agent-core",
			mustHit: []string{"§42"},
		},
		{
			name:    "Chinese section",
			body:    "见需求文档第 38 节描述",
			mustHit: []string{"第 38 节"},
		},
		{
			name:    "Phase and Module",
			body:    "Phase 2 完成；Module 4 进入开发",
			mustHit: []string{"Phase 2", "Module 4"},
		},
		{
			name:    "Step number",
			body:    "完成 Step 3 之后进入下一步",
			mustHit: []string{"Step 3"},
		},
		{
			name:  "normal prose is not flagged",
			body:  "今天我们完成了用户故事讨论、目录拆解、状态字段定型，并把测试补齐。",
		},
		{
			name:    "HTML comments are exempt",
			body:    "正常文档内容。\n\n<!-- §42 trace:goal=GOAL-1 -->\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			violations := lintBoardFile(c.body)
			for _, want := range c.mustHit {
				found := false
				for _, v := range violations {
					if v == want {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected lint to flag %q in %q, got %v", want, c.body, violations)
				}
			}
			if len(c.mustHit) == 0 && len(violations) > 0 {
				t.Fatalf("expected no violations for %q, got %v", c.body, violations)
			}
		})
	}
}

func TestRenderPlanDocForwardsInScopeSectionRefsSoTheyCanBeLinted(t *testing.T) {
	// This test documents the current behaviour: a goal whose in_scope list
	// contains a §N reference will currently be forwarded verbatim into the
	// rendered plan document, so the lint will catch it. The lint is the
	// safety net that motivates rewriting the source data instead of letting
	// §N reach the rendered output.
	root := t.TempDir()
	for _, d := range []string{
		"state", "goals", "requirements", "plans", "work-items",
		"decisions", "tests", "evidence", "releases", "runs",
	} {
		if err := mkdirAllForTest(filepathJoin(root, ".ai-flow", d)); err != nil {
			t.Fatal(err)
		}
	}
	writeBoardTextFixture(t, root, ".ai-flow/state/project.json",
		`{"schema_version":1,"project_name":"smoke","current_version":"v0.1.0","phase":"draft","status":"draft","updated_at":"2026-08-19T00:00:00Z"}`)
	const goalID = "GOAL-20260819-aaaa1111"
	const planID = "PLAN-20260819-bbbb2222"
	const workID = "WI-20260819-cccc3333"
	goal := map[string]any{
		"id": goalID, "status": "active", "title": "Agent Core 目录结构",
		"problem": "需要落 agent-core", "outcome": "目录可见",
		"target_release": "v0.5.0",
		"in_scope": []string{"按 §42 字面布局建 agent-core/", "对外汇报里不能再出现 §N"},
		"non_goals":      []string{"自动发布"},
		"acceptance_criteria": []string{"目录结构与 §42 一致"},
	}
	writeBoardJSONFixture(t, root, ".ai-flow/goals/"+goalID+".json", goal)
	writeBoardJSONFixture(t, root, ".ai-flow/plans/"+planID+".json", map[string]any{
		"id": planID, "goal_id": goalID, "status": "active", "title": "Agent Core 阶段安排",
		"milestones": []map[string]any{
			{"title": "Phase A", "outcome": "目录落地", "target_release": "v0.5.0", "exit_gates": []string{"目录符合 §42"}},
		},
	})
	data, err := loadBoardData(root, projectStatus{ProjectName: "smoke", CurrentVersion: "v0.1.0", Phase: "draft", Status: "draft", UpdatedAt: "2026-08-19T00:00:00Z"})
	if err != nil {
		t.Fatalf("loadBoardData: %v", err)
	}
	body := renderPlanDoc(data, boardPlan{
		ID: planID, GoalID: goalID, Title: "Agent Core 阶段安排",
		Milestones: []boardMilestone{{Title: "Phase A", Outcome: "目录落地", TargetRelease: "v0.5.0", ExitGates: []string{"目录符合 §42"}}},
	})
	if body == "" {
		t.Fatal("expected a non-empty plan document body")
	}
	violations := lintBoardFile(body)
	if len(violations) == 0 {
		t.Fatalf("expected lint to flag §N references forwarded from goal/plan, got none; body was:\n%s", body)
	}
	// The lint must specifically surface at least one §N occurrence.
	found := false
	for _, v := range violations {
		if strings.Contains(v, "§") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected a §N violation, got %v", violations)
	}
}

// filepathJoin exists so the test file does not need to import path/filepath
// directly when the helper only does one thing. Keeping it tiny reduces the
// risk of unused-import churn.
func filepathJoin(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for _, p := range parts[1:] {
		if len(p) == 0 {
			continue
		}
		if out != "" && !strings.HasSuffix(out, "/") && !strings.HasPrefix(p, "/") {
			out += "/"
		}
		out += p
	}
	return out
}

func mkdirAllForTest(path string) error {
	return os.MkdirAll(path, 0o755)
}

func TestWriteBoardFileRejectsForbiddenContent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "STATUS.md")
	body := "今天聊一聊需求文档 §42 这一节。"
	err := writeBoardFile(path, body)
	if err == nil {
		t.Fatalf("expected writeBoardFile to refuse content containing §42, got nil error")
	}
	if !strings.Contains(err.Error(), "§42") {
		t.Fatalf("expected error to mention §42, got: %v", err)
	}
	if _, statErr := os.Stat(path); statErr == nil {
		t.Fatalf("expected %s to NOT exist after refused write", path)
	}
}

func TestWriteBoardFileAcceptsCleanContent(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "STATUS.md")
	body := "今天聊一聊需求文档里关于目录结构那一段。"
	if err := writeBoardFile(path, body); err != nil {
		t.Fatalf("expected writeBoardFile to accept clean content, got %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if string(got) != body {
		t.Fatalf("body mismatch: got %q want %q", string(got), body)
	}
}
