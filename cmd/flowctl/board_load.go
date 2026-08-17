package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
)

func loadBoardData(root string, status projectStatus) (boardData, error) {
	data := boardData{Status: status}
	var err error
	if data.Goals, err = loadBoardObjects[boardGoal](root, "goals"); err != nil {
		return data, err
	}
	if data.Requirements, err = loadBoardObjects[boardRequirement](root, "requirements"); err != nil {
		return data, err
	}
	if data.Plans, err = loadBoardObjects[boardPlan](root, "plans"); err != nil {
		return data, err
	}
	if data.WorkItems, err = loadBoardObjects[WorkItem](root, "work-items"); err != nil {
		return data, err
	}
	if data.Decisions, err = loadBoardObjects[boardDecision](root, "decisions"); err != nil {
		return data, err
	}
	if data.Tests, err = loadBoardObjects[boardTestSpec](root, "tests"); err != nil {
		return data, err
	}
	if data.Evidence, err = loadBoardObjects[Evidence](root, "evidence"); err != nil {
		return data, err
	}
	if data.Releases, err = loadBoardObjects[boardRelease](root, "releases"); err != nil {
		return data, err
	}

	profilePath := filepath.Join(root, ".ai-flow", "baseline", "engineering-profile.json")
	if _, err := os.Stat(profilePath); err == nil {
		var profile boardEngineeringProfile
		if err := readBoardJSON(profilePath, &profile); err != nil {
			return data, err
		}
		data.Engineering = &profile
	} else if !os.IsNotExist(err) {
		return data, err
	}
	return data, nil
}

func loadBoardObjects[T any](root, directory string) ([]T, error) {
	files, err := listJSONFiles(filepath.Join(root, ".ai-flow", directory))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	objects := make([]T, 0, len(files))
	for _, path := range files {
		var object T
		if err := readBoardJSON(path, &object); err != nil {
			return nil, err
		}
		objects = append(objects, object)
	}
	return objects, nil
}

func readBoardJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
