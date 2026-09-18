package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type ProfileConfirmation struct {
	Reviewer    string           `json:"reviewer"`
	WorkID      string           `json:"work_id"`
	ConfirmedAt string           `json:"confirmed_at"`
	Sources     []SourceSnapshot `json:"sources"`
}

func runProfileConfirmation(args []string) error {
	fs := flag.NewFlagSet("memory confirm-profile", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	work := fs.String("work", "", "originating task")
	by := fs.String("by", "", "reviewer who inspected the profile")
	file := fs.String("file", ".ai-flow/baseline/engineering-candidate.json", "reviewed profile path within project")
	var sources stringListFlag
	fs.Var(&sources, "source", "inspected manifest/build/CI source; repeatable")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	if _, err := readWorkItem(root, *work); err != nil {
		return err
	}
	if strings.TrimSpace(*by) == "" || len(sources) == 0 {
		return errors.New("record the profile reviewer and inspected configuration sources")
	}
	unlock, err := projectMutationLock(root)
	if err != nil {
		return err
	}
	defer unlock()
	data, err := readProjectSource(root, *file)
	if err != nil {
		return err
	}
	var profile any
	if err = json.Unmarshal(data, &profile); err != nil {
		return err
	}
	schemaRoot, err := findSchemaRoot(root)
	if err != nil {
		return err
	}
	schemas, err := compileSchemas(schemaRoot)
	if err != nil {
		return err
	}
	if err = schemas["engineering-profile.schema.json"].Validate(profile); err != nil {
		return err
	}
	snapshots, err := snapshotSources(root, sources)
	if err != nil {
		return err
	}
	canonical := ".ai-flow/baseline/engineering-profile.json"
	// Hash the promoted bytes too, so later hand edits cannot masquerade as the
	// reviewed profile. Configuration changes invalidate this confirmation.
	snapshots = append(snapshots, SourceSnapshot{Path: canonical, SHA256: hashBytes(data)})
	if err = writeBytesAtomic(filepath.Join(root, canonical), data); err != nil {
		return err
	}
	c := ProfileConfirmation{Reviewer: *by, WorkID: *work, ConfirmedAt: time.Now().UTC().Format(time.RFC3339Nano), Sources: snapshots}
	return writeJSONAtomic(filepath.Join(root, ".ai-flow/baseline/profile-confirmation.json"), c)
}
func checkProfileConfirmation(root string) error {
	var c ProfileConfirmation
	if err := readJSON(filepath.Join(root, ".ai-flow/baseline/profile-confirmation.json"), &c); err != nil {
		return fmt.Errorf("review the project profile before model handoff: %w", err)
	}
	if c.Reviewer == "" || len(c.Sources) < 2 {
		return errors.New("profile review needs its source evidence")
	}
	return checkSnapshots(root, c.Sources)
}
