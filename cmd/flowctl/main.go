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
	"time"
)

const packName = "mxm-ai-flow"

var packVersion = "0.2.1"

var coreSkills = []string{
	"initialize-ai-project",
	"orchestrate-ai-delivery",
	"adopt-existing-project",
	"discover-product-goal",
	"plan-product-delivery",
	"profile-project-engineering",
	"research-and-design-solution",
	"specify-tests",
	"implement-work-item",
	"diagnose-and-verify",
	"review-change",
	"integrate-git-change",
	"manage-release",
	"sync-project-knowledge",
}

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

type doctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Printf("flowctl %s\n", packVersion)
		return
	case "doctor":
		err = runDoctor(os.Args[2:])
	case "status":
		err = runStatus(os.Args[2:])
	case "work":
		err = runWork(os.Args[2:])
	case "checkpoint":
		err = runCheckpoint(os.Args[2:])
	case "evidence":
		err = runEvidence(os.Args[2:])
	case "validate":
		err = runValidate(os.Args[2:])
	case "project":
		if len(os.Args) < 3 || os.Args[2] != "init" {
			err = errors.New("usage: flowctl project init [--root PATH] --mode greenfield|existing --name NAME")
		} else {
			err = runProjectInit(os.Args[3:])
		}
	case "render-board":
		err = runRenderBoard(os.Args[2:])
	case "help", "--help", "-h":
		usage()
		return
	default:
		err = fmt.Errorf("unknown command %q", os.Args[1])
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "flowctl: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`flowctl manages the deterministic state of an AI Flow project.

Usage:
  flowctl version
  flowctl doctor [--root PATH] [--json]
  flowctl project init [--root PATH] --mode greenfield|existing --name NAME
  flowctl status [--root PATH] [--json]
  flowctl work <create|list|show|ready|start|block|review-ready|complete|cancel>
  flowctl checkpoint <save|list|show|latest|resume>
  flowctl evidence <run|record|list|show|verify>
  flowctl validate [--root PATH] [--json]
  flowctl render-board [--root PATH]`)
}

func runProjectInit(args []string) error {
	fs := flag.NewFlagSet("project init", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	mode := fs.String("mode", "", "greenfield or existing")
	name := fs.String("name", "", "project name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *mode != "greenfield" && *mode != "existing" {
		return errors.New("--mode must be greenfield or existing")
	}
	if strings.TrimSpace(*name) == "" {
		return errors.New("--name is required")
	}

	root, err := resolveRoot(*rootArg, false)
	if err != nil {
		return err
	}
	if err := ensureInstalled(root); err != nil {
		return err
	}

	dirs := []string{
		"state", "events", "baseline", "goals", "requirements", "plans",
		"work-items", "decisions", "tests", "evidence", "runs", "reports",
		"releases", "locks", "archive",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(root, ".ai-flow", dir), 0o755); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "docs", "board"), 0o755); err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	phase := "goal_alignment"
	if *mode == "existing" {
		phase = "baselining"
	}

	files := map[string]string{
		".ai-flow/manifest.yaml":        fmt.Sprintf("schema_version: 1\npack_name: %s\npack_version: %s\ninitialized_at: %s\n", packName, packVersion, now),
		".ai-flow/project.yaml":         fmt.Sprintf("schema_version: 1\nname: %s\nmode: %s\nprofile: core\nversion_policy: semver\ncurrent_version: v0.0.0\n", yamlScalar(*name), *mode),
		".ai-flow/state/current.yaml":   fmt.Sprintf("schema_version: 1\nrevision: 1\nphase: %s\nstatus: active\ncurrent_version: v0.0.0\nactive_goal: none\nnext_action: %s\ntests: not_run\nupdated_at: %s\n", phase, nextActionForMode(*mode), now),
		".ai-flow/skill-pack.lock.yaml": fmt.Sprintf("schema_version: 1\nname: %s\nversion: %s\nsource: installed\n", packName, packVersion),
	}

	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := writeIfMissing(path, content, 0o644); err != nil {
			return err
		}
	}

	if err := renderBoard(root); err != nil {
		return err
	}
	fmt.Printf("Initialized %s AI Flow project %q at %s\n", *mode, *name, root)
	return nil
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

func runRenderBoard(args []string) error {
	fs := flag.NewFlagSet("render-board", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	if err := renderBoard(root); err != nil {
		return err
	}
	fmt.Printf("Rendered human board at %s\n", filepath.Join(root, "docs", "board"))
	return nil
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

func renderBoard(root string) error {
	status, err := readStatus(root)
	if err != nil {
		return err
	}
	if !status.Initialized {
		return errors.New("project is not initialized")
	}
	boardDir := filepath.Join(root, "docs", "board")
	if err := os.MkdirAll(boardDir, 0o755); err != nil {
		return err
	}

	statusDoc := fmt.Sprintf("# Project Status\n\n- Project: %s\n- Version: %s\n- Phase: %s\n- Status: %s\n- Active goal: %s\n- Tests: %s\n- Next action: %s\n- State revision: %s\n- Updated: %s\n\n## Work items\n\n| Draft | Ready | In progress | Review | Blocked | Done | Cancelled |\n| ---: | ---: | ---: | ---: | ---: | ---: | ---: |\n| %d | %d | %d | %d | %d | %d | %d |\n\n## Evidence\n\n| Passed | Failed | Unverified |\n| ---: | ---: | ---: |\n| %d | %d | %d |\n\n> Generated from `.ai-flow/`. Do not edit project facts here.\n", status.ProjectName, status.CurrentVersion, status.Phase, status.Status, status.ActiveGoal, status.Tests, status.NextAction, status.Revision, status.UpdatedAt, status.WorkDraft, status.WorkReady, status.WorkInProgress, status.WorkReview, status.WorkBlocked, status.WorkDone, status.WorkCancelled, status.EvidencePassed, status.EvidenceFailed, status.EvidenceOther)
	roadmap := fmt.Sprintf("# Roadmap\n\n## Active goal\n\n%s\n\n## Next milestone\n\n%s\n\n> Detailed plans live in `.ai-flow/plans/`.\n", status.ActiveGoal, status.NextAction)
	current := fmt.Sprintf("# Current State\n\n- Mode: %s\n- Profile: %s\n- Workflow phase: %s\n- Current version: %s\n\nCurrent capabilities and accepted decisions are indexed by `.ai-flow/state/current.yaml`.\n", status.Mode, status.Profile, status.Phase, status.CurrentVersion)
	releases := "# Releases\n\nNo observed releases have been recorded. Machine release records live in `.ai-flow/releases/`.\n"

	files := map[string]string{
		"STATUS.md":        statusDoc,
		"ROADMAP.md":       roadmap,
		"CURRENT_STATE.md": current,
		"RELEASES.md":      releases,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(boardDir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func readStatus(root string) (projectStatus, error) {
	status := projectStatus{Root: root, Initialized: false}
	manifest := filepath.Join(root, ".ai-flow", "manifest.yaml")
	if _, err := os.Stat(manifest); err != nil {
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

func readFlatYAML(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		values[strings.TrimSpace(parts[0])] = strings.Trim(strings.TrimSpace(parts[1]), "\"'")
	}
	return values, nil
}

func resolveRoot(rootArg string, requireInstall bool) (string, error) {
	start := rootArg
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = cwd
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("root is not a directory: %s", abs)
	}
	if rootArg != "" {
		if requireInstall {
			if err := ensureInstalled(abs); err != nil {
				return "", err
			}
		}
		return filepath.Clean(abs), nil
	}

	current := filepath.Clean(abs)
	for {
		if _, err := os.Stat(filepath.Join(current, ".ai-flow")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	if requireInstall {
		return "", errors.New("no .ai-flow directory found; run the installer first")
	}
	return filepath.Clean(abs), nil
}

func ensureInstalled(root string) error {
	path := filepath.Join(root, ".ai-flow", "bin", executableName("flowctl"))
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("AI Flow runtime is not installed at %s", path)
	}
	return nil
}

func writeIfMissing(path, content string, mode os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), mode)
}

func yamlScalar(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", " ", "\r", " ")
	return "\"" + replacer.Replace(strings.TrimSpace(value)) + "\""
}

func nextActionForMode(mode string) string {
	if mode == "existing" {
		return "adopt_existing_project"
	}
	return "discover_product_goal"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func executableName(name string) string {
	if filepath.Separator == '\\' {
		return name + ".exe"
	}
	return name
}

func missingMessage(items []string) string {
	if len(items) == 0 {
		return "all required files are present"
	}
	return "missing: " + strings.Join(items, ", ")
}

func relativeDisplay(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
