package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompileAllSchemas(t *testing.T) {
	schemaRoot := filepath.Join("..", "..", "schemas")
	compiled, err := compileSchemas(schemaRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(compiled) != 16 {
		t.Fatalf("compiled %d schemas, want 16 object schemas plus common definitions", len(compiled))
	}
	for _, required := range []string{
		"work-item.schema.json",
		"run.schema.json",
		"checkpoint.schema.json",
		"evidence.schema.json",
		"event.schema.json",
		"engineering-profile.schema.json",
		"workspace-document-inventory.schema.json",
		"workspace-structure-inventory.schema.json",
		"workspace-cleanup-plan.schema.json",
	} {
		if compiled[required] == nil {
			t.Fatalf("missing compiled schema %s", required)
		}
	}
}

func TestObjectIDFormat(t *testing.T) {
	id, err := newObjectID("WI")
	if err != nil {
		t.Fatal(err)
	}
	if err := requireObjectID(id, "WI"); err != nil {
		t.Fatalf("generated invalid id %s: %v", id, err)
	}
	if err := requireObjectID(id, "EV"); err == nil {
		t.Fatal("accepted an ID with the wrong prefix")
	}
}

func TestEngineeringProfileSchema(t *testing.T) {
	compiled, err := compileSchemas(filepath.Join("..", "..", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	profile := map[string]any{
		"schema_version": 1,
		"revision":       1,
		"detected_at":    "2026-08-17T12:00:00Z",
		"git_sha":        "abcdef1234567",
		"languages": []any{
			map[string]any{"name": "TypeScript", "version": "5", "evidence": []any{"package.json", "tsconfig.json"}},
		},
		"frameworks": []any{
			map[string]any{"name": "Next.js", "category": "frontend", "evidence": []any{"package.json"}},
		},
		"package_managers": []any{"pnpm"},
		"build_systems":    []any{"Next.js"},
		"architecture": map[string]any{
			"style":            "feature modules",
			"module_roots":     []any{"src"},
			"generated_roots":  []any{},
			"public_api_roots": []any{"src/app"},
			"constraints":      []any{"server and client modules remain explicit"},
		},
		"commands": map[string]any{
			"install":     []any{"pnpm install --frozen-lockfile"},
			"build":       []any{"pnpm build"},
			"format":      []any{"pnpm format:check"},
			"lint":        []any{"pnpm lint"},
			"typecheck":   []any{"pnpm typecheck"},
			"unit":        []any{"pnpm test"},
			"integration": []any{},
			"e2e":         []any{"pnpm playwright test"},
			"visual":      []any{"pnpm playwright test visual"},
		},
		"selected_playbooks": []any{"typescript-web", "web-and-visual"},
		"community_skills": []any{
			map[string]any{"name": "vercel-react-best-practices", "source": "https://github.com/vercel-labs/agent-skills", "version": "1.0.0", "reason": "Detected React UI", "trust": "vendor"},
		},
		"visual_testing": map[string]any{"required": true, "tool": "Playwright", "browsers": []any{"chromium"}, "viewports": []any{"1440x900", "390x844"}},
		"unknowns":       []any{},
	}
	if err := compiled["engineering-profile.schema.json"].Validate(profile); err != nil {
		t.Fatal(err)
	}
}

func TestDecisionSchemaSupportsInteractiveTechnologyAndUXConfirmation(t *testing.T) {
	compiled, err := compileSchemas(filepath.Join("..", "..", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	decision := map[string]any{
		"schema_version":      1,
		"id":                  "ADR-20260817-a7b8c9d0",
		"revision":            1,
		"status":              "accepted",
		"title":               "运营页面体验方向",
		"context":             "需要在快速浏览和逐步引导之间选择",
		"decision_type":       "frontend-ux-ui",
		"evaluation_criteria": []any{"信息获取速度", "移动端可用性", "实现和维护成本"},
		"options": []any{
			map[string]any{
				"name": "快速浏览", "summary": "高信息密度看板", "strengths": []any{"比较速度快"},
				"weaknesses": []any{"新用户学习成本较高"}, "project_fit": "沿用现有表格组件",
				"risks": []any{"移动端拥挤"}, "adoption_impact": "无需新增依赖", "testing_impact": "增加手机视口测试",
				"rollback": "恢复当前页面", "tradeoffs": []any{"速度优先于引导"},
				"prototype_path": ".ai-flow/prototypes/seller-risk/fast-scan/index.html", "prototype_focus": "快速对比",
			},
			map[string]any{"name": "逐步引导", "tradeoffs": []any{"理解更容易但操作步骤更多"}, "prototype_path": ".ai-flow/prototypes/seller-risk/guided-flow/index.html"},
		},
		"recommended_option":    "快速浏览",
		"recommendation_reason": "现有用户每天需要比较大量卖家",
		"confirmation": map[string]any{
			"status": "confirmed", "selected_option": "快速浏览", "feedback": "保留移动端详情抽屉", "confirmed_at": "2026-08-17T12:00:00Z",
		},
		"decision":     "采用快速浏览方向，并保留移动端详情抽屉",
		"consequences": []any{"桌面端对比效率提高"},
		"rollback":     "恢复当前页面",
		"sources":      []any{"现有组件和用户反馈"},
		"created_at":   "2026-08-17T11:00:00Z",
		"updated_at":   "2026-08-17T12:00:00Z",
	}
	if err := compiled["decision.schema.json"].Validate(decision); err != nil {
		t.Fatal(err)
	}
}

func TestInteractiveDecisionRequiresConsistentChoiceAndSafePrototype(t *testing.T) {
	root := t.TempDir()
	decisionDirectory := filepath.Join(root, ".ai-flow", "decisions")
	prototypePath := filepath.Join(root, ".ai-flow", "prototypes", "seller-risk", "fast-scan", "index.html")
	secondPrototypePath := filepath.Join(root, ".ai-flow", "prototypes", "seller-risk", "guided-flow", "index.html")
	if err := os.MkdirAll(decisionDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(prototypePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(secondPrototypePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(prototypePath, []byte("<!doctype html><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>体验方案</title><button>查看</button><style>@media (max-width: 640px) { button { transition: all .2s ease; } }</style>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secondPrototypePath, []byte("<!doctype html><meta name=\"viewport\" content=\"width=device-width, initial-scale=1\"><title>体验方案</title><a href=\"#more\">更多</a><style>@media (prefers-reduced-motion: reduce) { * { animation: none; } }</style>"), 0o644); err != nil {
		t.Fatal(err)
	}
	decisionPath := filepath.Join(decisionDirectory, "ADR-20260817-a7b8c9d0.json")
	decision := map[string]any{
		"decision_type": "frontend-ux-ui", "status": "accepted",
		"options": []any{
			map[string]any{"name": "快速浏览", "prototype_path": ".ai-flow/prototypes/seller-risk/fast-scan/index.html"},
			map[string]any{"name": "逐步引导", "prototype_path": ".ai-flow/prototypes/seller-risk/guided-flow/index.html"},
		},
		"recommended_option": "快速浏览", "recommendation_reason": "符合高频比较场景",
		"confirmation": map[string]any{"status": "confirmed", "selected_option": "快速浏览"},
	}
	if err := writeJSONAtomic(decisionPath, decision); err != nil {
		t.Fatal(err)
	}
	if issues := validateSolutionDecisions(root); len(issues) != 0 {
		t.Fatalf("valid interactive decision produced issues: %#v", issues)
	}

	decision["recommended_option"] = "不存在的方向"
	decision["confirmation"] = map[string]any{"status": "pending", "selected_option": nil}
	if err := writeJSONAtomic(decisionPath, decision); err != nil {
		t.Fatal(err)
	}
	if issues := validateSolutionDecisions(root); len(issues) < 2 {
		t.Fatalf("inconsistent interactive decision produced too few issues: %#v", issues)
	}

	decision["recommended_option"] = "快速浏览"
	decision["confirmation"] = map[string]any{"status": "confirmed", "selected_option": "快速浏览"}
	decision["options"] = []any{
		map[string]any{"name": "快速浏览", "prototype_path": ".ai-flow/prototypes/seller-risk/../../secrets.html"},
		map[string]any{"name": "逐步引导", "prototype_path": ".ai-flow/prototypes/seller-risk/guided-flow/index.html"},
	}
	if err := writeJSONAtomic(decisionPath, decision); err != nil {
		t.Fatal(err)
	}
	if issues := validateSolutionDecisions(root); len(issues) == 0 {
		t.Fatal("prototype path traversal unexpectedly passed validation")
	}
}
