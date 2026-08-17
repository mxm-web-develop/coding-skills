package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type solutionDecision struct {
	SchemaVersion        int                           `json:"schema_version"`
	DecisionType         string                        `json:"decision_type"`
	Status               string                        `json:"status"`
	Options              []solutionDecisionOption      `json:"options"`
	RecommendedOption    string                        `json:"recommended_option"`
	RecommendationReason string                        `json:"recommendation_reason"`
	Confirmation         *solutionDecisionConfirmation `json:"confirmation"`
}

type solutionDecisionOption struct {
	Name          string  `json:"name"`
	PrototypePath *string `json:"prototype_path"`
}

type solutionDecisionConfirmation struct {
	Status         string  `json:"status"`
	SelectedOption *string `json:"selected_option"`
}

func validateSolutionDecisions(root string) []validationIssue {
	issues := []validationIssue{}
	decisionFiles, _ := listJSONFiles(filepath.Join(root, ".ai-flow", "decisions"))
	for _, path := range decisionFiles {
		var decision solutionDecision
		if readJSON(path, &decision) != nil {
			continue
		}
		if strings.TrimSpace(decision.DecisionType) == "" {
			continue
		}
		add := func(message string) {
			issues = append(issues, validationIssue{Path: relativeDisplay(root, path), Schema: "solution-confirmation", Message: message})
		}
		if len(decision.Options) < 2 {
			add("interactive solution decision must contain at least two viable options")
		}
		optionNames := map[string]bool{}
		for _, option := range decision.Options {
			name := strings.TrimSpace(option.Name)
			if optionNames[name] {
				add("interactive solution decision contains a duplicate option: " + name)
			}
			optionNames[name] = true
			if option.PrototypePath != nil {
				if err := validatePrototypePath(root, *option.PrototypePath); err != nil {
					add(fmt.Sprintf("prototype for option %s is unavailable or unsafe: %v", name, err))
				}
			}
		}
		recommended := strings.TrimSpace(decision.RecommendedOption)
		if recommended == "" || !optionNames[recommended] {
			add("recommended option must name one of the compared options")
		}
		if strings.TrimSpace(decision.RecommendationReason) == "" {
			add("interactive solution decision must explain the recommendation")
		}
		prototypeCount := 0
		prototypePaths := map[string]bool{}
		if decision.Confirmation == nil {
			add("interactive solution decision must record whether the user has confirmed a direction")
			continue
		}
		selected := ""
		if decision.Confirmation.SelectedOption != nil {
			selected = strings.TrimSpace(*decision.Confirmation.SelectedOption)
			if selected != "" && !optionNames[selected] {
				add("confirmed option must name one of the compared options")
			}
		}
		if decision.Status == "accepted" && (decision.Confirmation.Status != "confirmed" || selected == "") {
			add("accepted interactive solution decision requires a confirmed option")
		}
		if decision.DecisionType == "frontend-ux-ui" {
			if len(decision.Options) < 2 || len(decision.Options) > 3 {
				add("frontend UX decision should compare two or three HTML directions")
			}
			for _, option := range decision.Options {
				if option.PrototypePath == nil || strings.TrimSpace(*option.PrototypePath) == "" {
					add("frontend UX decision must keep every compared direction in an HTML prototype")
					continue
				}
				prototypeCount++
				normalized := filepath.ToSlash(strings.TrimSpace(*option.PrototypePath))
				if prototypePaths[normalized] {
					add("frontend UX decision must use distinct prototype files")
					continue
				}
				prototypePaths[normalized] = true
				if err := validatePrototypePath(root, *option.PrototypePath); err != nil {
					add(fmt.Sprintf("prototype for option %s is unavailable or unsafe: %v", option.Name, err))
					continue
				}
				if err := validatePrototypeExperience(root, *option.PrototypePath); err != nil {
					add(fmt.Sprintf("prototype for option %s does not show responsive, interactive HTML evidence: %v", option.Name, err))
				}
			}
			if prototypeCount < 2 {
				add("frontend UX decision must include at least two distinct prototype directions")
			}
		}
	}
	return issues
}

func validatePrototypePath(root, prototypePath string) error {
	normalized := filepath.ToSlash(strings.TrimSpace(prototypePath))
	if (!strings.HasPrefix(normalized, ".ai-flow/prototypes/") && !strings.HasPrefix(normalized, ".ai-flow/archive/design-explorations/")) ||
		!strings.HasSuffix(strings.ToLower(normalized), ".html") {
		return fmt.Errorf("path must be an HTML file in the managed exploration directories")
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

func validatePrototypeExperience(root, prototypePath string) error {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(strings.TrimSpace(prototypePath))))
	if err != nil {
		return err
	}
	normalized := strings.ToLower(string(data))
	if !strings.Contains(normalized, "meta name=\"viewport\"") {
		return fmt.Errorf("missing responsive viewport metadata")
	}
	if !strings.Contains(normalized, "<button") && !strings.Contains(normalized, "href=") && !strings.Contains(normalized, "role=\"button\"") {
		return fmt.Errorf("missing clickable interaction evidence")
	}
	if !strings.Contains(normalized, "@media") && !strings.Contains(normalized, "prefers-reduced-motion") && !strings.Contains(normalized, "transition") && !strings.Contains(normalized, "animation") {
		return fmt.Errorf("missing responsive or motion evidence")
	}
	return nil
}
