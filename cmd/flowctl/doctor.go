package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type doctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	jsonOutput := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, false)
	if err != nil {
		return err
	}

	checks := []doctorCheck{}
	add := func(name string, ok bool, warning bool, message string) {
		status := "ok"
		if !ok && warning {
			status = "warning"
		} else if !ok {
			status = "error"
		}
		checks = append(checks, doctorCheck{Name: name, Status: status, Message: message})
	}

	flowctlPath := filepath.Join(root, ".ai-flow", "bin", executableName("flowctl"))
	_, flowctlErr := os.Stat(flowctlPath)
	add("flowctl", flowctlErr == nil, false, flowctlPath)

	platforms := installedPlatforms(root)
	add("platforms", len(platforms) > 0, false, strings.Join(platforms, ", "))
	entries := []string{}
	for _, platform := range platforms {
		var checkName, skillRoot string
		switch platform {
		case "codex":
			checkName = "codex-skills"
			skillRoot = filepath.Join(".agents", "skills")
			entries = append(entries, filepath.Join(root, "AGENTS.md"))
		case "cursor":
			checkName = "cursor-skills"
			skillRoot = filepath.Join(".cursor", "skills")
			entries = append(entries, filepath.Join(root, ".cursor", "rules", "ai-flow.mdc"))
		case "claude":
			checkName = "claude-skills"
			skillRoot = filepath.Join(".claude", "skills")
			entries = append(entries, filepath.Join(root, "CLAUDE.md"), filepath.Join(root, ".claude", "skills", "ai-flow", "SKILL.md"))
		}
		missing := missingCoreSkills(root, skillRoot)
		add(checkName, len(missing) == 0, false, missingMessage(missing))
	}
	entryMissing := []string{}
	for _, path := range entries {
		if _, err := os.Stat(path); err != nil {
			entryMissing = append(entryMissing, relativeDisplay(root, path))
		}
	}
	add("platform-entries", len(entryMissing) == 0, false, missingMessage(entryMissing))

	schemaFiles, _ := filepath.Glob(filepath.Join(root, ".ai-flow", "runtime", "schemas", "*.schema.json"))
	add("json-schemas", len(schemaFiles) >= 15, false, fmt.Sprintf("%d schema files installed", len(schemaFiles)))

	_, manifestErr := os.Stat(filepath.Join(root, ".ai-flow", "manifest.yaml"))
	projectMessage := "project state initialized"
	if manifestErr != nil {
		projectMessage = "run initialize-ai-project when not initialized"
	}
	add("project-state", manifestErr == nil, true, projectMessage)

	if *jsonOutput {
		out, _ := json.MarshalIndent(map[string]any{"root": root, "version": packVersion, "checks": checks}, "", "  ")
		fmt.Println(string(out))
	} else {
		for _, check := range checks {
			fmt.Printf("%-18s %-8s %s\n", check.Name, strings.ToUpper(check.Status), check.Message)
		}
	}

	for _, check := range checks {
		if check.Status == "error" {
			return errors.New("installation health check failed")
		}
	}
	return nil
}

func installedPlatforms(root string) []string {
	path := filepath.Join(root, ".ai-flow", "install", "platforms")
	if data, err := os.ReadFile(path); err == nil {
		selected := []string{}
		for _, value := range strings.FieldsFunc(string(data), func(r rune) bool {
			return r == ',' || r == '\n' || r == '\r' || r == ' ' || r == '\t'
		}) {
			value = strings.TrimPrefix(value, "\ufeff")
			if contains([]string{"cursor", "codex", "claude"}, value) && !contains(selected, value) {
				selected = append(selected, value)
			}
		}
		if len(selected) > 0 {
			sort.Strings(selected)
			return selected
		}
	}

	detected := []string{}
	locations := map[string]string{
		"codex":  filepath.Join(root, ".agents", "skills", "initialize-ai-project", "SKILL.md"),
		"cursor": filepath.Join(root, ".cursor", "skills", "initialize-ai-project", "SKILL.md"),
		"claude": filepath.Join(root, ".claude", "skills", "initialize-ai-project", "SKILL.md"),
	}
	for _, platform := range []string{"claude", "codex", "cursor"} {
		if _, err := os.Stat(locations[platform]); err == nil {
			detected = append(detected, platform)
		}
	}
	return detected
}

func missingCoreSkills(root, relativeRoot string) []string {
	missing := []string{}
	for _, skill := range coreSkills {
		path := filepath.Join(root, relativeRoot, skill, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, skill)
		}
	}
	sort.Strings(missing)
	return missing
}
