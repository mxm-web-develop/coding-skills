package main

import (
	"path/filepath"
	"testing"
)

func TestCompileAllSchemas(t *testing.T) {
	schemaRoot := filepath.Join("..", "..", "schemas")
	compiled, err := compileSchemas(schemaRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(compiled) != 14 {
		t.Fatalf("compiled %d schemas, want 14 object schemas plus common definitions", len(compiled))
	}
	for _, required := range []string{
		"work-item.schema.json",
		"run.schema.json",
		"checkpoint.schema.json",
		"evidence.schema.json",
		"event.schema.json",
		"engineering-profile.schema.json",
		"workspace-document-inventory.schema.json",
	} {
		if compiled[required] == nil {
			t.Fatalf("missing compiled schema %s", required)
		}
	}
}

func TestWorkspaceDocumentInventoryFixture(t *testing.T) {
	compiled, err := compileSchemas(filepath.Join("..", "..", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	values, err := readValidationValues(validationTarget{Path: filepath.Join("..", "..", "tests", "fixtures", "workspace-document-inventory.json")})
	if err != nil {
		t.Fatal(err)
	}
	if err := compiled["workspace-document-inventory.schema.json"].Validate(values[0]); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceDocumentInventoryRejectsUnsafePlans(t *testing.T) {
	compiled, err := compileSchemas(filepath.Join("..", "..", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join("..", "..", "tests", "fixtures", "workspace-document-inventory.json")
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "audit only cannot move",
			mutate: func(inventory map[string]any) {
				inventory["mode"] = "audit-only"
			},
		},
		{
			name: "protected document cannot archive",
			mutate: func(inventory map[string]any) {
				document := inventory["documents"].([]any)[1].(map[string]any)
				document["action"] = "archive"
				document["target_path"] = ".ai-flow/archive/legacy-documents/v1.2.0/README.md"
				document["approval"] = "approved"
			},
		},
		{
			name: "source path cannot escape root",
			mutate: func(inventory map[string]any) {
				document := inventory["documents"].([]any)[0].(map[string]any)
				document["source_path"] = "../outside.md"
			},
		},
		{
			name: "moved document requires approval",
			mutate: func(inventory map[string]any) {
				document := inventory["documents"].([]any)[0].(map[string]any)
				document["approval"] = "rejected"
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			values, err := readValidationValues(validationTarget{Path: fixture})
			if err != nil {
				t.Fatal(err)
			}
			inventory := values[0].(map[string]any)
			test.mutate(inventory)
			if err := compiled["workspace-document-inventory.schema.json"].Validate(inventory); err == nil {
				t.Fatal("unsafe inventory unexpectedly validated")
			}
		})
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
