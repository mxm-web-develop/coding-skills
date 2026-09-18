package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readProjectSource(root, path string) ([]byte, error) {
	if path == "" || filepath.IsAbs(path) || filepath.ToSlash(filepath.Clean(path)) != path || strings.Contains(path, "\\") {
		return nil, fmt.Errorf("source must be a normalized repository-relative file: %s", path)
	}
	if err := ensurePathInsideRepository(root, path); err != nil {
		return nil, err
	}
	resolved := resolveRecordPath(root, filepath.Join(root, filepath.FromSlash(path)))
	real, err := filepath.EvalSymlinks(resolved)
	if err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(absolute, real)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, errors.New("source escapes project")
	}
	return os.ReadFile(real)
}
func snapshotSources(root string, paths []string) ([]SourceSnapshot, error) {
	result := []SourceSnapshot{}
	seen := map[string]bool{}
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		data, err := readProjectSource(root, p)
		if err != nil {
			return nil, err
		}
		result = append(result, SourceSnapshot{p, hashBytes(data)})
	}
	return result, nil
}
func checkSnapshots(root string, sources []SourceSnapshot) error {
	for _, s := range sources {
		data, err := readProjectSource(root, s.Path)
		if err != nil {
			return err
		}
		if hashBytes(data) != s.SHA256 {
			return fmt.Errorf("confirmed input changed; review before continuing: %s", s.Path)
		}
	}
	return nil
}

// A pending linked choice is not silently omitted from an executor's context.
func executionDecisionSources(root string, w WorkItem) ([]string, error) {
	files, err := recordFiles(root, "decisions", false)
	if err != nil {
		return nil, err
	}
	paths := []string{}
	for _, path := range files {
		var d struct {
			ID           string `json:"id"`
			Status       string `json:"status"`
			Confirmation struct {
				Status string `json:"status"`
			} `json:"confirmation"`
			WorkIDs        []string `json:"work_item_ids"`
			RequirementIDs []string `json:"requirement_ids"`
			GoalID         string   `json:"goal_id"`
		}
		if err := readSemanticJSON(path, &d); err != nil {
			return nil, err
		}
		relevant := contains(d.WorkIDs, w.ID)
		for _, id := range w.RequirementIDs {
			if contains(d.RequirementIDs, id) {
				relevant = true
			}
		}
		if w.GoalID != nil && d.GoalID == *w.GoalID {
			relevant = true
		}
		if !relevant || contains([]string{"superseded", "rejected", "archived"}, d.Status) {
			continue
		}
		if d.Status != "accepted" || (d.Confirmation.Status != "" && d.Confirmation.Status != "confirmed") {
			return nil, fmt.Errorf("resolve the linked design choice before execution: %s", d.ID)
		}
		paths = append(paths, relativeDisplay(root, path))
	}
	return paths, nil
}
