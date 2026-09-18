package main

import (
	"errors"
	"strings"
	"time"
)

func runWorkReopen(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work reopen")
	reason := fs.String("reason", "", "why current verification must be repeated")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if !contains([]string{"ready_for_review", "awaiting_acceptance", "accepted"}, item.Status) || strings.TrimSpace(*reason) == "" {
		return errors.New("reopen requires an unfinished reviewed task and a reason")
	}
	item.Status = "in_progress"
	item.Review = nil
	item.Acceptance = nil
	return saveWorkMutation(root, &item, "work.reopened", map[string]any{"reason": *reason})
}

func runWorkBudget(args []string) error {
	fs, rootArg, id, expected := workMutationFlags("work budget")
	elapsed := fs.Int("max-elapsed-minutes", 0, "total elapsed minutes since run start")
	retries := fs.Int("max-retries", -1, "retries per failing test")
	files := fs.Int("max-changed-files", 0, "maximum changed files")
	reason := fs.String("reason", "", "diagnosis or scope change supporting this revision")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, item, err := loadWorkMutation(*rootArg, *id, *expected)
	if err != nil {
		return err
	}
	if item.RunID == nil || contains([]string{"done", "cancelled", "closing"}, item.Status) || strings.TrimSpace(*reason) == "" {
		return errors.New("an active run and budget revision reason are required")
	}
	run, err := readRun(root, *item.RunID)
	if err != nil {
		return err
	}
	if *elapsed < 0 || *retries < -1 || *files < 0 {
		return errors.New("invalid budget")
	}
	if *elapsed == 0 && *retries == -1 && *files == 0 {
		return errors.New("provide at least one budget change")
	}
	if *elapsed > 0 {
		run.Budgets.MaxElapsedMinutes = *elapsed
	}
	if *retries >= 0 {
		run.Budgets.MaxRetries = *retries
	}
	if *files > 0 {
		run.Budgets.MaxChangedFiles = *files
	}
	run.Revision++
	run.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	if err := writeJSONAtomic(runPath(root, run.ID), run); err != nil {
		return err
	}
	return appendEvent(root, "run.budget_revised", "run", run.ID, run.ID, run.Revision, map[string]any{"reason": *reason, "budgets": run.Budgets})
}
