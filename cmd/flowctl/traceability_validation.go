package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type traceGoal struct {
	ID           string   `json:"id"`
	Supersedes   []string `json:"supersedes"`
	SupersededBy *string  `json:"superseded_by"`
}

type traceRequirement struct {
	ID           string   `json:"id"`
	GoalID       string   `json:"goal_id"`
	Dependencies []string `json:"dependencies"`
	TestIDs      []string `json:"test_ids"`
	Supersedes   []string `json:"supersedes"`
	SupersededBy *string  `json:"superseded_by"`
}

type tracePlan struct {
	ID          string           `json:"id"`
	GoalID      string           `json:"goal_id"`
	Milestones  []traceMilestone `json:"milestones"`
	WorkItemIDs []string         `json:"work_item_ids"`
}

type traceMilestone struct {
	RequirementIDs []string `json:"requirement_ids"`
}

type traceDecision struct {
	ID             string   `json:"id"`
	GoalID         *string  `json:"goal_id"`
	RequirementIDs []string `json:"requirement_ids"`
	WorkItemIDs    []string `json:"work_item_ids"`
	Supersedes     []string `json:"supersedes"`
	SupersededBy   *string  `json:"superseded_by"`
}

type traceTest struct {
	ID             string   `json:"id"`
	RequirementIDs []string `json:"requirement_ids"`
	WorkItemID     string   `json:"work_item_id"`
	Status         string   `json:"status"`
	EvidenceIDs    []string `json:"evidence_ids"`
}

type traceRelease struct {
	ID          string   `json:"id"`
	Status      string   `json:"status"`
	WorkItemIDs []string `json:"work_item_ids"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type traceObject[T any] struct {
	Path  string
	Value T
}

type traceGraph struct {
	goals        map[string]traceObject[traceGoal]
	requirements map[string]traceObject[traceRequirement]
	plans        map[string]traceObject[tracePlan]
	decisions    map[string]traceObject[traceDecision]
	workItems    map[string]traceObject[WorkItem]
	tests        map[string]traceObject[traceTest]
	runs         map[string]traceObject[HarnessRun]
	checkpoints  map[string]traceObject[Checkpoint]
	evidence     map[string]traceObject[Evidence]
	releases     map[string]traceObject[traceRelease]
}

func validateTraceability(root string) []validationIssue {
	graph := loadTraceGraph(root)
	issues := validateTraceStorage(root)
	add := func(path, message string) {
		issues = append(issues, validationIssue{Path: relativeDisplay(root, path), Schema: "semantic-links", Message: message})
	}
	require := func(sourcePath, kind, id string, found bool) {
		if !found {
			add(sourcePath, fmt.Sprintf("missing linked %s: %s", kind, id))
		}
	}

	state, err := readFlatYAML(filepath.Join(root, ".ai-flow", "state", "current.yaml"))
	if err == nil && state["active_goal"] != "" && state["active_goal"] != "none" {
		_, found := graph.goals[state["active_goal"]]
		require(filepath.Join(root, ".ai-flow", "state", "current.yaml"), "goal", state["active_goal"], found)
	}

	for id, object := range graph.requirements {
		_, found := graph.goals[object.Value.GoalID]
		require(object.Path, "goal", object.Value.GoalID, found)
		for _, dependencyID := range object.Value.Dependencies {
			_, found = graph.requirements[dependencyID]
			require(object.Path, "requirement", dependencyID, found)
		}
		for _, testID := range object.Value.TestIDs {
			test, testFound := graph.tests[testID]
			require(object.Path, "test", testID, testFound)
			if testFound && !contains(test.Value.RequirementIDs, id) {
				add(object.Path, "linked test does not point back to requirement: "+testID)
			}
		}
		validateReplacementLinks(object.Path, id, object.Value.Supersedes, object.Value.SupersededBy, graph.requirements, add)
	}
	for id, object := range graph.goals {
		validateReplacementLinks(object.Path, id, object.Value.Supersedes, object.Value.SupersededBy, graph.goals, add)
	}
	for id, object := range graph.decisions {
		if object.Value.GoalID != nil {
			_, found := graph.goals[*object.Value.GoalID]
			require(object.Path, "goal", *object.Value.GoalID, found)
		}
		for _, requirementID := range object.Value.RequirementIDs {
			_, found := graph.requirements[requirementID]
			require(object.Path, "requirement", requirementID, found)
		}
		for _, workID := range object.Value.WorkItemIDs {
			_, found := graph.workItems[workID]
			require(object.Path, "development task", workID, found)
		}
		validateReplacementLinks(object.Path, id, object.Value.Supersedes, object.Value.SupersededBy, graph.decisions, add)
	}
	for _, object := range graph.plans {
		_, found := graph.goals[object.Value.GoalID]
		require(object.Path, "goal", object.Value.GoalID, found)
		for _, milestone := range object.Value.Milestones {
			for _, requirementID := range milestone.RequirementIDs {
				_, found = graph.requirements[requirementID]
				require(object.Path, "requirement", requirementID, found)
			}
		}
		for _, workID := range object.Value.WorkItemIDs {
			work, workFound := graph.workItems[workID]
			require(object.Path, "development task", workID, workFound)
			if workFound && work.Value.GoalID != nil && *work.Value.GoalID != object.Value.GoalID {
				add(object.Path, "plan and linked development task belong to different goals: "+workID)
			}
		}
	}
	for _, object := range graph.workItems {
		if object.Value.GoalID != nil {
			_, found := graph.goals[*object.Value.GoalID]
			require(object.Path, "goal", *object.Value.GoalID, found)
		}
		for _, requirementID := range object.Value.RequirementIDs {
			_, found := graph.requirements[requirementID]
			require(object.Path, "requirement", requirementID, found)
		}
		if object.Value.RunID != nil {
			run, found := graph.runs[*object.Value.RunID]
			require(object.Path, "development run", *object.Value.RunID, found)
			if found && run.Value.WorkItemID != object.Value.ID {
				add(object.Path, "linked development run belongs to another task: "+run.Value.ID)
			}
		}
		for _, evidenceID := range object.Value.EvidenceIDs {
			evidence, found := graph.evidence[evidenceID]
			require(object.Path, "verification record", evidenceID, found)
			if found && evidence.Value.WorkItemID != object.Value.ID {
				add(object.Path, "linked verification record belongs to another task: "+evidenceID)
			}
		}
	}
	for _, object := range graph.tests {
		work, found := graph.workItems[object.Value.WorkItemID]
		require(object.Path, "development task", object.Value.WorkItemID, found)
		for _, requirementID := range object.Value.RequirementIDs {
			requirement, requirementFound := graph.requirements[requirementID]
			require(object.Path, "requirement", requirementID, requirementFound)
			if requirementFound && !contains(requirement.Value.TestIDs, object.Value.ID) {
				add(object.Path, "linked requirement does not point back to test: "+requirementID)
			}
			if found && !contains(work.Value.RequirementIDs, requirementID) {
				add(object.Path, "test covers a requirement outside its development task: "+requirementID)
			}
		}
		for _, evidenceID := range object.Value.EvidenceIDs {
			evidence, evidenceFound := graph.evidence[evidenceID]
			require(object.Path, "verification record", evidenceID, evidenceFound)
			if evidenceFound && (evidence.Value.TestID != object.Value.ID || evidence.Value.WorkItemID != object.Value.WorkItemID) {
				add(object.Path, "linked verification record belongs to another test or task: "+evidenceID)
			}
		}
	}
	for _, object := range graph.runs {
		_, found := graph.workItems[object.Value.WorkItemID]
		require(object.Path, "development task", object.Value.WorkItemID, found)
		for _, checkpointID := range object.Value.CheckpointIDs {
			checkpoint, checkpointFound := graph.checkpoints[checkpointID]
			require(object.Path, "saved progress", checkpointID, checkpointFound)
			if checkpointFound && (checkpoint.Value.RunID != object.Value.ID || checkpoint.Value.WorkItemID != object.Value.WorkItemID) {
				add(object.Path, "linked saved progress belongs to another run or task: "+checkpointID)
			}
		}
		for _, evidenceID := range object.Value.EvidenceIDs {
			evidence, evidenceFound := graph.evidence[evidenceID]
			require(object.Path, "verification record", evidenceID, evidenceFound)
			if evidenceFound && (evidence.Value.RunID != object.Value.ID || evidence.Value.WorkItemID != object.Value.WorkItemID) {
				add(object.Path, "linked verification record belongs to another run or task: "+evidenceID)
			}
		}
	}
	for _, object := range graph.checkpoints {
		run, runFound := graph.runs[object.Value.RunID]
		require(object.Path, "development run", object.Value.RunID, runFound)
		_, workFound := graph.workItems[object.Value.WorkItemID]
		require(object.Path, "development task", object.Value.WorkItemID, workFound)
		if runFound && (!contains(run.Value.CheckpointIDs, object.Value.ID) || run.Value.WorkItemID != object.Value.WorkItemID) {
			add(object.Path, "saved progress is not linked consistently from its run")
		}
	}
	for _, object := range graph.evidence {
		work, workFound := graph.workItems[object.Value.WorkItemID]
		require(object.Path, "development task", object.Value.WorkItemID, workFound)
		run, runFound := graph.runs[object.Value.RunID]
		require(object.Path, "development run", object.Value.RunID, runFound)
		if workFound && !contains(work.Value.EvidenceIDs, object.Value.ID) {
			add(object.Path, "development task does not point back to verification record")
		}
		if runFound && (!contains(run.Value.EvidenceIDs, object.Value.ID) || run.Value.WorkItemID != object.Value.WorkItemID) {
			add(object.Path, "development run does not point back consistently to verification record")
		}
		if requireObjectID(object.Value.TestID, "TEST") == nil {
			test, testFound := graph.tests[object.Value.TestID]
			require(object.Path, "test", object.Value.TestID, testFound)
			if testFound && (!contains(test.Value.EvidenceIDs, object.Value.ID) || test.Value.WorkItemID != object.Value.WorkItemID) {
				add(object.Path, "test does not point back consistently to verification record")
			}
		}
		if err := validateEvidenceLogPath(root, object.Value.LogPath); err != nil {
			add(object.Path, "verification log is unavailable or unsafe: "+err.Error())
		}
	}
	for _, object := range graph.releases {
		validateReleaseTraceability(object, graph, require, add)
	}
	return issues
}

func validateTraceStorage(root string) []validationIssue {
	issues := []validationIssue{}
	seen := map[string]string{}
	check := func(path, id, expected string) {
		if id == "" {
			return
		}
		displayPath := relativeDisplay(root, path)
		if previous, duplicate := seen[id]; duplicate {
			issues = append(issues, validationIssue{Path: displayPath, Schema: "semantic-links", Message: "duplicate object id also stored at: " + previous})
		} else {
			seen[id] = displayPath
		}
		if filepath.Clean(path) != filepath.Clean(expected) {
			issues = append(issues, validationIssue{Path: displayPath, Schema: "semantic-links", Message: "object is not stored at its canonical path for id: " + id})
		}
	}
	for _, directory := range []string{"goals", "requirements", "plans", "decisions", "work-items", "tests", "evidence", "releases"} {
		files, _ := listJSONFiles(filepath.Join(root, ".ai-flow", directory))
		for _, path := range files {
			var identity struct {
				ID string `json:"id"`
			}
			if readSemanticJSON(path, &identity) == nil {
				check(path, identity.ID, filepath.Join(root, ".ai-flow", directory, identity.ID+".json"))
			}
		}
	}
	runFiles, _ := filepath.Glob(filepath.Join(root, ".ai-flow", "runs", "RUN-*", "run.json"))
	for _, path := range runFiles {
		var identity struct {
			ID string `json:"id"`
		}
		if readSemanticJSON(path, &identity) == nil {
			check(path, identity.ID, runPath(root, identity.ID))
		}
	}
	checkpointFiles, _ := filepath.Glob(filepath.Join(root, ".ai-flow", "runs", "RUN-*", "checkpoints", "CP-*.json"))
	for _, path := range checkpointFiles {
		var identity struct {
			ID    string `json:"id"`
			RunID string `json:"run_id"`
		}
		if readSemanticJSON(path, &identity) == nil {
			check(path, identity.ID, checkpointPath(root, identity.RunID, identity.ID))
		}
	}
	return issues
}

func validateReplacementLinks[T interface {
	traceGoal | traceRequirement | traceDecision
}](path, id string, supersedes []string, supersededBy *string, objects map[string]traceObject[T], add func(string, string)) {
	for _, previousID := range supersedes {
		previous, found := objects[previousID]
		if !found {
			add(path, "missing replaced record: "+previousID)
			continue
		}
		backlink := replacementBacklink(previous.Value)
		if backlink == nil || *backlink != id {
			add(path, "replaced record does not point back to its replacement: "+previousID)
		}
	}
	if supersededBy == nil {
		return
	}
	next, found := objects[*supersededBy]
	if !found {
		add(path, "missing replacement record: "+*supersededBy)
		return
	}
	if !contains(replacementSupersedes(next.Value), id) {
		add(path, "replacement record does not list the record it replaces: "+*supersededBy)
	}
}

func replacementBacklink[T interface {
	traceGoal | traceRequirement | traceDecision
}](value T) *string {
	switch item := any(value).(type) {
	case traceGoal:
		return item.SupersededBy
	case traceRequirement:
		return item.SupersededBy
	case traceDecision:
		return item.SupersededBy
	}
	return nil
}

func replacementSupersedes[T interface {
	traceGoal | traceRequirement | traceDecision
}](value T) []string {
	switch item := any(value).(type) {
	case traceGoal:
		return item.Supersedes
	case traceRequirement:
		return item.Supersedes
	case traceDecision:
		return item.Supersedes
	}
	return nil
}

func validateReleaseTraceability(object traceObject[traceRelease], graph traceGraph, require func(string, string, string, bool), add func(string, string)) {
	includedWork := map[string]bool{}
	for _, workID := range object.Value.WorkItemIDs {
		work, found := graph.workItems[workID]
		require(object.Path, "development task", workID, found)
		includedWork[workID] = true
		if found && (object.Value.Status == "ready" || object.Value.Status == "released") && work.Value.Status != "done" {
			add(object.Path, "release includes unfinished development task: "+workID)
		}
	}
	for _, evidenceID := range object.Value.EvidenceIDs {
		evidence, found := graph.evidence[evidenceID]
		require(object.Path, "verification record", evidenceID, found)
		if found && !includedWork[evidence.Value.WorkItemID] {
			add(object.Path, "release verification record belongs to a task outside the release: "+evidenceID)
		}
	}
	if object.Value.Status != "ready" && object.Value.Status != "released" {
		return
	}
	for workID := range includedWork {
		work, found := graph.workItems[workID]
		if !found {
			continue
		}
		for _, requirementID := range work.Value.RequirementIDs {
			requirement, requirementFound := graph.requirements[requirementID]
			if !requirementFound || len(requirement.Value.TestIDs) == 0 {
				add(object.Path, "release task has a requirement without a linked test: "+requirementID)
				continue
			}
			covered := false
			for _, testID := range requirement.Value.TestIDs {
				test, testFound := graph.tests[testID]
				if !testFound || test.Value.WorkItemID != workID || test.Value.Status == "retired" {
					continue
				}
				for _, evidenceID := range test.Value.EvidenceIDs {
					evidence, evidenceFound := graph.evidence[evidenceID]
					if evidenceFound && contains(object.Value.EvidenceIDs, evidenceID) && evidence.Value.Result == "passed" && contains([]string{"verified-local", "verified-ci"}, evidence.Value.Trust) {
						covered = true
					}
				}
			}
			if !covered {
				add(object.Path, "release requirement lacks trusted passing verification: "+requirementID)
			}
		}
	}
}

func validateEvidenceLogPath(root, logPath string) error {
	normalized := filepath.ToSlash(strings.TrimSpace(logPath))
	if !strings.HasPrefix(normalized, ".ai-flow/evidence/logs/") {
		return fmt.Errorf("path must be inside the verification log directory")
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == ".." {
			return fmt.Errorf("path traversal is not allowed")
		}
	}
	if err := ensurePathInsideRepository(root, normalized); err != nil {
		return err
	}
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(normalized)))
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("path is not a regular file")
	}
	return nil
}

func loadTraceGraph(root string) traceGraph {
	graph := traceGraph{
		goals: map[string]traceObject[traceGoal]{}, requirements: map[string]traceObject[traceRequirement]{},
		plans: map[string]traceObject[tracePlan]{}, decisions: map[string]traceObject[traceDecision]{},
		workItems: map[string]traceObject[WorkItem]{}, tests: map[string]traceObject[traceTest]{},
		runs: map[string]traceObject[HarnessRun]{}, checkpoints: map[string]traceObject[Checkpoint]{},
		evidence: map[string]traceObject[Evidence]{}, releases: map[string]traceObject[traceRelease]{},
	}
	loadTraceDirectory(root, "goals", graph.goals)
	loadTraceDirectory(root, "requirements", graph.requirements)
	loadTraceDirectory(root, "plans", graph.plans)
	loadTraceDirectory(root, "decisions", graph.decisions)
	loadTraceDirectory(root, "work-items", graph.workItems)
	loadTraceDirectory(root, "tests", graph.tests)
	loadTraceDirectory(root, "evidence", graph.evidence)
	loadTraceDirectory(root, "releases", graph.releases)
	runFiles, _ := filepath.Glob(filepath.Join(root, ".ai-flow", "runs", "RUN-*", "run.json"))
	for _, path := range runFiles {
		var value HarnessRun
		if readSemanticJSON(path, &value) == nil {
			graph.runs[value.ID] = traceObject[HarnessRun]{Path: path, Value: value}
		}
	}
	checkpointFiles, _ := filepath.Glob(filepath.Join(root, ".ai-flow", "runs", "RUN-*", "checkpoints", "CP-*.json"))
	for _, path := range checkpointFiles {
		var value Checkpoint
		if readSemanticJSON(path, &value) == nil {
			graph.checkpoints[value.ID] = traceObject[Checkpoint]{Path: path, Value: value}
		}
	}
	return graph
}

func loadTraceDirectory[T any](root, directory string, target map[string]traceObject[T]) {
	files, _ := listJSONFiles(filepath.Join(root, ".ai-flow", directory))
	for _, path := range files {
		var value T
		if readSemanticJSON(path, &value) != nil {
			continue
		}
		var identity struct {
			ID string `json:"id"`
		}
		if readSemanticJSON(path, &identity) == nil && identity.ID != "" {
			target[identity.ID] = traceObject[T]{Path: path, Value: value}
		}
	}
}
