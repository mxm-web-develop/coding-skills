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
	for name, expected := range expectedBoardFiles(data) {
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
	return issues
}
