package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func currentObjectStatus(status string) bool {
	return !contains([]string{"superseded", "archived", "rejected", "cancelled"}, status)
}

// Select a single current plan per version. A multi-version plan is projected
// into its own milestones rather than silently assigning it to the first one.
func currentVersionPlans(data boardData) []boardPlan {
	released := map[string]bool{}
	for _, release := range data.Releases {
		if release.Status == "released" {
			released[normalizeVersionLabel(release.Version)] = true
		}
	}
	chosen := map[string]boardPlan{}
	for _, plan := range data.Plans {
		if !currentObjectStatus(plan.Status) || plan.SupersededBy != nil {
			continue
		}
		versions := map[string][]boardMilestone{}
		for _, milestone := range plan.Milestones {
			version := normalizeVersionLabel(milestone.TargetRelease)
			if version == "" {
				version = planVersionLabel(plan, data)
			}
			versions[version] = append(versions[version], milestone)
		}
		if len(versions) == 0 {
			versions[planVersionLabel(plan, data)] = nil
		}
		for version, milestones := range versions {
			if version == "" || released[version] {
				continue
			}
			projected := plan
			projected.Milestones = append([]boardMilestone(nil), milestones...)
			for i := range projected.Milestones {
				projected.Milestones[i].TargetRelease = version
			}
			old, exists := chosen[version]
			if !exists || plan.UpdatedAt > old.UpdatedAt || (plan.UpdatedAt == old.UpdatedAt && (plan.Revision > old.Revision || plan.Revision == old.Revision && plan.ID > old.ID)) {
				chosen[version] = projected
			}
		}
	}
	result := make([]boardPlan, 0, len(chosen))
	for _, plan := range chosen {
		result = append(result, plan)
	}
	sort.Slice(result, func(i, j int) bool {
		return compareVersions(planVersionLabel(result[i], data), planVersionLabel(result[j], data)) < 0
	})
	return result
}

func validBoardName(name string) bool {
	if contains([]string{"STATUS.md", "ROADMAP.md", "CURRENT_STATE.md", "RELEASES.md", "PLANS.md"}, name) {
		return true
	}
	return strings.HasPrefix(name, "plans/v") && filepath.Ext(name) == ".md" && !strings.Contains(strings.TrimPrefix(name, "plans/"), "/") && !strings.Contains(name, "..") && !strings.Contains(name, "\\")
}

// Only generator-owned paths are reconciled. Archive obsolete views by content
// hash so retries are idempotent and unrelated user documents are untouched.
func reconcileBoardFiles(root string, files map[string]string) error {
	boardDir := filepath.Join(root, "docs", "board")
	var previous []string
	index := filepath.Join(root, ".ai-flow", "state", "generated-board-files.json")
	if err := readJSON(index, &previous); err != nil && !os.IsNotExist(err) {
		return err
	}
	// Adopt only the documented generated filenames from earlier versions.
	if len(previous) == 0 {
		for _, name := range []string{"STATUS.md", "ROADMAP.md", "CURRENT_STATE.md", "RELEASES.md", "PLANS.md"} {
			if _, err := os.Stat(filepath.Join(boardDir, name)); err == nil {
				previous = append(previous, name)
			}
		}
		old, _ := filepath.Glob(filepath.Join(boardDir, "plans", "v*.md"))
		for _, path := range old {
			previous = append(previous, "plans/"+filepath.Base(path))
		}
	}
	for name, content := range files {
		if !validBoardName(name) {
			return fmt.Errorf("invalid generated board name: %s", name)
		}
		if violations := lintMessage(stripHTMLCommentsForLint(content)); len(violations) > 0 {
			return fmt.Errorf("board contains internal wording: %s: %v", name, violations)
		}
	}
	for _, name := range previous {
		if !validBoardName(name) {
			return fmt.Errorf("invalid previous generated board name")
		}
		if _, keep := files[name]; keep {
			continue
		}
		path := filepath.Join(boardDir, filepath.FromSlash(name))
		content, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		digest, err := sha256File(path)
		if err != nil {
			return err
		}
		dest := filepath.Join(root, ".ai-flow", "archive", "generated", digest, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, content, 0644); err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	names := make([]string, 0, len(files))
	for name, content := range files {
		path := filepath.Join(boardDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := writeBoardFile(path, content); err != nil {
			return err
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return writeJSONAtomic(index, names)
}

func latestEvidence(evidence []Evidence) []Evidence {
	latest := map[string]Evidence{}
	for _, e := range evidence {
		key := e.WorkItemID + "/" + e.TestID
		old, ok := latest[key]
		stamp, _ := time.Parse(time.RFC3339Nano, e.CreatedAt)
		oldStamp, _ := time.Parse(time.RFC3339Nano, old.CreatedAt)
		if !ok || stamp.After(oldStamp) || stamp.Equal(oldStamp) && e.EndedAt > old.EndedAt || stamp.Equal(oldStamp) && e.EndedAt == old.EndedAt && e.ID > old.ID {
			latest[key] = e
		}
	}
	result := make([]Evidence, 0, len(latest))
	for _, e := range latest {
		result = append(result, e)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].TestID < result[j].TestID })
	return result
}

func evidenceResult(e Evidence) string {
	if !contains([]string{"verified-local", "verified-ci"}, e.Trust) {
		return "unverified"
	}
	return e.Result
}

func currentDecision(data boardData, decision boardDecision) bool {
	if !currentObjectStatus(decision.Status) {
		return false
	}
	goal := activeBoardGoal(data)
	if goal == nil {
		return true
	}
	if decision.GoalID != nil {
		return *decision.GoalID == goal.ID
	}
	for _, id := range decision.WorkItemIDs {
		if w := boardWorkByID(data, id); w != nil && w.GoalID != nil && *w.GoalID == goal.ID {
			return true
		}
	}
	for _, r := range data.Requirements {
		if r.GoalID == goal.ID && contains(decision.RequirementIDs, r.ID) {
			return true
		}
	}
	return len(decision.WorkItemIDs) == 0 && len(decision.RequirementIDs) == 0
}

func currentPlansIncludingReleased(data boardData) []boardPlan {
	data.Releases = nil
	return currentVersionPlans(data)
}
