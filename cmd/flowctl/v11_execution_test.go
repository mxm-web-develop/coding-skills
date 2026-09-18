package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executionFixture(t *testing.T) (string, WorkItem, HarnessRun) {
	t.Helper()
	root, w, r := v1Fixture(t)
	writeBoardTextFixture(t, root, "go.mod", "module example\n\ngo 1.22\n")
	writeBoardTextFixture(t, root, "src/export.go", "package export\n")
	writeBoardTextFixture(t, root, "requirements.md", "CSV export must escape commas and quotes; include all selected orders.\n")
	if err := scanUpgradeEngineering(root); err != nil {
		t.Fatal(err)
	}
	if err := runProfileConfirmation([]string{"--root", root, "--work", w.ID, "--by", "planner", "--source", "go.mod"}); err != nil {
		t.Fatal(err)
	}
	spec := ExecutionSpec{Outcome: "Download selected orders", Interfaces: []string{"GET /orders/export -> text/csv"}, Invariants: []string{"authorization remains enforced"}, Steps: []string{"add escaping cases then implement export"}, EdgeCases: []string{"empty selection yields header only"}, Tests: []ExecutionTest{{ID: "export-test", Command: []string{"go", "version"}, Expected: "test harness command exits zero"}}, EntryPoints: []string{"src/export.go"}, Sources: []string{"requirements.md"}, Facts: []string{}, StopConditions: []string{"stop on authorization ambiguity"}, Capabilities: []string{"Go changes and command execution"}}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/reports", w.ID, "execution-input.json"), spec); err != nil {
		t.Fatal(err)
	}
	args := []string{"prepare", "--root", root, "--work", w.ID, "--file", filepath.Join(root, ".ai-flow/reports", w.ID, "execution-input.json"), "--by", "planner", "--executor", "test", "--reviewer", "reviewer", "--model", "unmeasured-model"}
	if err := runHandoff(args); err != nil {
		t.Fatal(err)
	}
	w, err := readWorkItem(root, w.ID)
	if err != nil {
		t.Fatal(err)
	}
	return root, w, r
}
func TestExecutionContextContainsCompleteInputsAndNeverTruncates(t *testing.T) {
	root, w, _ := executionFixture(t)
	b, err := executionContext(root, w.ID, 65536)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "escape commas") || !strings.Contains(string(b), "authorization remains enforced") {
		t.Fatal("missing execution input")
	}
	if _, err = executionContext(root, w.ID, 50); err == nil {
		t.Fatal("silently truncated contract")
	}
	writeBoardTextFixture(t, root, "src/export.go", "package export // implementation changed\n")
	if _, err = executionContext(root, w.ID, 65536); err != nil {
		t.Fatal("legitimate implementation invalidated the plan", err)
	}
	writeBoardTextFixture(t, root, "requirements.md", "new behavior")
	if _, err = executionContext(root, w.ID, 65536); err == nil {
		t.Fatal("stale requirements accepted")
	}
}
func TestEscalationSurvivesReopenAndOnlyPlannerResolves(t *testing.T) {
	root, w, r := executionFixture(t)
	args := []string{"escalate", "--root", root, "--work", w.ID, "--by", "test", "--reason", "selection API has no authorization context", "--evidence", "src/export.go"}
	if err := runHandoff(args); err != nil {
		t.Fatal(err)
	}
	w, _ = readWorkItem(root, w.ID)
	if err := checkRunBudget(root, w, r); err == nil {
		t.Fatal("continued unresolved conflict")
	}
	if err := runHandoff([]string{"resolve", "--root", root, "--work", w.ID, "--by", "test", "--reason", "ignore it"}); err == nil {
		t.Fatal("executor redesigned task")
	}
	if err := runHandoff([]string{"resolve", "--root", root, "--work", w.ID, "--by", "planner", "--reason", "existing authenticated wrapper provides it; inspect before continuing"}); err != nil {
		t.Fatal(err)
	}
	w, _ = readWorkItem(root, w.ID)
	if err := checkExecution(root, w, true); err != nil {
		t.Fatal(err)
	}
	if w.Execution.Escalations[0].Resolution == "" {
		t.Fatal("decision lost")
	}
}
func TestContractPreventsSelfReviewAndTestSubstitution(t *testing.T) {
	root, w, r := executionFixture(t)
	if err := runEvidenceCommand([]string{"--root", root, "--work", w.ID, "--run", r.ID, "--test", "export-test", "--quiet", "--", "go", "env", "GOOS"}); err == nil {
		t.Fatal("substituted weaker test")
	}
	if err := runEvidenceCommand([]string{"--root", root, "--work", w.ID, "--run", r.ID, "--test", "export-test", "--quiet", "--", "go", "version"}); err != nil {
		t.Fatal(err)
	}
	if err := runWorkReview([]string{"--root", root, "--id", w.ID, "--reviewer", "test", "--decision", "approved", "--summary", "self approved"}); err == nil {
		t.Fatal("self review accepted")
	}
	if err := runWorkReview([]string{"--root", root, "--id", w.ID, "--reviewer", "reviewer", "--decision", "approved", "--summary", "independent verification"}); err != nil {
		t.Fatal(err)
	}
}
func TestProfileScanIsCandidateAndFixturesAreNotProduction(t *testing.T) {
	root, w, _ := executionFixture(t)
	confirmed, _ := os.ReadFile(filepath.Join(root, ".ai-flow/baseline/engineering-profile.json"))
	writeBoardTextFixture(t, root, "tests/fixtures/web/package.json", `{"scripts":{"test":"broken"}}`)
	writeBoardTextFixture(t, root, "apps/service/package.json", `{"scripts":{"test":"test runner"}}`)
	if err := scanUpgradeEngineering(root); err != nil {
		t.Fatal(err)
	}
	current, _ := os.ReadFile(filepath.Join(root, ".ai-flow/baseline/engineering-profile.json"))
	if string(current) != string(confirmed) {
		t.Fatal("scan overwrote confirmed knowledge")
	}
	candidate, _ := os.ReadFile(filepath.Join(root, ".ai-flow/baseline/engineering-candidate.json"))
	if strings.Contains(string(candidate), "tests/fixtures") || !strings.Contains(string(candidate), "apps/service/package.json") {
		t.Fatal("component classification incorrect")
	}
	if err := checkExecution(root, w, true); err != nil {
		t.Fatal("candidate scan changed confirmed execution", err)
	}
	writeBoardTextFixture(t, root, "go.mod", "module changed\n")
	if err := checkExecution(root, w, true); err == nil {
		t.Fatal("changed manifest accepted")
	}
}
func TestFactsNeedEvidenceAndExplicitSupersession(t *testing.T) {
	root, w, _ := executionFixture(t)
	args := []string{"record", "--root", root, "--work", w.ID, "--key", "exports", "--claim", "CSV quotes escaped", "--by", "planner", "--source", "requirements.md"}
	if err := runMemory(args); err != nil {
		t.Fatal(err)
	}
	fact, err := verifiedFact(root, "exports")
	if err != nil {
		t.Fatal(err)
	}
	if err := runMemory(args); err == nil {
		t.Fatal("overwrote fact without review")
	}
	writeBoardTextFixture(t, root, "requirements.md", "new requirement")
	if _, err := verifiedFact(root, "exports"); err == nil {
		t.Fatal("stale fact usable")
	}
	digest, _ := hashJSON(fact)
	if err := runMemory(append(args, "--supersedes", digest)); err != nil {
		t.Fatal(err)
	}
	store, _ := loadFacts(root)
	if len(store.Facts) != 2 {
		t.Fatal("lost historical fact")
	}
}
func TestNewPendingDecisionAndScopeChangesInvalidateHandoff(t *testing.T) {
	root, w, _ := executionFixture(t)
	w.Scope = append(w.Scope, "api")
	if err := checkExecution(root, w, true); err == nil {
		t.Fatal("silent scope expansion")
	}
	w, _ = readWorkItem(root, w.ID)
	d := map[string]any{"id": "ADR-20260918-11111111", "status": "proposed", "work_item_ids": []string{w.ID}}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/decisions/ADR-20260918-11111111.json"), d); err != nil {
		t.Fatal(err)
	}
	if err := checkExecution(root, w, true); err == nil {
		t.Fatal("pending decision ignored")
	}
}
func TestContractSchemaAndSourceConfinement(t *testing.T) {
	root, w, _ := executionFixture(t)
	schemas, err := compileSchemas(filepath.Join(root, "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(w)
	var value any
	_ = json.Unmarshal(b, &value)
	if err := schemas["work-item.schema.json"].Validate(value); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(outside, []byte("outside"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip(err)
	}
	if _, err := readProjectSource(root, "link"); err == nil {
		t.Fatal("followed source outside project")
	}
}

func TestContractRevisionPreservesModelAndPriorPlan(t *testing.T) {
	root, w, _ := executionFixture(t)
	args := []string{"prepare", "--root", root, "--work", w.ID, "--file", filepath.Join(root, ".ai-flow/reports", w.ID, "execution-input.json"), "--by", "planner", "--executor", "test", "--reviewer", "reviewer", "--reason", "clarify implementation guidance"}
	if err := runHandoff(args); err != nil {
		t.Fatal(err)
	}
	current, _ := readWorkItem(root, w.ID)
	if current.Execution.ExecutorModel != "unmeasured-model" || current.Execution.Revision != 2 {
		t.Fatal("lost model identity or revision")
	}
	var previous ExecutionContract
	if err := readJSON(filepath.Join(root, ".ai-flow/reports", w.ID, "execution-history/revision-1.json"), &previous); err != nil {
		t.Fatal(err)
	}
	if previous.Revision != 1 || previous.Spec.Outcome != w.Execution.Spec.Outcome {
		t.Fatal("lost prior plan")
	}
}
