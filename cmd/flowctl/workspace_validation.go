package main

import (
	"fmt"
	"path/filepath"
)

type workspaceStructureSemantic struct {
	Revision       int      `json:"revision"`
	GitSHA         string   `json:"git_sha"`
	Status         string   `json:"status"`
	ProtectedPaths []string `json:"protected_paths"`
	Repository     struct {
		WorkspaceManifests []string `json:"workspace_manifests"`
		NestedRepositories []string `json:"nested_repositories"`
		SharedRoots        []string `json:"shared_roots"`
	} `json:"repository"`
	Components []struct {
		ID                   string   `json:"id"`
		Root                 string   `json:"root"`
		Manifests            []string `json:"manifests"`
		DeploymentEvidence   []string `json:"deployment_evidence"`
		InternalDependencies []string `json:"internal_dependencies"`
	} `json:"components"`
	Candidates []struct {
		Path             string   `json:"path"`
		OwningComponents []string `json:"owning_components"`
	} `json:"candidates"`
}

type workspaceCleanupSemantic struct {
	SchemaVersion           int                          `json:"schema_version"`
	Revision                int                          `json:"revision"`
	PlanID                  string                       `json:"plan_id"`
	CreatedAt               string                       `json:"created_at"`
	GitSHA                  string                       `json:"git_sha"`
	InventoryRevision       int                          `json:"inventory_revision"`
	InventorySHA256         string                       `json:"inventory_sha256"`
	BoundaryFingerprints    []cleanupBoundaryFingerprint `json:"boundary_fingerprints"`
	WorkItemID              string                       `json:"work_item_id"`
	Status                  string                       `json:"status"`
	RequestedIntent         string                       `json:"requested_intent"`
	CompletedAt             *string                      `json:"completed_at"`
	CompletedGitSHA         *string                      `json:"completed_git_sha"`
	VerificationEvidenceIDs []string                     `json:"verification_evidence_ids"`
	Scope                   cleanupScopeSemantic         `json:"scope"`
	Items                   []cleanupItemSemantic        `json:"items"`
	Batches                 []cleanupBatchSemantic       `json:"batches"`
	Summary                 cleanupSummarySemantic       `json:"summary"`
	Approval                cleanupApprovalSemantic      `json:"approval"`
}

type cleanupBoundaryFingerprint struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type cleanupScopeSemantic struct {
	Paths      []string `json:"paths"`
	Components []string `json:"components"`
}

type cleanupItemSemantic struct {
	ID                      string                  `json:"id"`
	SourcePath              string                  `json:"source_path"`
	SourceFingerprint       string                  `json:"source_fingerprint"`
	Tracked                 bool                    `json:"tracked"`
	PathType                string                  `json:"path_type"`
	Category                string                  `json:"category"`
	Classification          string                  `json:"classification"`
	Action                  string                  `json:"action"`
	TargetPath              *string                 `json:"target_path"`
	TargetExistedBefore     bool                    `json:"target_existed_before"`
	TargetFingerprintBefore *string                 `json:"target_fingerprint_before"`
	TargetFingerprintAfter  *string                 `json:"target_fingerprint_after"`
	OwningComponents        []string                `json:"owning_components"`
	Evidence                []string                `json:"evidence"`
	Risk                    string                  `json:"risk"`
	IgnorePattern           *string                 `json:"ignore_pattern"`
	Approval                string                  `json:"approval"`
	Result                  string                  `json:"result"`
	Recovery                cleanupRecoverySemantic `json:"recovery"`
}

type cleanupRecoverySemantic struct {
	Kind      string  `json:"kind"`
	Reference *string `json:"reference"`
	Verified  bool    `json:"verified"`
}

type cleanupBatchSemantic struct {
	ID                  string   `json:"id"`
	ItemIDs             []string `json:"item_ids"`
	AffectedComponents  []string `json:"affected_components"`
	Languages           []string `json:"languages"`
	Preconditions       []string `json:"preconditions"`
	BeforeCommands      []string `json:"before_commands"`
	BeforeEvidenceIDs   []string `json:"before_evidence_ids"`
	AfterCommands       []string `json:"after_commands"`
	AfterEvidenceIDs    []string `json:"after_evidence_ids"`
	RollbackCommands    []string `json:"rollback_commands"`
	RecoveryEvidenceIDs []string `json:"recovery_evidence_ids"`
	MutationStartedAt   *string  `json:"mutation_started_at"`
	MutationCompletedAt *string  `json:"mutation_completed_at"`
	Status              string   `json:"status"`
}

type cleanupSummarySemantic struct {
	AffectedComponents []string `json:"affected_components"`
	ArchiveCount       int      `json:"archive_count"`
	RemovalCount       int      `json:"removal_count"`
	IgnoreRuleCount    int      `json:"ignore_rule_count"`
	KeptCount          int      `json:"kept_count"`
	ResidualRisks      []string `json:"residual_risks"`
}

type cleanupApprovalSemantic struct {
	Required      bool    `json:"required"`
	Status        string  `json:"status"`
	ApprovedBy    *string `json:"approved_by"`
	ApprovedAt    *string `json:"approved_at"`
	MappingSHA256 *string `json:"mapping_sha256"`
}

func validateWorkspaceStructureInventory(root string) []validationIssue {
	path := filepath.Join(root, ".ai-flow", "baseline", "workspace-structure-inventory.json")
	var inventory workspaceStructureSemantic
	if err := readSemanticJSON(path, &inventory); err != nil {
		return nil
	}
	issues := []validationIssue{}
	componentIDs := map[string]bool{}
	componentRoots := map[string]bool{}
	add := func(message string) {
		issues = append(issues, validationIssue{Path: relativeDisplay(root, path), Schema: "semantic-links", Message: message})
	}
	if inventory.Status == "stale" {
		add("workspace structure inventory is stale")
	}
	for _, component := range inventory.Components {
		if componentIDs[component.ID] {
			add("duplicate component id: " + component.ID)
		}
		componentIDs[component.ID] = true
		if componentRoots[component.Root] {
			add("duplicate component root: " + component.Root)
		}
		componentRoots[component.Root] = true
	}
	for _, component := range inventory.Components {
		for _, dependency := range component.InternalDependencies {
			if !componentIDs[dependency] {
				add(fmt.Sprintf("component %s references missing dependency %s", component.ID, dependency))
			}
			if dependency == component.ID {
				add("component cannot depend on itself: " + component.ID)
			}
		}
	}
	candidatePaths := map[string]bool{}
	for _, candidate := range inventory.Candidates {
		if candidatePaths[candidate.Path] {
			add("duplicate cleanup candidate path: " + candidate.Path)
		}
		candidatePaths[candidate.Path] = true
		for _, componentID := range candidate.OwningComponents {
			if !componentIDs[componentID] {
				add(fmt.Sprintf("candidate %s references missing component %s", candidate.Path, componentID))
			}
		}
		if err := ensurePathInsideRepository(root, candidate.Path); err != nil {
			add(fmt.Sprintf("candidate path %s is unsafe: %v", candidate.Path, err))
		}
	}
	return issues
}
