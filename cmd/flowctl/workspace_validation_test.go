package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestWorkspaceStructureAndCleanupFixtures(t *testing.T) {
	compiled, err := compileSchemas(filepath.Join("..", "..", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	for fixture, schema := range map[string]string{
		"workspace-structure-inventory.json": "workspace-structure-inventory.schema.json",
		"workspace-cleanup-plan.json":        "workspace-cleanup-plan.schema.json",
	} {
		values, err := readValidationValues(validationTarget{Path: filepath.Join("..", "..", "tests", "fixtures", fixture)})
		if err != nil {
			t.Fatal(err)
		}
		if err := compiled[schema].Validate(values[0]); err != nil {
			t.Fatalf("%s: %v", fixture, err)
		}
	}
}

func TestWorkspaceCleanupPlanRejectsUnsafeActions(t *testing.T) {
	compiled, err := compileSchemas(filepath.Join("..", "..", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join("..", "..", "tests", "fixtures", "workspace-cleanup-plan.json")
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "shared code cannot archive",
			mutate: func(plan map[string]any) {
				item := plan["items"].([]any)[2].(map[string]any)
				item["action"] = "archive-code"
				item["target_path"] = ".ai-flow/archive/legacy-code/v1.2.0/shared"
				item["approval"] = "approved"
			},
		},
		{
			name: "code archive cannot target files archive",
			mutate: func(plan map[string]any) {
				item := plan["items"].([]any)[0].(map[string]any)
				item["target_path"] = ".ai-flow/archive/legacy-files/v1.2.0/services/api-old"
			},
		},
		{
			name: "removed output requires approval",
			mutate: func(plan map[string]any) {
				item := plan["items"].([]any)[1].(map[string]any)
				item["result"] = "removed"
			},
		},
		{
			name: "windows drive path cannot leave root",
			mutate: func(plan map[string]any) {
				item := plan["items"].([]any)[0].(map[string]any)
				item["source_path"] = `C:\outside`
			},
		},
		{
			name: "ignore rule must target gitignore",
			mutate: func(plan map[string]any) {
				item := plan["items"].([]any)[3].(map[string]any)
				item["target_path"] = "apps/web/package.json"
			},
		},
		{
			name: "approved plan requires every path approval",
			mutate: func(plan map[string]any) {
				plan["status"] = "approved"
				approval := plan["approval"].(map[string]any)
				approval["status"] = "approved"
				approval["approved_by"] = "project owner"
				approval["approved_at"] = "2026-08-17T13:00:00Z"
				approval["mapping_sha256"] = strings.Repeat("f", 64)
			},
		},
		{
			name: "mutation requires recovery reference",
			mutate: func(plan map[string]any) {
				item := plan["items"].([]any)[0].(map[string]any)
				recovery := item["recovery"].(map[string]any)
				recovery["reference"] = nil
			},
		},
		{
			name: "verified batch requires before and after commands",
			mutate: func(plan map[string]any) {
				batch := plan["batches"].([]any)[0].(map[string]any)
				batch["status"] = "verified"
				batch["before_commands"] = []any{}
				batch["after_commands"] = []any{}
				batch["before_evidence_ids"] = []any{"EV-20260817-11111111"}
				batch["after_evidence_ids"] = []any{"EV-20260817-22222222"}
				batch["mutation_started_at"] = "2026-08-17T13:00:00Z"
				batch["mutation_completed_at"] = "2026-08-17T13:01:00Z"
			},
		},
		{
			name: "restored batch requires failure and recovery proof",
			mutate: func(plan map[string]any) {
				batch := plan["batches"].([]any)[0].(map[string]any)
				batch["status"] = "restored"
				batch["before_commands"] = []any{}
				batch["after_commands"] = []any{}
				batch["recovery_evidence_ids"] = []any{"EV-20260817-33333333"}
				batch["mutation_started_at"] = "2026-08-17T13:00:00Z"
				batch["mutation_completed_at"] = "2026-08-17T13:01:00Z"
			},
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			values, err := readValidationValues(validationTarget{Path: fixture})
			if err != nil {
				t.Fatal(err)
			}
			plan := values[0].(map[string]any)
			test.mutate(plan)
			if err := compiled["workspace-cleanup-plan.schema.json"].Validate(plan); err == nil {
				t.Fatal("unsafe cleanup plan unexpectedly validated")
			}
		})
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
			name: "windows source path cannot escape root",
			mutate: func(inventory map[string]any) {
				document := inventory["documents"].([]any)[0].(map[string]any)
				document["source_path"] = `D:\outside.md`
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

func TestWorkspaceSemanticValidationRejectsBrokenGraphsAndPlans(t *testing.T) {
	t.Run("valid workspace plan", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		if issues := validateSemanticLinks(root); len(issues) != 0 {
			t.Fatalf("valid workspace plan has semantic issues: %+v", issues)
		}
	})

	t.Run("dangling component dependency", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		path := filepath.Join(root, ".ai-flow", "baseline", "workspace-structure-inventory.json")
		var inventory map[string]any
		if err := readJSON(path, &inventory); err != nil {
			t.Fatal(err)
		}
		component := inventory["components"].([]any)[0].(map[string]any)
		component["internal_dependencies"] = []any{"CMP-missing"}
		if err := writeJSONAtomic(path, inventory); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("dangling component dependency unexpectedly passed semantic validation")
		}
	})

	t.Run("batch references missing item", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		path := filepath.Join(root, ".ai-flow", "workspace-cleanup", "PLAN-example.json")
		var plan map[string]any
		if err := readJSON(path, &plan); err != nil {
			t.Fatal(err)
		}
		batch := plan["batches"].([]any)[0].(map[string]any)
		batch["item_ids"] = []any{"CLN-I999"}
		if err := writeJSONAtomic(path, plan); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("dangling cleanup item unexpectedly passed semantic validation")
		}
	})

	t.Run("archive mapping loses original path", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		path := filepath.Join(root, ".ai-flow", "workspace-cleanup", "PLAN-example.json")
		var plan map[string]any
		if err := readJSON(path, &plan); err != nil {
			t.Fatal(err)
		}
		item := plan["items"].([]any)[0].(map[string]any)
		item["target_path"] = ".ai-flow/archive/legacy-code/v1.2.0/renamed-api"
		if err := writeJSONAtomic(path, plan); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("lossy archive mapping unexpectedly passed semantic validation")
		}
	})

	t.Run("approved content drifts from digest", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		path := filepath.Join(root, ".ai-flow", "workspace-cleanup", "PLAN-example.json")
		var planMap map[string]any
		if err := readJSON(path, &planMap); err != nil {
			t.Fatal(err)
		}
		planMap["status"] = "approved"
		for _, rawItem := range planMap["items"].([]any) {
			item := rawItem.(map[string]any)
			if item["action"] != "keep" {
				item["approval"] = "approved"
			}
		}
		approval := planMap["approval"].(map[string]any)
		approval["status"] = "approved"
		approval["approved_by"] = "project owner"
		approval["approved_at"] = "2026-08-17T13:00:00Z"
		if err := writeJSONAtomic(path, planMap); err != nil {
			t.Fatal(err)
		}
		var plan workspaceCleanupSemantic
		if err := readSemanticJSON(path, &plan); err != nil {
			t.Fatal(err)
		}
		digest, err := cleanupApprovalDigest(plan)
		if err != nil {
			t.Fatal(err)
		}
		approval["mapping_sha256"] = digest
		if err := writeJSONAtomic(path, planMap); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) != 0 {
			t.Fatalf("approved unchanged plan has semantic issues: %+v", issues)
		}
		item := planMap["items"].([]any)[0].(map[string]any)
		recovery := item["recovery"].(map[string]any)
		recovery["reference"] = "changed after approval"
		if err := writeJSONAtomic(path, planMap); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("approved cleanup drift unexpectedly passed semantic validation")
		}
	})

	t.Run("source content drifts from fingerprint", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		if err := os.WriteFile(filepath.Join(root, "services", "api-old", "unexpected.go"), []byte("package old\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("changed cleanup source unexpectedly passed semantic validation")
		}
	})

	t.Run("ignore target drifts from approved snapshot", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		if err := os.WriteFile(filepath.Join(root, "apps", "web", ".gitignore"), []byte("other-output/\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("changed ignore target unexpectedly passed semantic validation")
		}
	})

	t.Run("component boundary drifts without a new commit", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("{\"workspaces\":[]}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("changed workspace manifest unexpectedly passed semantic validation")
		}
	})

	t.Run("cleanup plan requires structure inventory", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		if err := os.Remove(filepath.Join(root, ".ai-flow", "baseline", "workspace-structure-inventory.json")); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("cleanup plan without structure inventory unexpectedly passed semantic validation")
		}
	})

	t.Run("cleanup cannot cross protected or nested boundaries", func(t *testing.T) {
		root := workspaceValidationFixtureRoot(t)
		path := filepath.Join(root, ".ai-flow", "baseline", "workspace-structure-inventory.json")
		var inventory map[string]any
		if err := readJSON(path, &inventory); err != nil {
			t.Fatal(err)
		}
		inventory["protected_paths"] = append(inventory["protected_paths"].([]any), "services/api-old")
		repository := inventory["repository"].(map[string]any)
		repository["nested_repositories"] = []any{"apps/web"}
		if err := writeJSONAtomic(path, inventory); err != nil {
			t.Fatal(err)
		}
		if issues := validateSemanticLinks(root); len(issues) == 0 {
			t.Fatal("protected and nested cleanup paths unexpectedly passed semantic validation")
		}
	})
}

func TestWorkspacePathResolutionRejectsSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "outside-link")); err != nil {
		t.Fatal(err)
	}
	if err := ensurePathInsideRepository(root, "outside-link/file.txt"); err == nil {
		t.Fatal("symlink escape unexpectedly accepted")
	}
}

func TestCleanupApprovalDigestChangesWithApprovedContent(t *testing.T) {
	root := workspaceValidationFixtureRoot(t)
	path := filepath.Join(root, ".ai-flow", "workspace-cleanup", "PLAN-example.json")
	var plan workspaceCleanupSemantic
	if err := readSemanticJSON(path, &plan); err != nil {
		t.Fatal(err)
	}
	first, err := cleanupApprovalDigest(plan)
	if err != nil {
		t.Fatal(err)
	}
	reference := "different recovery reference"
	plan.Items[0].Recovery.Reference = &reference
	second, err := cleanupApprovalDigest(plan)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("approval digest did not change with approved recovery content")
	}
}

func TestCleanupCompletionEvidenceMustMatchTaskResultAndRevision(t *testing.T) {
	root := t.TempDir()
	logDir := filepath.Join(root, ".ai-flow", "evidence", "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"before.log", "after.log"} {
		if err := os.WriteFile(filepath.Join(logDir, name), []byte("passed\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	logDigest, err := sha256File(filepath.Join(logDir, "before.log"))
	if err != nil {
		t.Fatal(err)
	}
	beforeEvidenceID := "EV-20260817-22222222"
	afterEvidenceID := "EV-20260817-44444444"
	completedGitSHA := "not-a-git-repository"
	beforeEvidence := Evidence{
		SchemaVersion: 1, ID: beforeEvidenceID, WorkItemID: "WI-20260817-11111111", RunID: "RUN-20260817-33333333",
		TestID: "workspace-cleanup", Source: "local", Trust: "verified-local", Result: "passed", Command: []string{"verify"},
		ExitCode: 0, GitSHA: completedGitSHA, StartedAt: "2026-08-17T11:58:00Z", EndedAt: "2026-08-17T11:59:00Z",
		LogPath: ".ai-flow/evidence/logs/before.log", LogSHA256: logDigest, CreatedAt: "2026-08-17T11:59:00Z",
	}
	afterEvidence := beforeEvidence
	afterEvidence.ID = afterEvidenceID
	afterEvidence.StartedAt = "2026-08-17T12:02:00Z"
	afterEvidence.EndedAt = "2026-08-17T12:03:00Z"
	afterEvidence.CreatedAt = "2026-08-17T12:03:00Z"
	afterEvidence.LogPath = ".ai-flow/evidence/logs/after.log"
	if err := writeJSONAtomic(evidencePath(root, beforeEvidenceID), beforeEvidence); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(evidencePath(root, afterEvidenceID), afterEvidence); err != nil {
		t.Fatal(err)
	}
	mutationStarted := "2026-08-17T12:00:00Z"
	mutationCompleted := "2026-08-17T12:01:00Z"
	plan := workspaceCleanupSemantic{Status: "applied", GitSHA: completedGitSHA, WorkItemID: beforeEvidence.WorkItemID, CompletedGitSHA: &completedGitSHA, VerificationEvidenceIDs: []string{afterEvidenceID}}
	plan.Batches = []cleanupBatchSemantic{{
		ID: "CLN-B001", BeforeCommands: []string{"verify"}, BeforeEvidenceIDs: []string{beforeEvidenceID},
		AfterCommands: []string{"verify"}, AfterEvidenceIDs: []string{afterEvidenceID},
		MutationStartedAt: &mutationStarted, MutationCompletedAt: &mutationCompleted, Status: "verified",
	}}
	issues := []string{}
	validateCleanupCompletionEvidence(root, plan, func(message string) { issues = append(issues, message) })
	if len(issues) != 0 {
		t.Fatalf("valid completion evidence rejected: %v", issues)
	}
	plan.Batches[0].AfterEvidenceIDs = []string{beforeEvidenceID}
	plan.VerificationEvidenceIDs = []string{beforeEvidenceID}
	issues = nil
	validateCleanupCompletionEvidence(root, plan, func(message string) { issues = append(issues, message) })
	if len(issues) == 0 {
		t.Fatal("reused before/after evidence unexpectedly accepted")
	}
	plan.VerificationEvidenceIDs = []string{afterEvidenceID}
	plan.Batches[0].AfterEvidenceIDs = nil
	issues = nil
	validateCleanupCompletionEvidence(root, plan, func(message string) { issues = append(issues, message) })
	if len(issues) == 0 {
		t.Fatal("verified batch without command evidence unexpectedly accepted")
	}
	plan.Batches[0].AfterEvidenceIDs = []string{afterEvidenceID}
	afterEvidence.Result = "failed"
	if err := writeJSONAtomic(evidencePath(root, afterEvidenceID), afterEvidence); err != nil {
		t.Fatal(err)
	}
	issues = nil
	validateCleanupCompletionEvidence(root, plan, func(message string) { issues = append(issues, message) })
	if len(issues) == 0 {
		t.Fatal("failed completion evidence unexpectedly accepted")
	}
}

func TestRestoredBatchRequiresFailureBeforeRecovery(t *testing.T) {
	root := t.TempDir()
	beforeID := "EV-20260817-11111111"
	failureID := "EV-20260817-22222222"
	recoveryID := "EV-20260817-33333333"
	for _, evidence := range []Evidence{
		{ID: beforeID, StartedAt: "2026-08-17T11:58:00Z", EndedAt: "2026-08-17T11:59:00Z"},
		{ID: failureID, StartedAt: "2026-08-17T12:02:00Z", EndedAt: "2026-08-17T12:04:00Z"},
		{ID: recoveryID, StartedAt: "2026-08-17T12:03:00Z", EndedAt: "2026-08-17T12:05:00Z"},
	} {
		if err := writeJSONAtomic(evidencePath(root, evidence.ID), evidence); err != nil {
			t.Fatal(err)
		}
	}
	mutationStarted := "2026-08-17T12:00:00Z"
	mutationCompleted := "2026-08-17T12:01:00Z"
	batch := cleanupBatchSemantic{
		ID: "CLN-B001", Status: "restored", BeforeEvidenceIDs: []string{beforeID},
		AfterEvidenceIDs: []string{failureID}, RecoveryEvidenceIDs: []string{recoveryID},
		MutationStartedAt: &mutationStarted, MutationCompletedAt: &mutationCompleted,
	}
	issues := []string{}
	validateCleanupBatchTiming(root, batch, func(message string) { issues = append(issues, message) })
	if len(issues) == 0 {
		t.Fatal("recovery that starts before failure completes unexpectedly accepted")
	}
	var recovery Evidence
	if err := readJSON(evidencePath(root, recoveryID), &recovery); err != nil {
		t.Fatal(err)
	}
	recovery.StartedAt = "2026-08-17T12:05:00Z"
	recovery.EndedAt = "2026-08-17T12:06:00Z"
	if err := writeJSONAtomic(evidencePath(root, recoveryID), recovery); err != nil {
		t.Fatal(err)
	}
	issues = nil
	validateCleanupBatchTiming(root, batch, func(message string) { issues = append(issues, message) })
	if len(issues) != 0 {
		t.Fatalf("ordered failure and recovery evidence rejected: %v", issues)
	}
}

func workspaceValidationFixtureRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	structureValues, err := readValidationValues(validationTarget{Path: filepath.Join("..", "..", "tests", "fixtures", "workspace-structure-inventory.json")})
	if err != nil {
		t.Fatal(err)
	}
	planValues, err := readValidationValues(validationTarget{Path: filepath.Join("..", "..", "tests", "fixtures", "workspace-cleanup-plan.json")})
	if err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", "baseline", "workspace-structure-inventory.json"), structureValues[0]); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", "workspace-cleanup", "PLAN-example.json"), planValues[0]); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"services/api-old", "apps/web/dist", "apps/web/coverage", "shared"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(path)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{"package.json", "go.work", "apps/web/package.json", ".github/workflows/web.yml", "services/api/go.mod", "deploy/api.yaml"} {
		absolutePath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(absolutePath, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", "work-items", "WI-20260817-11111111.json"), map[string]any{}); err != nil {
		t.Fatal(err)
	}
	return root
}
