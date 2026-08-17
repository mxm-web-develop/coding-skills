package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
)

type cleanupApprovalBinding struct {
	SchemaVersion        int                           `json:"schema_version"`
	Revision             int                           `json:"revision"`
	PlanID               string                        `json:"plan_id"`
	CreatedAt            string                        `json:"created_at"`
	GitSHA               string                        `json:"git_sha"`
	InventoryRevision    int                           `json:"inventory_revision"`
	InventorySHA256      string                        `json:"inventory_sha256"`
	BoundaryFingerprints []cleanupBoundaryFingerprint  `json:"boundary_fingerprints"`
	WorkItemID           string                        `json:"work_item_id"`
	RequestedIntent      string                        `json:"requested_intent"`
	Scope                cleanupScopeSemantic          `json:"scope"`
	Items                []cleanupItemApprovalBinding  `json:"items"`
	Batches              []cleanupBatchApprovalBinding `json:"batches"`
	Summary              cleanupSummarySemantic        `json:"summary"`
}

type cleanupItemApprovalBinding struct {
	ID                      string                         `json:"id"`
	SourcePath              string                         `json:"source_path"`
	SourceFingerprint       string                         `json:"source_fingerprint"`
	Tracked                 bool                           `json:"tracked"`
	PathType                string                         `json:"path_type"`
	Category                string                         `json:"category"`
	Classification          string                         `json:"classification"`
	Action                  string                         `json:"action"`
	TargetPath              *string                        `json:"target_path"`
	TargetExistedBefore     bool                           `json:"target_existed_before"`
	TargetFingerprintBefore *string                        `json:"target_fingerprint_before"`
	TargetFingerprintAfter  *string                        `json:"target_fingerprint_after"`
	OwningComponents        []string                       `json:"owning_components"`
	Evidence                []string                       `json:"evidence"`
	Risk                    string                         `json:"risk"`
	IgnorePattern           *string                        `json:"ignore_pattern"`
	Recovery                cleanupRecoveryApprovalBinding `json:"recovery"`
}

type cleanupRecoveryApprovalBinding struct {
	Kind      string  `json:"kind"`
	Reference *string `json:"reference"`
}

type cleanupBatchApprovalBinding struct {
	ID                 string   `json:"id"`
	ItemIDs            []string `json:"item_ids"`
	AffectedComponents []string `json:"affected_components"`
	Languages          []string `json:"languages"`
	Preconditions      []string `json:"preconditions"`
	BeforeCommands     []string `json:"before_commands"`
	AfterCommands      []string `json:"after_commands"`
	RollbackCommands   []string `json:"rollback_commands"`
}

func cleanupApprovalDigest(plan workspaceCleanupSemantic) (string, error) {
	binding := cleanupApprovalBinding{
		SchemaVersion:        plan.SchemaVersion,
		Revision:             plan.Revision,
		PlanID:               plan.PlanID,
		CreatedAt:            plan.CreatedAt,
		GitSHA:               plan.GitSHA,
		InventoryRevision:    plan.InventoryRevision,
		InventorySHA256:      plan.InventorySHA256,
		BoundaryFingerprints: plan.BoundaryFingerprints,
		WorkItemID:           plan.WorkItemID,
		RequestedIntent:      plan.RequestedIntent,
		Scope:                plan.Scope,
		Summary:              plan.Summary,
	}
	for _, item := range plan.Items {
		binding.Items = append(binding.Items, cleanupItemApprovalBinding{
			ID: item.ID, SourcePath: item.SourcePath, SourceFingerprint: item.SourceFingerprint,
			Tracked: item.Tracked, PathType: item.PathType, Category: item.Category,
			Classification: item.Classification, Action: item.Action, TargetPath: item.TargetPath,
			TargetExistedBefore: item.TargetExistedBefore, TargetFingerprintBefore: item.TargetFingerprintBefore, TargetFingerprintAfter: item.TargetFingerprintAfter,
			OwningComponents: item.OwningComponents, Evidence: item.Evidence, Risk: item.Risk, IgnorePattern: item.IgnorePattern,
			Recovery: cleanupRecoveryApprovalBinding{Kind: item.Recovery.Kind, Reference: item.Recovery.Reference},
		})
	}
	for _, batch := range plan.Batches {
		binding.Batches = append(binding.Batches, cleanupBatchApprovalBinding{
			ID: batch.ID, ItemIDs: batch.ItemIDs, AffectedComponents: batch.AffectedComponents,
			Languages: batch.Languages, Preconditions: batch.Preconditions,
			BeforeCommands: batch.BeforeCommands, AfterCommands: batch.AfterCommands,
			RollbackCommands: batch.RollbackCommands,
		})
	}
	data, err := json.Marshal(binding)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func runCleanup(args []string) error {
	if len(args) == 0 || args[0] != "digest" {
		return errors.New("usage: flowctl cleanup digest [--root PATH] --plan PATH")
	}
	fs := flag.NewFlagSet("cleanup digest", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	planArg := fs.String("plan", "", "cleanup plan path")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if strings.TrimSpace(*planArg) == "" {
		return errors.New("--plan is required")
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	planPath := *planArg
	if !filepath.IsAbs(planPath) {
		planPath = filepath.Join(root, filepath.FromSlash(planPath))
	}
	relativePlan, err := filepath.Rel(root, planPath)
	if err != nil || relativePlan == ".." || strings.HasPrefix(relativePlan, ".."+string(filepath.Separator)) {
		return errors.New("cleanup plan must be inside the project")
	}
	var plan workspaceCleanupSemantic
	if err := readSemanticJSON(planPath, &plan); err != nil {
		return err
	}
	digest, err := cleanupApprovalDigest(plan)
	if err != nil {
		return err
	}
	fmt.Println(digest)
	return nil
}
