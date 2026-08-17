package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderBoardCreatesReadableVersionTaskDecisionAndTestViews(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{"state", "goals", "requirements", "plans", "work-items", "decisions", "tests", "evidence", "releases", "baseline"} {
		if err := os.MkdirAll(filepath.Join(root, ".ai-flow", directory), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeBoardTextFixture(t, root, ".ai-flow/manifest.yaml", "schema_version: 1\n")
	writeBoardTextFixture(t, root, ".ai-flow/project.yaml", "name: Seller Platform\nmode: existing\nprofile: core\ncurrent_version: v1.0.0\n")
	writeBoardTextFixture(t, root, ".ai-flow/state/current.yaml", "revision: 8\nphase: verifying\nstatus: active\ncurrent_version: v1.1.0\nactive_goal: GOAL-20260817-a1b2c3d4\nnext_action: review_visual_evidence\ntests: partial_pass\nupdated_at: 2026-08-17T13:30:00Z\n")

	writeBoardJSONFixture(t, root, ".ai-flow/goals/GOAL-20260817-a1b2c3d4.json", map[string]any{
		"id": "GOAL-20260817-a1b2c3d4", "status": "accepted", "title": "卖家画像与风险评分",
		"problem": "平台缺少卖家风险判断", "outcome": "运营可以查看可解释评分", "target_release": "v1.1.0",
		"acceptance_criteria": []string{"评分可解释", "页面可查看"}, "non_goals": []string{"自动封禁"}, "risks": []string{"历史数据不足"},
	})
	writeBoardJSONFixture(t, root, ".ai-flow/goals/GOAL-20260701-11111111.json", map[string]any{
		"id": "GOAL-20260701-11111111", "status": "archived", "title": "旧版基础能力", "target_release": "v0.9.0",
	})
	writeBoardJSONFixture(t, root, ".ai-flow/requirements/REQ-20260817-b2c3d4e5.json", map[string]any{
		"id": "REQ-20260817-b2c3d4e5", "goal_id": "GOAL-20260817-a1b2c3d4", "status": "accepted",
		"statement": "计算卖家风险评分", "acceptance_criteria": []string{"评分范围为0到100"}, "test_ids": []string{"TEST-20260817-e5f6a7b8"},
	})
	writeBoardJSONFixture(t, root, ".ai-flow/plans/PLAN-20260817-c3d4e5f6.json", map[string]any{
		"id": "PLAN-20260817-c3d4e5f6", "goal_id": "GOAL-20260817-a1b2c3d4", "status": "active", "title": "V1.1交付计划",
		"work_item_ids": []string{"WI-20260817-d4e5f6a7"},
		"milestones":    []map[string]any{{"id": "MS-score", "title": "评分内核", "outcome": "完成纯函数评分", "requirement_ids": []string{"REQ-20260817-b2c3d4e5"}, "exit_gates": []string{"单元和视觉测试通过"}, "target_release": "v1.1.0"}},
	})

	goalID := "GOAL-20260817-a1b2c3d4"
	owner := "codex-agent"
	work := WorkItem{SchemaVersion: 1, ID: "WI-20260817-d4e5f6a7", Revision: 2, Kind: "feature", Title: "实现卖家风险评分页面", Status: "in_progress", Priority: "high", GoalID: &goalID, RequirementIDs: []string{"REQ-20260817-b2c3d4e5"}, AcceptanceCriteria: []string{"运营可查看评分"}, Scope: []string{"src/seller-risk/**"}, Owner: &owner, EvidenceIDs: []string{"EV-20260817-b8c9d0e1"}, CreatedAt: "2026-08-17T10:30:00Z", UpdatedAt: "2026-08-17T12:30:00Z"}
	writeBoardJSONFixture(t, root, ".ai-flow/work-items/"+work.ID+".json", work)
	oldGoalID := "GOAL-20260701-11111111"
	oldWork := WorkItem{SchemaVersion: 1, ID: "WI-20260701-22222222", Revision: 1, Kind: "feature", Title: "旧版已经完成的任务", Status: "done", Priority: "low", GoalID: &oldGoalID, RequirementIDs: []string{}, AcceptanceCriteria: []string{"旧版验收"}, Scope: []string{"legacy/**"}, EvidenceIDs: []string{}, CreatedAt: "2026-07-01T10:00:00Z", UpdatedAt: "2026-07-01T11:00:00Z"}
	writeBoardJSONFixture(t, root, ".ai-flow/work-items/"+oldWork.ID+".json", oldWork)
	writeBoardJSONFixture(t, root, ".ai-flow/decisions/ADR-20260817-a7b8c9d0.json", map[string]any{
		"id": "ADR-20260817-a7b8c9d0", "status": "accepted", "title": "评分内核架构",
		"decision": "采用纯函数评分内核，I/O放在适配器", "recommended_option": "纯函数评分内核",
		"recommendation_reason": "最符合当前团队的测试和维护要求",
		"confirmation":          map[string]any{"status": "confirmed", "selected_option": "纯函数评分内核", "feedback": "先保持规则透明"},
		"options": []map[string]any{
			{"name": "纯函数评分内核", "summary": "强调可解释评分", "tradeoffs": []string{"需要显式适配I/O"}},
			{"name": "数据对比页面", "summary": "强调横向比较", "tradeoffs": []string{"信息密度较高"}, "prototype_path": ".ai-flow/prototypes/seller-risk/comparison/index.html", "prototype_focus": "快速比较多个卖家的风险差异"},
		},
		"consequences": []string{"易测试", "依赖显式"},
	})
	writeBoardJSONFixture(t, root, ".ai-flow/tests/TEST-20260817-e5f6a7b8.json", map[string]any{
		"id": "TEST-20260817-e5f6a7b8", "work_item_id": work.ID, "requirement_ids": []string{"REQ-20260817-b2c3d4e5"}, "level": "visual", "purpose": "确认桌面和移动端布局", "status": "green", "evidence_ids": []string{"EV-20260817-b8c9d0e1"},
	})
	evidence := Evidence{SchemaVersion: 1, ID: "EV-20260817-b8c9d0e1", WorkItemID: work.ID, RunID: "RUN-20260817-f6a7b8c9", TestID: "TEST-20260817-e5f6a7b8", Source: "local", Trust: "verified-local", Result: "passed", Command: []string{"pnpm", "playwright", "test"}, ExitCode: 0, GitSHA: "abcdef1234567", StartedAt: "2026-08-17T13:00:00Z", EndedAt: "2026-08-17T13:01:00Z", LogPath: ".ai-flow/evidence/logs/test.log", LogSHA256: strings.Repeat("a", 64), CreatedAt: "2026-08-17T13:01:00Z"}
	writeBoardJSONFixture(t, root, ".ai-flow/evidence/"+evidence.ID+".json", evidence)
	writeBoardJSONFixture(t, root, ".ai-flow/releases/REL-20260817-d0e1f2a3.json", map[string]any{
		"id": "REL-20260817-d0e1f2a3", "status": "released", "previous_version": "v1.0.0", "version": "v1.1.0", "work_item_ids": []string{work.ID}, "evidence_ids": []string{evidence.ID}, "summary": "新增卖家风险评分", "known_issues": []string{"暂不支持跨店铺分析"}, "migration": "无需迁移", "rollback": "关闭 seller-risk 特性开关", "updated_at": "2026-08-17T14:00:00Z",
	})
	writeBoardJSONFixture(t, root, ".ai-flow/baseline/engineering-profile.json", map[string]any{
		"languages": []map[string]any{{"name": "TypeScript", "version": "5"}}, "frameworks": []map[string]any{{"name": "Next.js", "version": "15"}},
		"architecture": map[string]any{"style": "feature modules", "constraints": []string{"server/client边界显式"}}, "selected_playbooks": []string{"typescript-web", "web-and-visual"},
		"visual_testing": map[string]any{"required": true, "tool": "Playwright", "browsers": []string{"chromium"}, "viewports": []string{"1440x900", "390x844"}}, "unknowns": []string{},
	})

	if err := renderBoard(root); err != nil {
		t.Fatal(err)
	}
	assertBoardContains(t, root, "STATUS.md", "项目当前处于 **V1** 大版本", "版本进度", "实现卖家风险评分页面", "界面效果", "确认桌面和移动端布局", "通过 1 / 失败 0")
	assertBoardContains(t, root, "ROADMAP.md", "卖家画像与风险评分", "版本与开发阶段", "评分内核", "单元和视觉测试通过")
	assertBoardContains(t, root, "CURRENT_STATE.md", "计算卖家风险评分", "采用纯函数评分内核", "最符合当前团队的测试和维护要求", "已确认", "界面体验方案", "打开体验方案", "快速比较多个卖家的风险差异", "TypeScript 5", "按产品功能划分模块", "开发与测试规范", "Playwright")
	assertBoardContainsRaw(t, root, "CURRENT_STATE.md", "../../.ai-flow/prototypes/seller-risk/comparison/index.html")
	assertBoardContains(t, root, "RELEASES.md", "v1.1.0", "新增卖家风险评分", "暂不支持跨店铺分析", "关闭 seller-risk 特性开关")
	assertBoardNotContains(t, root, "STATUS.md", oldWork.Title, "1 个已完成", ".ai-flow", "Evidence", "Work Item", "TEST-", "WI-", "Playbook")
	assertBoardNotContains(t, root, "ROADMAP.md", ".ai-flow", "Milestone", "Work Item", "完成门禁", "PLAN-")
	assertBoardNotContains(t, root, "CURRENT_STATE.md", "Playbook", "feature modules", "REQ-", "Goal", "Requirement")
	assertBoardNotContains(t, root, "RELEASES.md", ".ai-flow", "Evidence", "REL-")
	assertBoardContainsRaw(t, root, "STATUS.md", work.ID, "PLAN-20260817-c3d4e5f6", "TEST-20260817-e5f6a7b8", evidence.ID)
	assertBoardContainsRaw(t, root, "STATUS.md", "ai-flow-trace:version=v1.1.0", "goal=GOAL-20260817-a1b2c3d4", "plan=PLAN-20260817-c3d4e5f6")
	assertBoardContainsRaw(t, root, "ROADMAP.md", "GOAL-20260817-a1b2c3d4", "PLAN-20260817-c3d4e5f6", "MS-score")
	assertBoardContainsRaw(t, root, "CURRENT_STATE.md", "REQ-20260817-b2c3d4e5", "ADR-20260817-a7b8c9d0")
	assertBoardContainsRaw(t, root, "RELEASES.md", "REL-20260817-d0e1f2a3", work.ID, evidence.ID)

	statusPath := filepath.Join(root, "docs", "board", "STATUS.md")
	statusData, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statusPath, append(statusData, []byte("\nmanual status\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	issues := validateBoardFreshness(root)
	if len(issues) != 1 || !strings.Contains(issues[0].Message, "stale or manually edited") {
		t.Fatalf("manual board edit was not detected: %#v", issues)
	}
}

func TestCompareVersionsUsesNumericSemverOrder(t *testing.T) {
	if compareVersions("v1.10.0", "v1.2.0") <= 0 {
		t.Fatal("v1.10.0 should sort after v1.2.0")
	}
}

func writeBoardTextFixture(t *testing.T, root, relativePath, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeBoardJSONFixture(t *testing.T, root, relativePath string, value any) {
	t.Helper()
	if err := writeJSONAtomic(filepath.Join(root, filepath.FromSlash(relativePath)), value); err != nil {
		t.Fatal(err)
	}
}

func assertBoardContains(t *testing.T, root, name string, values ...string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "docs", "board", name))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, value := range values {
		if !strings.Contains(content, value) {
			t.Fatalf("%s does not contain %q:\n%s", name, value, content)
		}
	}
}

func assertBoardNotContains(t *testing.T, root, name string, values ...string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "docs", "board", name))
	if err != nil {
		t.Fatal(err)
	}
	content := stripHTMLComments(string(data))
	for _, value := range values {
		if strings.Contains(content, value) {
			t.Fatalf("%s unexpectedly contains %q:\n%s", name, value, content)
		}
	}
}

func assertBoardContainsRaw(t *testing.T, root, name string, values ...string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "docs", "board", name))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, value := range values {
		if !strings.Contains(content, value) {
			t.Fatalf("%s raw trace does not contain %q:\n%s", name, value, content)
		}
	}
}

func stripHTMLComments(content string) string {
	for {
		start := strings.Index(content, "<!--")
		if start < 0 {
			return content
		}
		endOffset := strings.Index(content[start+4:], "-->")
		if endOffset < 0 {
			return content[:start]
		}
		end := start + 4 + endOffset + 3
		content = content[:start] + content[end:]
	}
}
