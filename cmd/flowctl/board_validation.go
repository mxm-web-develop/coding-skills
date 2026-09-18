package main

import (
	"bytes"
	"os"
	"path/filepath"
)

func validateBoardFreshness(root string) []validationIssue {
	status, err := readStatus(root)
	if err != nil || !status.Initialized {
		return nil
	}
	data, err := loadBoardData(root, status)
	if err != nil {
		return nil // Object validation reports the underlying problem.
	}
	issues := []validationIssue{}
	expectedFiles := expectedBoardFiles(data)
	for name, content := range expectedPlanFiles(data) {
		expectedFiles[name] = content
	}
	for name, expected := range expectedFiles {
		path := filepath.Join(root, "docs", "board", name)
		actual, readErr := os.ReadFile(path)
		if readErr != nil {
			issues = append(issues, validationIssue{Path: relativeDisplay(root, path), Schema: "human-board", Message: "human board is missing; run flowctl render-board"})
			continue
		}
		if !bytes.Equal(actual, []byte(expected)) {
			issues = append(issues, validationIssue{Path: relativeDisplay(root, path), Schema: "human-board", Message: "human board is stale or manually edited; run flowctl render-board"})
		}
	}
	var managed []string
	if readJSON(filepath.Join(root, ".ai-flow", "state", "generated-board-files.json"), &managed) == nil {
		for _, name := range managed {
			if _, ok := expectedFiles[name]; !ok {
				issues = append(issues, validationIssue{Path: "docs/board/" + name, Schema: "human-board", Message: "inactive generated page remains in current index; render board"})
			}
		}
	}
	return issues
}
