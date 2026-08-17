package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type projectStatus struct {
	Root           string `json:"root"`
	Initialized    bool   `json:"initialized"`
	ProjectName    string `json:"project_name"`
	Mode           string `json:"mode"`
	Profile        string `json:"profile"`
	CurrentVersion string `json:"current_version"`
	Phase          string `json:"phase"`
	Status         string `json:"status"`
	Revision       string `json:"revision"`
	ActiveGoal     string `json:"active_goal"`
	NextAction     string `json:"next_action"`
	Tests          string `json:"tests"`
	UpdatedAt      string `json:"updated_at"`
	WorkDraft      int    `json:"work_draft"`
	WorkReady      int    `json:"work_ready"`
	WorkInProgress int    `json:"work_in_progress"`
	WorkReview     int    `json:"work_ready_for_review"`
	WorkBlocked    int    `json:"work_blocked"`
	WorkDone       int    `json:"work_done"`
	WorkCancelled  int    `json:"work_cancelled"`
	EvidencePassed int    `json:"evidence_passed"`
	EvidenceFailed int    `json:"evidence_failed"`
	EvidenceOther  int    `json:"evidence_unverified"`
}

func runStatus(args []string) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	jsonOutput := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	status, err := readStatus(root)
	if err != nil {
		return err
	}
	if *jsonOutput {
		out, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(out))
		return nil
	}
	if !status.Initialized {
		fmt.Printf("AI Flow is installed but project state is not initialized at %s\n", root)
		return nil
	}
	fmt.Printf("Project: %s\nVersion: %s\nPhase: %s\nStatus: %s\nActive goal: %s\nTests: %s\nWork: ready=%d in_progress=%d review=%d blocked=%d done=%d\nEvidence: passed=%d failed=%d unverified=%d\nNext: %s\n", status.ProjectName, status.CurrentVersion, status.Phase, status.Status, status.ActiveGoal, status.Tests, status.WorkReady, status.WorkInProgress, status.WorkReview, status.WorkBlocked, status.WorkDone, status.EvidencePassed, status.EvidenceFailed, status.EvidenceOther, status.NextAction)
	return nil
}

func readStatus(root string) (projectStatus, error) {
	status := projectStatus{Root: root, Initialized: false}
	if _, err := os.Stat(filepath.Join(root, ".ai-flow", "manifest.yaml")); err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return status, err
	}
	project, err := readFlatYAML(filepath.Join(root, ".ai-flow", "project.yaml"))
	if err != nil {
		return status, err
	}
	state, err := readFlatYAML(filepath.Join(root, ".ai-flow", "state", "current.yaml"))
	if err != nil {
		return status, err
	}
	status.Initialized = true
	status.ProjectName = project["name"]
	status.Mode = project["mode"]
	status.Profile = project["profile"]
	status.CurrentVersion = firstNonEmpty(state["current_version"], project["current_version"], "unknown")
	status.Phase = firstNonEmpty(state["phase"], "unknown")
	status.Status = firstNonEmpty(state["status"], "unknown")
	status.Revision = firstNonEmpty(state["revision"], "unknown")
	status.ActiveGoal = firstNonEmpty(state["active_goal"], "none")
	status.NextAction = firstNonEmpty(state["next_action"], "not_recorded")
	status.Tests = firstNonEmpty(state["tests"], "unknown")
	status.UpdatedAt = firstNonEmpty(state["updated_at"], "unknown")
	if err := addObjectCounts(root, &status); err != nil {
		return status, err
	}
	return status, nil
}

func addObjectCounts(root string, status *projectStatus) error {
	workFiles, err := listJSONFiles(filepath.Join(root, ".ai-flow", "work-items"))
	if err != nil {
		return err
	}
	for _, path := range workFiles {
		var item WorkItem
		if err := readJSON(path, &item); err != nil {
			return err
		}
		switch item.Status {
		case "draft":
			status.WorkDraft++
		case "ready":
			status.WorkReady++
		case "in_progress":
			status.WorkInProgress++
		case "ready_for_review":
			status.WorkReview++
		case "blocked":
			status.WorkBlocked++
		case "done":
			status.WorkDone++
		case "cancelled":
			status.WorkCancelled++
		}
	}
	evidenceFiles, err := listJSONFiles(filepath.Join(root, ".ai-flow", "evidence"))
	if err != nil {
		return err
	}
	for _, path := range evidenceFiles {
		var evidence Evidence
		if err := readJSON(path, &evidence); err != nil {
			return err
		}
		switch evidence.Result {
		case "passed":
			status.EvidencePassed++
		case "failed":
			status.EvidenceFailed++
		default:
			status.EvidenceOther++
		}
	}
	return nil
}
