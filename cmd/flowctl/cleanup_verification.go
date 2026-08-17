package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func validateCleanupStructureBoundaries(itemID, path string, structure workspaceStructureSemantic, add func(string)) {
	for _, protectedPath := range structure.ProtectedPaths {
		if pathsOverlap(path, protectedPath) {
			add(fmt.Sprintf("cleanup item %s overlaps protected path %s", itemID, protectedPath))
		}
	}
	for _, sharedRoot := range structure.Repository.SharedRoots {
		if pathsOverlap(path, sharedRoot) {
			add(fmt.Sprintf("cleanup item %s overlaps shared root %s", itemID, sharedRoot))
		}
	}
	for _, nestedRepository := range structure.Repository.NestedRepositories {
		if pathsOverlap(path, nestedRepository) {
			add(fmt.Sprintf("cleanup item %s crosses nested repository %s; clean that repository separately", itemID, nestedRepository))
		}
	}
}

func validateCleanupInventoryAndBoundaries(root, inventoryPath string, plan workspaceCleanupSemantic, structure workspaceStructureSemantic, add func(string)) {
	inventoryDigest, err := canonicalJSONFileDigest(inventoryPath)
	if err != nil || inventoryDigest != plan.InventorySHA256 {
		add("workspace structure inventory changed after the cleanup plan was created")
	}
	if plan.Status != "proposed" && plan.Status != "approved" && plan.Status != "executing" {
		return
	}
	declared := map[string]string{}
	for _, fingerprint := range plan.BoundaryFingerprints {
		path := filepath.ToSlash(strings.ReplaceAll(fingerprint.Path, "\\", "/"))
		if declared[path] != "" {
			add("duplicate cleanup boundary fingerprint: " + path)
		}
		declared[path] = fingerprint.SHA256
		actual, err := sha256RepositoryPath(root, path)
		if err != nil || actual != fingerprint.SHA256 {
			add("cleanup boundary changed after the plan was created: " + path)
		}
	}
	required := map[string]bool{}
	for _, path := range structure.Repository.WorkspaceManifests {
		required[filepath.ToSlash(strings.ReplaceAll(path, "\\", "/"))] = true
	}
	scopeComponents := stringSet(plan.Scope.Components)
	for _, component := range structure.Components {
		if !scopeComponents[component.ID] {
			continue
		}
		for _, path := range component.Manifests {
			required[filepath.ToSlash(strings.ReplaceAll(path, "\\", "/"))] = true
		}
		for _, path := range component.DeploymentEvidence {
			if looksLikeRepositoryPath(path) {
				required[filepath.ToSlash(strings.ReplaceAll(path, "\\", "/"))] = true
			}
		}
	}
	for path := range required {
		if declared[path] == "" {
			add("cleanup plan is missing a component boundary fingerprint: " + path)
		}
	}
}

func looksLikeRepositoryPath(value string) bool {
	return value != "" && !strings.Contains(value, "://") && !strings.ContainsAny(value, "\t\r\n") && !strings.Contains(value, " ")
}

func validateCleanupApprovalBinding(plan workspaceCleanupSemantic, add func(string)) {
	if plan.Approval.Status != "approved" {
		return
	}
	expected, err := cleanupApprovalDigest(plan)
	if err != nil {
		add("cannot calculate cleanup approval digest: " + err.Error())
		return
	}
	if plan.Approval.MappingSHA256 == nil || *plan.Approval.MappingSHA256 != expected {
		add("approved cleanup content does not match its approval digest")
	}
}

func validateCleanupPathFingerprints(root string, plan workspaceCleanupSemantic, add func(string)) {
	if plan.Status == "rejected" {
		return
	}
	for _, item := range plan.Items {
		sourcePath := filepath.Join(root, filepath.FromSlash(strings.ReplaceAll(item.SourcePath, "\\", "/")))
		validateCleanupTargetFingerprint(root, item, add)
		switch item.Result {
		case "removed":
			if _, err := os.Lstat(sourcePath); !os.IsNotExist(err) {
				add("removed cleanup item still exists at its source: " + item.ID)
			}
			continue
		case "moved":
			if _, err := os.Lstat(sourcePath); !os.IsNotExist(err) {
				add("archived cleanup item still exists at its source: " + item.ID)
			}
			if item.TargetPath == nil {
				continue
			}
			validateCleanupFingerprint(root, item.ID, *item.TargetPath, item.SourceFingerprint, add)
			continue
		case "updated":
			validateCleanupFingerprint(root, item.ID, item.SourcePath, item.SourceFingerprint, add)
			if item.Action == "add-ignore-rule" {
				validateIgnoreRule(root, item, add)
			}
			continue
		default:
			validateCleanupFingerprint(root, item.ID, item.SourcePath, item.SourceFingerprint, add)
		}
	}
}

func validateCleanupTargetFingerprint(root string, item cleanupItemSemantic, add func(string)) {
	if item.TargetPath == nil {
		return
	}
	if item.Result == "moved" || item.Result == "updated" {
		if item.TargetFingerprintAfter == nil {
			add("cleanup item is missing its approved final target fingerprint: " + item.ID)
			return
		}
		validateCleanupFingerprint(root, item.ID, *item.TargetPath, *item.TargetFingerprintAfter, add)
		return
	}
	targetPath := filepath.Join(root, filepath.FromSlash(strings.ReplaceAll(*item.TargetPath, "\\", "/")))
	_, err := os.Lstat(targetPath)
	if !item.TargetExistedBefore {
		if !os.IsNotExist(err) {
			add("cleanup target changed after approval: " + item.ID)
		}
		return
	}
	if err != nil {
		add("cleanup target changed after approval: " + item.ID)
		return
	}
	if item.TargetFingerprintBefore == nil {
		add("cleanup item is missing its original target fingerprint: " + item.ID)
		return
	}
	validateCleanupFingerprint(root, item.ID, *item.TargetPath, *item.TargetFingerprintBefore, add)
}

func validateCleanupFingerprint(root, itemID, path, expected string, add func(string)) {
	actual, err := sha256RepositoryPath(root, path)
	if err != nil {
		add(fmt.Sprintf("cleanup item %s fingerprint path cannot be read: %v", itemID, err))
		return
	}
	if actual != expected {
		add("cleanup item content changed after the plan was created: " + itemID)
	}
}

func validateIgnoreRule(root string, item cleanupItemSemantic, add func(string)) {
	if item.TargetPath == nil || item.IgnorePattern == nil {
		return
	}
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(*item.TargetPath)))
	if err != nil {
		add(fmt.Sprintf("cleanup item %s ignore file cannot be read: %v", item.ID, err))
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == strings.TrimSpace(*item.IgnorePattern) {
			return
		}
	}
	add("cleanup item did not add its approved ignore rule: " + item.ID)
}

func validateCleanupCompletionEvidence(root string, plan workspaceCleanupSemantic, add func(string)) {
	final := plan.Status == "applied" || plan.Status == "partial" || plan.Status == "rolled-back"
	if final && (plan.CompletedGitSHA == nil || gitSHA(root) != *plan.CompletedGitSHA) {
		add("completed cleanup revision does not match the current Git revision")
	}
	for _, evidenceID := range plan.VerificationEvidenceIDs {
		expectedSHA := ""
		if final && plan.CompletedGitSHA != nil {
			expectedSHA = *plan.CompletedGitSHA
		}
		validateCleanupEvidenceRecord(root, plan.WorkItemID, evidenceID, expectedSHA, "passed", add)
	}
	for _, batch := range plan.Batches {
		if batch.Status == "verified" {
			validateCleanupBatchTiming(root, batch, add)
			validateCleanupBatchCommands(root, plan, batch.ID, "before", batch.BeforeCommands, batch.BeforeEvidenceIDs, plan.GitSHA, "passed", add)
			expectedSHA := ""
			if plan.CompletedGitSHA != nil {
				expectedSHA = *plan.CompletedGitSHA
			}
			validateCleanupBatchCommands(root, plan, batch.ID, "after", batch.AfterCommands, batch.AfterEvidenceIDs, expectedSHA, "passed", add)
			for _, evidenceID := range batch.AfterEvidenceIDs {
				if final && !contains(plan.VerificationEvidenceIDs, evidenceID) {
					add(fmt.Sprintf("verified cleanup batch %s evidence %s is missing from final verification records", batch.ID, evidenceID))
				}
			}
		}
		if batch.Status == "restored" {
			validateCleanupBatchTiming(root, batch, add)
			expectedSHA := ""
			if plan.CompletedGitSHA != nil {
				expectedSHA = *plan.CompletedGitSHA
			}
			validateCleanupBatchCommands(root, plan, batch.ID, "before", batch.BeforeCommands, batch.BeforeEvidenceIDs, plan.GitSHA, "passed", add)
			validateCleanupBatchCommands(root, plan, batch.ID, "after", batch.AfterCommands, batch.AfterEvidenceIDs, expectedSHA, "failed", add)
			validateCleanupBatchCommands(root, plan, batch.ID, "recovery", batch.RollbackCommands, batch.RecoveryEvidenceIDs, expectedSHA, "passed", add)
			for _, evidenceID := range batch.RecoveryEvidenceIDs {
				if final && !contains(plan.VerificationEvidenceIDs, evidenceID) {
					add(fmt.Sprintf("restored cleanup batch %s evidence %s is missing from final verification records", batch.ID, evidenceID))
				}
			}
		}
	}
}

func validateCleanupBatchTiming(root string, batch cleanupBatchSemantic, add func(string)) {
	if batch.MutationStartedAt == nil || batch.MutationCompletedAt == nil {
		add("cleanup batch is missing its mutation time window: " + batch.ID)
		return
	}
	started, startErr := time.Parse(time.RFC3339, *batch.MutationStartedAt)
	completed, completedErr := time.Parse(time.RFC3339, *batch.MutationCompletedAt)
	if startErr != nil || completedErr != nil || completed.Before(started) {
		add("cleanup batch has an invalid mutation time window: " + batch.ID)
		return
	}
	seen := map[string]string{}
	groups := []struct {
		phase string
		ids   []string
	}{
		{phase: "before", ids: batch.BeforeEvidenceIDs},
		{phase: "after", ids: batch.AfterEvidenceIDs},
		{phase: "recovery", ids: batch.RecoveryEvidenceIDs},
	}
	for _, group := range groups {
		for _, evidenceID := range group.ids {
			if previous := seen[evidenceID]; previous != "" {
				add(fmt.Sprintf("cleanup batch %s reuses evidence %s for %s and %s", batch.ID, evidenceID, previous, group.phase))
			}
			seen[evidenceID] = group.phase
			var evidence Evidence
			if readJSON(evidencePath(root, evidenceID), &evidence) != nil {
				continue
			}
			if group.phase == "before" {
				ended, err := time.Parse(time.RFC3339, evidence.EndedAt)
				if err != nil || ended.After(started) {
					add(fmt.Sprintf("cleanup batch %s before evidence %s was not completed before mutation", batch.ID, evidenceID))
				}
			} else {
				began, err := time.Parse(time.RFC3339, evidence.StartedAt)
				if err != nil || began.Before(completed) {
					add(fmt.Sprintf("cleanup batch %s %s evidence %s was not started after mutation", batch.ID, group.phase, evidenceID))
				}
			}
		}
	}
	if batch.Status == "restored" {
		latestFailureEnd := time.Time{}
		for _, evidenceID := range batch.AfterEvidenceIDs {
			var evidence Evidence
			if readJSON(evidencePath(root, evidenceID), &evidence) != nil {
				continue
			}
			ended, err := time.Parse(time.RFC3339, evidence.EndedAt)
			if err == nil && ended.After(latestFailureEnd) {
				latestFailureEnd = ended
			}
		}
		earliestRecoveryStart := time.Time{}
		for _, evidenceID := range batch.RecoveryEvidenceIDs {
			var evidence Evidence
			if readJSON(evidencePath(root, evidenceID), &evidence) != nil {
				continue
			}
			started, err := time.Parse(time.RFC3339, evidence.StartedAt)
			if err == nil && (earliestRecoveryStart.IsZero() || started.Before(earliestRecoveryStart)) {
				earliestRecoveryStart = started
			}
		}
		if latestFailureEnd.IsZero() || earliestRecoveryStart.IsZero() || earliestRecoveryStart.Before(latestFailureEnd) {
			add("restored cleanup batch does not prove failure before recovery: " + batch.ID)
		}
	}
}

func validateCleanupBatchCommands(root string, plan workspaceCleanupSemantic, batchID, phase string, commands, evidenceIDs []string, expectedSHA, expectedResult string, add func(string)) {
	actualCommands := map[string]bool{}
	for _, evidenceID := range evidenceIDs {
		evidence := validateCleanupEvidenceRecord(root, plan.WorkItemID, evidenceID, expectedSHA, expectedResult, add)
		if evidence != nil {
			actualCommands[strings.TrimSpace(strings.Join(evidence.Command, " "))] = true
		}
	}
	for _, command := range commands {
		if !actualCommands[strings.TrimSpace(command)] {
			add(fmt.Sprintf("cleanup batch %s has no trusted %s evidence for command: %s", batchID, phase, command))
		}
	}
}

func validateCleanupEvidenceRecord(root, workItemID, evidenceID, expectedSHA, expectedResult string, add func(string)) *Evidence {
	var evidence Evidence
	if err := readJSON(evidencePath(root, evidenceID), &evidence); err != nil {
		add("cleanup verification record cannot be read: " + evidenceID)
		return nil
	}
	if evidence.WorkItemID != workItemID {
		add("cleanup verification record belongs to another development task: " + evidenceID)
	}
	validResult := evidence.Result == "passed" && evidence.ExitCode == 0
	if expectedResult == "failed" {
		validResult = evidence.Result == "failed" && evidence.ExitCode != 0
	}
	if !validResult || evidence.Trust == "unverified" {
		add(fmt.Sprintf("cleanup verification record is not a trusted %s result: %s", expectedResult, evidenceID))
	}
	if expectedSHA != "" && evidence.GitSHA != expectedSHA {
		add("cleanup verification record does not match its expected Git revision: " + evidenceID)
	}
	logDigest, err := sha256File(filepath.Join(root, filepath.FromSlash(evidence.LogPath)))
	if err != nil || logDigest != evidence.LogSHA256 {
		add("cleanup verification log is missing or changed: " + evidenceID)
	}
	return &evidence
}
