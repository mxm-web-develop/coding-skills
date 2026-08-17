package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

type validationIssue struct {
	Path    string `json:"path"`
	Schema  string `json:"schema"`
	Message string `json:"message"`
}

type validationSummary struct {
	Root      string            `json:"root"`
	Validated int               `json:"validated"`
	Valid     bool              `json:"valid"`
	Issues    []validationIssue `json:"issues"`
}

type validationTarget struct {
	Path   string
	Schema string
	Line   int
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	jsonOutput := fs.Bool("json", false, "emit JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	schemaRoot, err := findSchemaRoot(root)
	if err != nil {
		return err
	}
	compiled, err := compileSchemas(schemaRoot)
	if err != nil {
		return err
	}
	targets, err := collectValidationTargets(root)
	if err != nil {
		return err
	}
	summary := validationSummary{Root: root, Validated: 0, Valid: true, Issues: []validationIssue{}}
	for _, target := range targets {
		schema, ok := compiled[target.Schema]
		if !ok {
			return fmt.Errorf("compiled schema not found: %s", target.Schema)
		}
		values, err := readValidationValues(target)
		if err != nil {
			summary.Valid = false
			summary.Issues = append(summary.Issues, validationIssue{Path: relativeDisplay(root, target.Path), Schema: target.Schema, Message: err.Error()})
			continue
		}
		for _, value := range values {
			summary.Validated++
			if err := schema.Validate(value); err != nil {
				summary.Valid = false
				summary.Issues = append(summary.Issues, validationIssue{Path: relativeDisplay(root, target.Path), Schema: target.Schema, Message: err.Error()})
			}
		}
	}
	semanticIssues := validateSemanticLinks(root)
	if len(semanticIssues) > 0 {
		summary.Valid = false
		summary.Issues = append(summary.Issues, semanticIssues...)
	}
	if *jsonOutput {
		_ = printJSON(summary)
	} else {
		for _, issue := range summary.Issues {
			fmt.Printf("INVALID\t%s\t%s\t%s\n", issue.Path, issue.Schema, issue.Message)
		}
		fmt.Printf("validated=%d valid=%t issues=%d\n", summary.Validated, summary.Valid, len(summary.Issues))
	}
	if !summary.Valid {
		return errors.New("validation failed")
	}
	return nil
}

func findSchemaRoot(root string) (string, error) {
	candidates := []string{
		filepath.Join(root, ".ai-flow", "runtime", "schemas"),
		filepath.Join(root, "schemas"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", errors.New("no runtime schemas found; reinstall AI Flow")
}

func compileSchemas(schemaRoot string) (map[string]*jsonschema.Schema, error) {
	files, err := filepath.Glob(filepath.Join(schemaRoot, "*.schema.json"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("schema directory is empty")
	}
	sort.Strings(files)
	compiler := jsonschema.NewCompiler()
	compiler.AssertFormat()
	ids := map[string]string{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var document any
		if err := json.Unmarshal(data, &document); err != nil {
			return nil, fmt.Errorf("invalid schema JSON %s: %w", path, err)
		}
		object, ok := document.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("schema is not an object: %s", path)
		}
		id, ok := object["$id"].(string)
		if !ok || id == "" {
			id = fileURL(path)
		}
		if err := compiler.AddResource(id, document); err != nil {
			return nil, err
		}
		ids[filepath.Base(path)] = id
	}
	compiled := map[string]*jsonschema.Schema{}
	for filename, id := range ids {
		if filename == "common.schema.json" {
			continue
		}
		schema, err := compiler.Compile(id)
		if err != nil {
			return nil, fmt.Errorf("compile %s: %w", filename, err)
		}
		compiled[filename] = schema
	}
	return compiled, nil
}

func collectValidationTargets(root string) ([]validationTarget, error) {
	targets := []validationTarget{}
	engineeringProfile := filepath.Join(root, ".ai-flow", "baseline", "engineering-profile.json")
	if info, err := os.Stat(engineeringProfile); err == nil && !info.IsDir() {
		targets = append(targets, validationTarget{Path: engineeringProfile, Schema: "engineering-profile.schema.json"})
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	flat := map[string]string{
		"goals":        "goal.schema.json",
		"requirements": "requirement.schema.json",
		"plans":        "plan.schema.json",
		"work-items":   "work-item.schema.json",
		"decisions":    "decision.schema.json",
		"tests":        "test-spec.schema.json",
		"evidence":     "evidence.schema.json",
		"locks":        "lock.schema.json",
		"releases":     "release.schema.json",
	}
	for dir, schema := range flat {
		files, err := listJSONFiles(filepath.Join(root, ".ai-flow", dir))
		if err != nil {
			return nil, err
		}
		for _, path := range files {
			targets = append(targets, validationTarget{Path: path, Schema: schema})
		}
	}
	runFiles, err := filepath.Glob(filepath.Join(root, ".ai-flow", "runs", "RUN-*", "run.json"))
	if err != nil {
		return nil, err
	}
	for _, path := range runFiles {
		targets = append(targets, validationTarget{Path: path, Schema: "run.schema.json"})
	}
	checkpointFiles, err := filepath.Glob(filepath.Join(root, ".ai-flow", "runs", "RUN-*", "checkpoints", "CP-*.json"))
	if err != nil {
		return nil, err
	}
	for _, path := range checkpointFiles {
		targets = append(targets, validationTarget{Path: path, Schema: "checkpoint.schema.json"})
	}
	eventFiles, err := filepath.Glob(filepath.Join(root, ".ai-flow", "events", "*.jsonl"))
	if err != nil {
		return nil, err
	}
	for _, path := range eventFiles {
		targets = append(targets, validationTarget{Path: path, Schema: "event.schema.json", Line: -1})
	}
	sort.Slice(targets, func(i, j int) bool { return targets[i].Path < targets[j].Path })
	return targets, nil
}

func readValidationValues(target validationTarget) ([]any, error) {
	if target.Line != -1 {
		data, err := os.ReadFile(target.Path)
		if err != nil {
			return nil, err
		}
		var value any
		decoder := json.NewDecoder(strings.NewReader(string(data)))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		return []any{value}, nil
	}
	file, err := os.Open(target.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	values := []any{}
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var value any
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		values = append(values, value)
	}
	return values, scanner.Err()
}

func validateSemanticLinks(root string) []validationIssue {
	issues := []validationIssue{}
	addMissing := func(path, target string) {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			issues = append(issues, validationIssue{Path: relativeDisplay(root, path), Schema: "semantic-links", Message: "missing linked object: " + relativeDisplay(root, target)})
		}
	}
	workFiles, _ := listJSONFiles(filepath.Join(root, ".ai-flow", "work-items"))
	for _, path := range workFiles {
		var item WorkItem
		if readJSON(path, &item) != nil {
			continue
		}
		if item.GoalID != nil {
			addMissing(path, filepath.Join(root, ".ai-flow", "goals", *item.GoalID+".json"))
		}
		for _, id := range item.RequirementIDs {
			addMissing(path, filepath.Join(root, ".ai-flow", "requirements", id+".json"))
		}
		for _, id := range item.EvidenceIDs {
			addMissing(path, evidencePath(root, id))
		}
	}
	runFiles, _ := filepath.Glob(filepath.Join(root, ".ai-flow", "runs", "RUN-*", "run.json"))
	for _, path := range runFiles {
		var run HarnessRun
		if readJSON(path, &run) != nil {
			continue
		}
		addMissing(path, workItemPath(root, run.WorkItemID))
		for _, id := range run.CheckpointIDs {
			addMissing(path, checkpointPath(root, run.ID, id))
		}
		for _, id := range run.EvidenceIDs {
			addMissing(path, evidencePath(root, id))
		}
	}
	evidenceFiles, _ := listJSONFiles(filepath.Join(root, ".ai-flow", "evidence"))
	for _, path := range evidenceFiles {
		var evidence Evidence
		if readJSON(path, &evidence) != nil {
			continue
		}
		addMissing(path, workItemPath(root, evidence.WorkItemID))
		addMissing(path, runPath(root, evidence.RunID))
		addMissing(path, filepath.Join(root, filepath.FromSlash(evidence.LogPath)))
	}
	return issues
}

func fileURL(path string) string {
	abs, _ := filepath.Abs(path)
	return (&url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}).String()
}
