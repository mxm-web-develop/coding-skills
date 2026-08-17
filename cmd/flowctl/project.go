package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
		"releases", "locks", "workspace-cleanup", "archive",
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
		if err := writeIfMissing(filepath.Join(root, filepath.FromSlash(rel)), content, 0o644); err != nil {
			return err
		}
	}
	if err := renderBoard(root); err != nil {
		return err
	}
	fmt.Printf("Initialized %s AI Flow project %q at %s\n", *mode, *name, root)
	return nil
}

func nextActionForMode(mode string) string {
	if mode == "existing" {
		return "adopt_existing_project"
	}
	return "discover_product_goal"
}
