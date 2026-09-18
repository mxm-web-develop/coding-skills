package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type archiveCatalog map[string]string

func readArchiveCatalog(root string) (archiveCatalog, error) {
	catalog := archiveCatalog{}
	err := readJSON(filepath.Join(root, ".ai-flow", "archive", "catalog.json"), &catalog)
	if os.IsNotExist(err) {
		return catalog, nil
	}
	if err != nil {
		return nil, err
	}
	for source, target := range catalog {
		if !strings.HasPrefix(source, ".ai-flow/") || !strings.HasPrefix(target, ".ai-flow/archive/completed/") {
			return nil, errors.New("invalid archive catalog path")
		}
		for _, path := range []string{source, target} {
			if err := ensurePathInsideRepository(root, path); err != nil {
				return nil, err
			}
			if strings.Contains(path, "..") || strings.Contains(path, "\\") {
				return nil, errors.New("unsafe archive catalog path")
			}
		}
	}
	return catalog, nil
}

func resolveRecordPath(root, path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	catalog, err := readArchiveCatalog(root)
	if err != nil {
		return path
	}
	if target := catalog[relativeDisplay(root, path)]; target != "" {
		return filepath.Join(root, filepath.FromSlash(target))
	}
	return path
}

func recordFiles(root, directory string, history bool) ([]string, error) {
	files, err := listJSONFiles(filepath.Join(root, ".ai-flow", directory))
	if err != nil || !history {
		return files, err
	}
	catalog, err := readArchiveCatalog(root)
	if err != nil {
		return nil, err
	}
	prefix := ".ai-flow/" + directory + "/"
	for source := range catalog {
		if strings.HasPrefix(source, prefix) && !strings.Contains(strings.TrimPrefix(source, prefix), "/") && strings.HasSuffix(source, ".json") {
			path := resolveRecordPath(root, filepath.Join(root, filepath.FromSlash(source)))
			if !contains(files, path) {
				files = append(files, path)
			}
		}
	}
	sort.Strings(files)
	return files, nil
}

func recordGlob(root, pattern string) []string {
	files, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(pattern)))
	catalog, _ := readArchiveCatalog(root)
	for source := range catalog {
		if match, _ := filepath.Match(pattern, source); match {
			path := resolveRecordPath(root, filepath.Join(root, filepath.FromSlash(source)))
			if !contains(files, path) {
				files = append(files, path)
			}
		}
	}
	sort.Strings(files)
	return files
}

func archiveCompletedWork(root string, item *WorkItem) error {
	catalog, err := readArchiveCatalog(root)
	if err != nil {
		return err
	}
	dir := ".ai-flow/archive/completed/" + item.ID
	sources := []string{".ai-flow/work-items/" + item.ID + ".json"}
	var run HarnessRun
	if item.RunID != nil {
		run, err = readRun(root, *item.RunID)
		if err != nil {
			return err
		}
		sources = append(sources, ".ai-flow/runs/"+run.ID+"/run.json")
		for _, id := range run.CheckpointIDs {
			sources = append(sources, ".ai-flow/runs/"+run.ID+"/checkpoints/"+id+".json")
		}
	}
	for _, id := range item.EvidenceIDs {
		e, err := readEvidence(root, id)
		if err != nil {
			return err
		}
		sources = append(sources, ".ai-flow/evidence/"+id+".json", e.LogPath)
	}
	files, err := recordFiles(root, "tests", true)
	if err != nil {
		return err
	}
	for _, path := range files {
		var test traceTest
		if readSemanticJSON(path, &test) == nil && test.WorkItemID == item.ID {
			sources = append(sources, ".ai-flow/tests/"+test.ID+".json")
		}
	}
	// Only task-scoped generated process material is automatically archived.
	for _, kind := range []string{"reports", "prototypes"} {
		base := filepath.Join(root, ".ai-flow", kind, item.ID)
		err := filepath.Walk(base, func(path string, info os.FileInfo, e error) error {
			if os.IsNotExist(e) {
				return nil
			}
			if e != nil {
				return e
			}
			if !info.IsDir() {
				sources = append(sources, relativeDisplay(root, path))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	completed := *item
	completed.Status = "done"
	completed.ArchivePath = dir
	completed.Revision++
	completed.UpdatedAt = now
	if run.ID != "" {
		run.Status = "completed"
		run.Phase = "completed"
		run.CompletedAt = &now
		run.UpdatedAt = now
		run.Revision++
	}
	for _, source := range sources {
		if err := ensurePathInsideRepository(root, source); err != nil {
			return err
		}
		path := resolveRecordPath(root, filepath.Join(root, filepath.FromSlash(source)))
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if source == ".ai-flow/work-items/"+item.ID+".json" {
			content, err = json.MarshalIndent(completed, "", "  ")
		}
		if source == ".ai-flow/runs/"+run.ID+"/run.json" {
			content, err = json.MarshalIndent(run, "", "  ")
		}
		if err != nil {
			return err
		}
		target := dir + "/records/" + strings.TrimPrefix(source, ".ai-flow/")
		if err := writeBytesAtomic(filepath.Join(root, filepath.FromSlash(target)), content); err != nil {
			return err
		}
		catalog[source] = target
	}
	if err := writeBytesAtomic(filepath.Join(root, filepath.FromSlash(dir), "summary.md"), []byte("# "+item.Title+"\n\n"+item.CompletionSummary+"\n\n人工验收："+item.Acceptance.Feedback+"\n")); err != nil {
		return err
	}
	// Publish the fallback index before removing canonical copies. Retrying a
	// closing task can resolve either location after any interrupted removal.
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", "archive", "catalog.json"), catalog); err != nil {
		return err
	}
	sort.SliceStable(sources, func(i, j int) bool {
		return sources[j] == ".ai-flow/work-items/"+item.ID+".json" && sources[i] != sources[j]
	})
	for _, source := range sources {
		if err := os.Remove(filepath.Join(root, filepath.FromSlash(source))); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	*item = completed
	return nil
}

func writeBytesAtomic(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".flowctl-write-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if err := temp.Chmod(0644); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(content); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

func projectMutationLock(root string) (func(), error) {
	path := filepath.Join(root, ".ai-flow", "locks", "mutation.lock")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("another state mutation is running; retry after it ends (after a crashed process, inspect and remove only .ai-flow/locks/mutation.lock): %w", err)
	}
	fmt.Fprintln(f, os.Getpid())
	f.Close()
	return func() { _ = os.Remove(path) }, nil
}

func globWithHistory(root, pattern string) ([]string, error) { return recordGlob(root, pattern), nil }
func skipArchiveSnapshot(root, path string) bool {
	rel := relativeDisplay(root, path)
	return contains([]string{".ai-flow/archive/completed", ".ai-flow/archive/upgrades", ".ai-flow/archive/generated"}, rel)
}

// A published archive can survive interruption between any two removals.
// Never delete a newly modified canonical file during recovery.
func finishArchiveRemoval(root string, item WorkItem) error {
	catalog, err := readArchiveCatalog(root)
	if err != nil {
		return err
	}
	for source, target := range catalog {
		if !strings.HasPrefix(target, item.ArchivePath+"/records/") {
			continue
		}
		original := filepath.Join(root, filepath.FromSlash(source))
		data, err := os.ReadFile(original)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		archived, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(target)))
		if err != nil {
			return err
		}
		if !sameBytes(data, archived) {
			return fmt.Errorf("canonical record changed during closeout; reconcile before removing: %s", source)
		}
		if err := os.Remove(original); err != nil {
			return err
		}
	}
	if err := os.Remove(leasePath(root, item.ID)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
