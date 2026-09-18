package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Execution contracts travel with the task, independent of an IDE or provider.
// Identities are declared workflow roles, not authenticated model identities.
type ExecutionSpec struct {
	Outcome        string          `json:"outcome"`
	Interfaces     []string        `json:"interfaces"`
	Invariants     []string        `json:"invariants"`
	Steps          []string        `json:"steps"`
	EdgeCases      []string        `json:"edge_cases"`
	Tests          []ExecutionTest `json:"tests"`
	EntryPoints    []string        `json:"entry_points"`
	Sources        []string        `json:"sources"`
	Facts          []string        `json:"facts"`
	StopConditions []string        `json:"stop_conditions"`
	Capabilities   []string        `json:"capabilities"`
}
type ExecutionTest struct {
	ID       string   `json:"id"`
	Command  []string `json:"command"`
	Expected string   `json:"expected"`
}
type SourceSnapshot struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type Escalation struct {
	PreviousStatus string   `json:"previous_status"`
	Reason         string   `json:"reason"`
	Evidence       []string `json:"evidence"`
	Resolution     string   `json:"resolution,omitempty"`
	ResolvedBy     string   `json:"resolved_by,omitempty"`
}
type ExecutionContract struct {
	FactHashes     map[string]string `json:"fact_hashes"`
	Revision       int               `json:"revision"`
	Planner        string            `json:"planner"`
	Executor       string            `json:"executor"`
	Reviewer       string            `json:"reviewer"`
	ExecutorModel  string            `json:"executor_model"`
	Spec           ExecutionSpec     `json:"spec"`
	PlanningSHA256 string            `json:"planning_sha256"`
	Sources        []SourceSnapshot  `json:"sources"`
	Escalations    []Escalation      `json:"escalations"`
	PreparedAt     string            `json:"prepared_at"`
}

func runHandoff(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: handoff prepare|check|escalate|resolve --work ID")
	}
	fs := flag.NewFlagSet("handoff "+args[0], flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	id := fs.String("work", "", "task ID")
	file := fs.String("file", "", "execution specification JSON")
	actor := fs.String("by", "", "declared actor identity")
	executor := fs.String("executor", "", "assigned implementation identity")
	reviewer := fs.String("reviewer", "", "assigned independent reviewer identity")
	model := fs.String("model", "", "executor model label; not a qualification claim")
	reason := fs.String("reason", "", "escalation or resolution with rationale")
	var evidence stringListFlag
	fs.Var(&evidence, "evidence", "task evidence ID or repository file; repeatable")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	*actor = strings.TrimSpace(*actor)
	*executor = strings.TrimSpace(*executor)
	*reviewer = strings.TrimSpace(*reviewer)
	*model = strings.TrimSpace(*model)
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	if args[0] == "check" {
		w, e := readWorkItem(root, *id)
		if e != nil {
			return e
		}
		return checkExecution(root, w, true)
	}
	if !contains([]string{"prepare", "escalate", "resolve"}, args[0]) {
		return errors.New("unknown handoff action")
	}
	if err := requireProjectCompatible(root); err != nil {
		return err
	}
	unlock, err := projectMutationLock(root)
	if err != nil {
		return err
	}
	defer unlock()
	w, err := readWorkItem(root, *id)
	if err != nil {
		return err
	}
	if contains([]string{"done", "cancelled", "accepted", "closing", "awaiting_acceptance", "ready_for_review"}, w.Status) {
		return errors.New("reopen this task before revising its execution contract")
	}
	if strings.TrimSpace(*actor) == "" {
		return errors.New("--by is required")
	}
	switch args[0] {
	case "prepare":
		var spec ExecutionSpec
		data, e := os.ReadFile(*file)
		if e != nil {
			return e
		}
		decoder := json.NewDecoder(strings.NewReader(string(data)))
		decoder.DisallowUnknownFields()
		if e = decoder.Decode(&spec); e != nil {
			return e
		}
		var trailing any
		if e = decoder.Decode(&trailing); e != io.EOF {
			return errors.New("execution input must contain exactly one JSON object")
		}
		if e = validateExecutionSpec(spec); e != nil {
			return e
		}
		if *executor == "" || *reviewer == "" || *actor == *executor || *reviewer == *executor {
			return errors.New("assign a planner, executor and independent reviewer; executor must differ from both")
		}
		if *model == "" {
			if w.Execution != nil {
				*model = w.Execution.ExecutorModel
			} else {
				*model = "unspecified"
			}
		}
		rev := 1
		previous := []Escalation{}
		if w.Execution != nil {
			if w.Execution.Planner != *actor {
				return errors.New("only the recorded planner can revise the contract")
			}
			if *reason == "" {
				return errors.New("record why the execution contract changes")
			}
			rev = w.Execution.Revision + 1
			previous = w.Execution.Escalations
			for _, e := range previous {
				if e.Resolution == "" {
					return errors.New("resolve the outstanding escalation before revising the contract")
				}
			}
		}
		// Required checks cannot disappear when the task moves between models.
		for _, t := range spec.Tests {
			w.RequiredTests = uniqueAppend(w.RequiredTests, t.ID)
		}
		for _, id := range w.RequiredTests {
			found := false
			for _, t := range spec.Tests {
				if t.ID == id {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("missing required test specification: %s", id)
			}
		}
		if e := checkProfileConfirmation(root); e != nil {
			return e
		}
		sources := append([]string(nil), spec.Sources...)
		sources = uniqueAppend(sources, ".ai-flow/baseline/engineering-profile.json")
		sources = uniqueAppend(sources, ".ai-flow/baseline/profile-confirmation.json")
		decisions, e := executionDecisionSources(root, w)
		if e != nil {
			return e
		}
		for _, p := range decisions {
			sources = uniqueAppend(sources, p)
		}
		if w.GoalID != nil {
			sources = uniqueAppend(sources, ".ai-flow/goals/"+*w.GoalID+".json")
		}
		for _, id := range w.RequirementIDs {
			sources = uniqueAppend(sources, ".ai-flow/requirements/"+id+".json")
		}
		snapshots, e := snapshotSources(root, sources)
		if e != nil {
			return e
		}
		for _, p := range spec.EntryPoints {
			if _, e := readProjectSource(root, p); e != nil {
				return e
			}
		}
		factHashes := map[string]string{}
		for _, key := range spec.Facts {
			fact, e := verifiedFact(root, key)
			if e != nil {
				return e
			}
			digest, e := hashJSON(fact)
			if e != nil {
				return e
			}
			factHashes[key] = digest
		}
		digest, e := planningFingerprint(w)
		if e != nil {
			return e
		}
		if w.Execution != nil {
			if e := writeJSONAtomic(filepath.Join(root, ".ai-flow/reports", w.ID, "execution-history", fmt.Sprintf("revision-%d.json", w.Execution.Revision)), w.Execution); e != nil {
				return e
			}
		}
		w.Execution = &ExecutionContract{FactHashes: factHashes, Revision: rev, Planner: *actor, Executor: *executor, Reviewer: *reviewer, ExecutorModel: *model, Spec: spec, PlanningSHA256: digest, Sources: snapshots, Escalations: previous, PreparedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	case "escalate":
		if w.Execution == nil {
			return errors.New("prepare the execution contract first")
		}
		if *actor != w.Execution.Executor {
			return errors.New("escalation must name the assigned executor")
		}
		if strings.TrimSpace(*reason) == "" || len(evidence) == 0 {
			return errors.New("provide the concrete conflict and observed evidence")
		}
		for _, p := range evidence {
			if strings.HasPrefix(p, "EV-") {
				e, err := readEvidence(root, p)
				if err != nil {
					return err
				}
				if e.WorkItemID != w.ID {
					return errors.New("evidence belongs to another task")
				}
			} else if _, err := readProjectSource(root, p); err != nil {
				return err
			}
		}
		w.Execution.Escalations = append(w.Execution.Escalations, Escalation{Reason: *reason, Evidence: evidence, PreviousStatus: w.Status})
		w.Status = "blocked"
		w.BlockedReason = pointer(*reason)
	case "resolve":
		if w.Execution == nil || *actor != w.Execution.Planner {
			return errors.New("only the recorded planner can resolve an escalation")
		}
		if strings.TrimSpace(*reason) == "" {
			return errors.New("record the decision, rationale and next action")
		}
		found := false
		resumeStatus := ""
		for i := range w.Execution.Escalations {
			e := &w.Execution.Escalations[i]
			if e.Resolution == "" {
				if resumeStatus == "" && e.PreviousStatus != "blocked" {
					resumeStatus = e.PreviousStatus
				}
				e.Resolution = *reason
				e.ResolvedBy = *actor
				found = true
			}
		}
		if !found {
			return errors.New("no outstanding escalation")
		}
		if resumeStatus == "" {
			if w.RunID == nil {
				resumeStatus = "ready"
			} else {
				resumeStatus = "in_progress"
			}
		}
		w.Status = resumeStatus
		w.BlockedReason = nil
	}
	return saveWorkMutation(root, &w, "work.handoff_"+args[0], map[string]any{"by": *actor, "reason": *reason})
}

func validateExecutionSpec(s ExecutionSpec) error {
	if strings.TrimSpace(s.Outcome) == "" {
		return errors.New("execution outcome is required")
	}
	for name, list := range map[string][]string{"interfaces": s.Interfaces, "invariants": s.Invariants, "steps": s.Steps, "edge_cases": s.EdgeCases, "entry_points": s.EntryPoints, "stop_conditions": s.StopConditions, "capabilities": s.Capabilities} {
		if len(list) == 0 {
			return fmt.Errorf("execution specification needs %s", name)
		}
		for _, v := range list {
			if strings.TrimSpace(v) == "" {
				return fmt.Errorf("blank %s entry", name)
			}
		}
	}
	if len(s.Tests) == 0 {
		return errors.New("specify executable tests")
	}
	seen := map[string]bool{}
	for _, t := range s.Tests {
		if t.ID == "" || seen[t.ID] || len(t.Command) == 0 || strings.TrimSpace(t.Command[0]) == "" || strings.TrimSpace(t.Expected) == "" {
			return errors.New("tests require unique IDs, command argument arrays and observable expected results")
		}
		seen[t.ID] = true
	}
	return nil
}

func planningFingerprint(w WorkItem) (string, error) {
	return hashJSON(map[string]any{"title": w.Title, "scope": w.Scope, "protected_paths": w.ProtectedPaths, "requirements": w.RequirementIDs, "acceptance": w.AcceptanceCriteria, "dependencies": w.Dependencies, "tests": w.RequiredTests})
}

func checkExecution(root string, w WorkItem, required bool) error {
	c := w.Execution
	if c == nil {
		if required {
			return errors.New("prepare a task execution contract before handing implementation to another model")
		}
		return nil
	}
	if err := checkProfileConfirmation(root); err != nil {
		return err
	}
	if err := validateExecutionSpec(c.Spec); err != nil {
		return err
	}
	for _, e := range c.Escalations {
		if e.Resolution == "" {
			return errors.New("execution paused: planner must resolve the recorded conflict before continuing")
		}
	}
	digest, err := planningFingerprint(w)
	if err != nil {
		return err
	}
	if digest != c.PlanningSHA256 {
		return errors.New("task scope or acceptance changed; planner must refresh the execution contract")
	}
	if err := checkSnapshots(root, c.Sources); err != nil {
		return err
	}
	for _, key := range c.Spec.Facts {
		f, err := verifiedFact(root, key)
		if err != nil {
			return err
		}
		digest, err := hashJSON(f)
		if err != nil {
			return err
		}
		if c.FactHashes[key] != digest {
			return errors.New("a planning fact was revised; refresh the execution contract")
		}
	}
	decisions, err := executionDecisionSources(root, w)
	if err != nil {
		return err
	}
	for _, p := range decisions {
		found := false
		for _, s := range c.Sources {
			if s.Path == p {
				found = true
			}
		}
		if !found {
			return errors.New("a new design decision needs inclusion in the execution contract")
		}
	}
	return nil
}

func executionOwner(w WorkItem, owner string) error {
	if w.Execution != nil && owner != w.Execution.Executor {
		return errors.New("use the assigned executor identity; changing IDE does not change its role")
	}
	return nil
}
