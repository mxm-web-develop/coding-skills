package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func validateWorkspaceCleanupPlans(root string) []validationIssue {
	paths, _ := filepath.Glob(filepath.Join(root, ".ai-flow", "workspace-cleanup", "PLAN-*.json"))
	sort.Strings(paths)
	issues := []validationIssue{}
	structurePath := filepath.Join(root, ".ai-flow", "baseline", "workspace-structure-inventory.json")
	var structure workspaceStructureSemantic
	hasStructure := readSemanticJSON(structurePath, &structure) == nil
	knownComponents := map[string]bool{}
	for _, component := range structure.Components {
		knownComponents[component.ID] = true
	}
	for _, path := range paths {
		var plan workspaceCleanupSemantic
		if readSemanticJSON(path, &plan) != nil {
			continue
		}
		add := func(message string) {
			issues = append(issues, validationIssue{Path: relativeDisplay(root, path), Schema: "semantic-links", Message: message})
		}
		if !hasStructure {
			add("cleanup plan requires a current workspace structure inventory")
		}
		if hasStructure && structure.Status != "confirmed" {
			add("cleanup plan requires a confirmed workspace structure inventory")
		}
		if hasStructure && plan.InventoryRevision != structure.Revision {
			add(fmt.Sprintf("cleanup plan inventory revision %d does not match current revision %d", plan.InventoryRevision, structure.Revision))
		}
		if hasStructure && plan.GitSHA != structure.GitSHA {
			add("cleanup plan and workspace structure inventory describe different Git revisions")
		}
		if hasStructure {
			validateCleanupInventoryAndBoundaries(root, structurePath, plan, structure, add)
		}
		if _, err := os.Stat(workItemPath(root, plan.WorkItemID)); os.IsNotExist(err) {
			add("cleanup plan references missing development task: " + plan.WorkItemID)
		}
		if (plan.Status == "proposed" || plan.Status == "approved" || plan.Status == "executing") && plan.GitSHA != gitSHA(root) {
			add("cleanup plan was created for a different Git revision")
		}
		validateCleanupApprovalBinding(plan, add)
		validateCleanupPathFingerprints(root, plan, add)
		scopeComponents := stringSet(plan.Scope.Components)
		for componentID := range scopeComponents {
			if hasStructure && !knownComponents[componentID] {
				add("cleanup scope references missing component: " + componentID)
			}
		}
		items := map[string]int{}
		sourcePaths := map[string]bool{}
		archiveCount, removalCount, ignoreRuleCount, keptCount := 0, 0, 0, 0
		for index, item := range plan.Items {
			if _, exists := items[item.ID]; exists {
				add("duplicate cleanup item id: " + item.ID)
			}
			items[item.ID] = index
			if sourcePaths[item.SourcePath] {
				add("duplicate cleanup source path: " + item.SourcePath)
			}
			sourcePaths[item.SourcePath] = true
			if !pathCoveredByScope(item.SourcePath, plan.Scope.Paths) {
				add(fmt.Sprintf("cleanup item %s is outside declared scope", item.ID))
			}
			if err := ensurePathInsideRepository(root, item.SourcePath); err != nil {
				add(fmt.Sprintf("cleanup item %s source is unsafe: %v", item.ID, err))
			}
			if item.Action != "keep" && hasStructure {
				validateCleanupStructureBoundaries(item.ID, item.SourcePath, structure, add)
			}
			for _, componentID := range item.OwningComponents {
				if !scopeComponents[componentID] {
					add(fmt.Sprintf("cleanup item %s references component outside scope: %s", item.ID, componentID))
				}
			}
			switch item.Action {
			case "archive-code", "archive-file":
				archiveCount++
				if item.TargetPath == nil || !archiveTargetPreservesPath(item.Action, item.SourcePath, *item.TargetPath) {
					add(fmt.Sprintf("cleanup item %s archive target must preserve its original relative path", item.ID))
				}
				if item.TargetFingerprintAfter == nil || *item.TargetFingerprintAfter != item.SourceFingerprint {
					add(fmt.Sprintf("cleanup item %s archive target fingerprint must equal its source fingerprint", item.ID))
				}
			case "remove-generated":
				removalCount++
			case "add-ignore-rule":
				ignoreRuleCount++
			case "keep":
				keptCount++
			}
			if item.TargetPath != nil {
				if err := ensurePathInsideRepository(root, *item.TargetPath); err != nil {
					add(fmt.Sprintf("cleanup item %s target is unsafe: %v", item.ID, err))
				}
				if item.Action != "keep" && hasStructure {
					validateCleanupStructureBoundaries(item.ID, *item.TargetPath, structure, add)
				}
			}
		}
		batchIDs := map[string]bool{}
		itemBatch := map[string]string{}
		for _, batch := range plan.Batches {
			if batchIDs[batch.ID] {
				add("duplicate cleanup batch id: " + batch.ID)
			}
			batchIDs[batch.ID] = true
			batchComponents := stringSet(batch.AffectedComponents)
			for componentID := range batchComponents {
				if !scopeComponents[componentID] {
					add(fmt.Sprintf("cleanup batch %s references component outside scope: %s", batch.ID, componentID))
				}
			}
			for _, itemID := range batch.ItemIDs {
				index, exists := items[itemID]
				if !exists {
					add(fmt.Sprintf("cleanup batch %s references missing item %s", batch.ID, itemID))
					continue
				}
				if previous, duplicate := itemBatch[itemID]; duplicate {
					add(fmt.Sprintf("cleanup item %s appears in both %s and %s", itemID, previous, batch.ID))
				}
				itemBatch[itemID] = batch.ID
				item := plan.Items[index]
				if item.Action == "keep" {
					add(fmt.Sprintf("keep-only item %s must not be placed in an execution batch", itemID))
				}
				for _, componentID := range item.OwningComponents {
					if !batchComponents[componentID] {
						add(fmt.Sprintf("cleanup batch %s does not include component %s owned by item %s", batch.ID, componentID, itemID))
					}
				}
			}
		}
		for _, item := range plan.Items {
			if item.Action != "keep" && itemBatch[item.ID] == "" {
				add("mutating cleanup item is not assigned to a batch: " + item.ID)
			}
		}
		if archiveCount != plan.Summary.ArchiveCount || removalCount != plan.Summary.RemovalCount || ignoreRuleCount != plan.Summary.IgnoreRuleCount || keptCount != plan.Summary.KeptCount {
			add("cleanup summary counts do not match plan items")
		}
		validateCleanupStatus(plan, add)
		validateCleanupCompletionEvidence(root, plan, add)
	}
	return issues
}

func validateCleanupStatus(plan workspaceCleanupSemantic, add func(string)) {
	expectedResult := func(action string) string {
		switch action {
		case "archive-code", "archive-file":
			return "moved"
		case "remove-generated":
			return "removed"
		case "add-ignore-rule":
			return "updated"
		default:
			return "kept"
		}
	}
	for _, item := range plan.Items {
		if item.Action == "keep" {
			continue
		}
		switch plan.Status {
		case "approved", "executing", "applied", "partial", "rolled-back":
			if item.Approval != "approved" {
				add(fmt.Sprintf("cleanup status %s requires path approval for item %s", plan.Status, item.ID))
			}
		}
		if plan.Status == "approved" && item.Result != "pending" {
			add("approved cleanup item must remain pending until execution: " + item.ID)
		}
		if plan.Status == "applied" && item.Result != expectedResult(item.Action) {
			add(fmt.Sprintf("applied cleanup item %s has result %s, want %s", item.ID, item.Result, expectedResult(item.Action)))
		}
		if plan.Status == "partial" && item.Result == "pending" {
			add("partial cleanup plan cannot leave a mutating item pending: " + item.ID)
		}
		if plan.Status == "rolled-back" && item.Result != "restored" && item.Result != "skipped" {
			add("rolled-back cleanup item must be restored or skipped: " + item.ID)
		}
		if plan.Status == "rejected" && (item.Result == "moved" || item.Result == "removed" || item.Result == "updated") {
			add("rejected cleanup plan contains an applied mutation: " + item.ID)
		}
	}
	for _, batch := range plan.Batches {
		switch plan.Status {
		case "approved":
			if batch.Status != "pending" {
				add("approved cleanup batch must remain pending: " + batch.ID)
			}
		case "applied":
			if batch.Status != "verified" {
				add("applied cleanup batch must be verified: " + batch.ID)
			}
		case "partial":
			if batch.Status == "pending" || batch.Status == "running" {
				add("partial cleanup plan cannot contain an unfinished batch: " + batch.ID)
			}
		case "rolled-back":
			if batch.Status != "restored" && batch.Status != "skipped" {
				add("rolled-back cleanup batch must be restored or skipped: " + batch.ID)
			}
		}
	}
}
