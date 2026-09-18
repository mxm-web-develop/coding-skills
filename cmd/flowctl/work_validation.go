package main

import "path/filepath"

func validateWorkLifecycle(root string) []validationIssue {
	issues := []validationIssue{}
	graph := loadTraceGraph(root)
	add := func(path, message string) {
		issues = append(issues, validationIssue{Path: relativeDisplay(root, path), Schema: "work-lifecycle", Message: message})
	}
	for id, obj := range graph.workItems {
		work := obj.Value
		if work.MilestoneID != "" {
			found := false
			for _, plan := range graph.plans {
				if !contains(plan.Value.WorkItemIDs, id) {
					continue
				}
				for _, ms := range plan.Value.Milestones {
					if ms.ID == work.MilestoneID {
						found = true
					}
				}
			}
			if !found {
				add(obj.Path, "task milestone is missing from its linked plan")
			}
		}
		for _, dependency := range work.Dependencies {
			if dependency == id {
				add(obj.Path, "task depends on itself")
			}
			if _, ok := graph.workItems[dependency]; !ok {
				add(obj.Path, "missing dependency: "+dependency)
			}
		}
		if work.WorkflowVersion < 1 {
			continue
		} // Historical completion is preserved, never fabricated as a v1 acceptance.
		if contains([]string{"awaiting_acceptance", "accepted", "closing", "done"}, work.Status) {
			if work.Review == nil || work.Review.Decision != "approved" {
				add(obj.Path, "task needs an approved review")
			}
			if work.Acceptance == nil || len(work.Acceptance.Steps) == 0 || len(work.Acceptance.Steps) != len(work.Acceptance.Expected) {
				add(obj.Path, "task needs executable human acceptance instructions")
			}
		}
		if contains([]string{"accepted", "closing", "done"}, work.Status) {
			if work.Acceptance == nil || work.Acceptance.Status != "passed" || work.Acceptance.VerifiedBy == "" {
				add(obj.Path, "task has not passed user acceptance")
				continue
			}
			if work.Review == nil || work.Review.GitSHA != work.Acceptance.GitSHA || work.Review.ContentSHA256 != work.Acceptance.ContentSHA256 {
				add(obj.Path, "review and acceptance cover different code")
			}
			if work.Status != "done" {
				if err := validateWorkEvidence(root, work); err != nil {
					add(obj.Path, err.Error())
				}
			}
			if work.Status == "done" && (work.ArchivePath == "" || work.CompletionSummary == "") {
				add(obj.Path, "completed task has not been archived with a capability summary")
			}
		}
	}
	// DFS catches dependency cycles before any member can start.
	visiting, visited := map[string]bool{}, map[string]bool{}
	var visit func(string)
	visit = func(id string) {
		if visiting[id] {
			add(filepath.Join(root, ".ai-flow", "work-items", id+".json"), "task dependency cycle")
			return
		}
		if visited[id] {
			return
		}
		visiting[id] = true
		for _, dep := range graph.workItems[id].Value.Dependencies {
			visit(dep)
		}
		visiting[id] = false
		visited[id] = true
	}
	for id := range graph.workItems {
		visit(id)
	}
	return issues
}
