package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func runWorkReview(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work review")
	reviewer := fs.String("reviewer", "", "reviewer identity")
	decision := fs.String("decision", "", "approved or changes_required")
	summary := fs.String("summary", "", "review conclusion")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if !contains([]string{"in_progress", "ready_for_review"}, item.Status) {
		return errors.New("task is not ready for review")
	}
	if strings.TrimSpace(*reviewer) == "" || strings.TrimSpace(*summary) == "" || !contains([]string{"approved", "changes_required"}, *decision) {
		return errors.New("reviewer, summary and review decision are required")
	}
	if item.Execution != nil {
		if *reviewer != item.Execution.Reviewer || *reviewer == item.Execution.Executor {
			return errors.New("review must be performed by the assigned independent reviewer")
		}
		if err := checkExecution(root, item, true); err != nil {
			return err
		}
	}
	if *decision == "approved" {
		if err := validateWorkEvidence(root, item); err != nil {
			return err
		}
	}
	fingerprint, err := verificationFingerprint(root)
	if err != nil {
		return err
	}
	item.Review = &WorkReview{Reviewer: *reviewer, Decision: *decision, Summary: *summary, GitSHA: gitSHA(root), ContentSHA256: fingerprint, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	item.Status = "ready_for_review"
	item.Acceptance = nil
	if *decision == "changes_required" {
		item.Status = "in_progress"
	}
	return saveWorkMutation(root, &item, "work.reviewed", map[string]any{"decision": *decision})
}

func runWorkAcceptance(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work acceptance")
	instructions := fs.String("instructions", "", "how to open the result and prepare verification")
	var steps, results, limitations stringListFlag
	fs.Var(&steps, "step", "user action; repeatable")
	fs.Var(&results, "expected", "expected result for each step; repeatable")
	fs.Var(&limitations, "limitation", "known limitation; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.Status != "ready_for_review" {
		return errors.New("review the task before requesting acceptance")
	}
	if strings.TrimSpace(*instructions) == "" || len(steps) == 0 || len(steps) != len(results) {
		return errors.New("instructions and matching step/expected pairs are required")
	}
	if item.Review == nil || item.Review.Decision != "approved" || !qualityMatches(root, item.Review.GitSHA, item.Review.ContentSHA256) {
		return errors.New("current code has no approved review")
	}
	if err := validateWorkEvidence(root, item); err != nil {
		return err
	}
	fingerprint, err := verificationFingerprint(root)
	if err != nil {
		return err
	}
	item.Acceptance = &WorkAcceptance{Status: "pending", Instructions: *instructions, Steps: steps, Expected: results, KnownLimitations: limitations, GitSHA: gitSHA(root), ContentSHA256: fingerprint, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	item.Status = "awaiting_acceptance"
	if err := saveWorkMutation(root, &item, "work.acceptance_requested", nil); err != nil {
		return err
	}
	return renderBoard(root)
}

func runWorkAccept(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work accept")
	by := fs.String("by", "", "person who explicitly verified this result")
	feedback := fs.String("feedback", "", "the user's actual feedback")
	result := fs.String("result", "", "passed or failed")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.Status != "awaiting_acceptance" || item.Acceptance == nil {
		return errors.New("no pending acceptance for this task")
	}
	if strings.TrimSpace(*by) == "" || strings.TrimSpace(*feedback) == "" || !contains([]string{"passed", "failed"}, *result) {
		return errors.New("record the user's identity, actual feedback and passed/failed result")
	}
	if !qualityMatches(root, item.Acceptance.GitSHA, item.Acceptance.ContentSHA256) {
		return errors.New("code changed after the acceptance instructions; verify and present the new result again")
	}
	if *result == "passed" {
		if err := validateWorkEvidence(root, item); err != nil {
			return err
		}
	}
	item.Acceptance.Status = *result
	item.Acceptance.VerifiedBy = *by
	item.Acceptance.Feedback = *feedback
	item.Acceptance.RecordedAt = time.Now().UTC().Format(time.RFC3339Nano)
	item.Status = "accepted"
	if *result == "failed" {
		item.Status = "in_progress"
		item.Review = nil
	}
	return saveWorkMutation(root, &item, "work.acceptance_recorded", map[string]any{"result": *result, "by": *by})
}

func runWorkComplete(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work complete")
	summary := fs.String("summary", "", "current capabilities and remaining limitations")
	var evidenceIDs stringListFlag
	fs.Var(&evidenceIDs, "evidence", "compatibility flag; all required checks are always validated")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	unlock, err := projectMutationLock(root)
	if err != nil {
		return err
	}
	defer unlock()
	item, err = readWorkItem(root, *id)
	if err != nil {
		return err
	}
	if err := checkExpectedRevision(item.Revision, *expected); err != nil {
		return err
	}
	if item.Status == "done" && item.ArchivePath != "" {
		if err := finishArchiveRemoval(root, item); err != nil {
			return err
		}
		return renderBoard(root)
	}
	if !contains([]string{"accepted", "closing"}, item.Status) {
		return errors.New("user acceptance must pass before completion")
	}
	if item.Review == nil || item.Review.Decision != "approved" || !qualityMatches(root, item.Review.GitSHA, item.Review.ContentSHA256) {
		return errors.New("review is missing or stale")
	}
	if item.Acceptance == nil || item.Acceptance.Status != "passed" || item.Acceptance.VerifiedBy == "" || !qualityMatches(root, item.Acceptance.GitSHA, item.Acceptance.ContentSHA256) {
		return errors.New("user acceptance is missing or stale")
	}
	if err := validateWorkEvidence(root, item); err != nil {
		return err
	}
	if *summary != "" {
		item.CompletionSummary = strings.TrimSpace(*summary)
	}
	if item.CompletionSummary == "" {
		return errors.New("a completion summary is required")
	}
	item.Status = "closing"
	if err := saveWorkMutation(root, &item, "work.closing", nil); err != nil {
		return err
	}
	if err := archiveCompletedWork(root, &item); err != nil {
		return fmt.Errorf("acceptance passed; closeout is incomplete and can be retried: %w", err)
	}
	if err := os.Remove(leasePath(root, item.ID)); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := appendEvent(root, "work.completed", "work-item", item.ID, "", item.Revision, map[string]any{"archive": item.ArchivePath}); err != nil {
		return err
	}
	if err := renderBoard(root); err != nil {
		return err
	}
	return printJSON(item)
}
