package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var objectIDPattern = regexp.MustCompile(`^[A-Z]+-[0-9]{8}-[a-f0-9]{8}$`)

type stringListFlag []string

func (s *stringListFlag) String() string { return strings.Join(*s, ",") }
func (s *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("list value cannot be empty")
	}
	*s = append(*s, value)
	return nil
}

func newObjectID(prefix string) (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().UTC().Format("20060102"), hex.EncodeToString(bytes)), nil
}

func requireObjectID(id, expectedPrefix string) error {
	if !objectIDPattern.MatchString(id) || !strings.HasPrefix(id, expectedPrefix+"-") {
		return fmt.Errorf("invalid %s id: %s", expectedPrefix, id)
	}
	return nil
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	temp, err := os.CreateTemp(filepath.Dir(path), ".flowctl-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0o644); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.Write(data); err != nil {
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
	return os.Rename(tempName, path)
}

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

func listJSONFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	files := []string{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(files)
	return files, nil
}

func appendEvent(root, eventType, objectType, objectID, runID string, revision int, data map[string]any) error {
	eventID, err := newObjectID("EVT")
	if err != nil {
		return err
	}
	record := eventRecord{
		SchemaVersion: 1,
		ID:            eventID,
		Type:          eventType,
		ObjectType:    objectType,
		ObjectID:      objectID,
		RunID:         runID,
		Revision:      revision,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Data:          data,
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		return err
	}
	eventsDir := filepath.Join(root, ".ai-flow", "events")
	if err := os.MkdirAll(eventsDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(eventsDir, time.Now().UTC().Format("2006-01")+".jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	if _, err := writer.Write(append(encoded, '\n')); err != nil {
		return err
	}
	return writer.Flush()
}

func gitSHA(root string) string {
	command := exec.Command("git", "-C", root, "rev-parse", "HEAD")
	output, err := command.Output()
	if err != nil {
		return "not-a-git-repository"
	}
	sha := strings.TrimSpace(string(output))
	if sha == "" {
		return "unknown"
	}
	return sha
}

func gitCommitExists(root, sha string) bool {
	if strings.TrimSpace(sha) == "" {
		return false
	}
	command := exec.Command("git", "-C", root, "cat-file", "-e", strings.TrimSpace(sha)+"^{commit}")
	return command.Run() == nil
}

func gitWorktreeClean(root string) (bool, error) {
	command := exec.Command("git", "-C", root, "status", "--porcelain=v1", "--untracked-files=all", "-z")
	output, err := command.CombinedOutput()
	if err != nil {
		if isNotGitRepositoryOutput(err, output) {
			return false, nil
		}
		return false, err
	}
	return len(bytes.TrimSpace(output)) == 0, nil
}

func gitWorktreeFingerprint(root string) (string, error) {
	hash := sha256.New()
	if err := hashWorkspaceTree(hash, root); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func hashWorkspaceTree(hash interface{ Write([]byte) (int, error) }, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		relSlash := filepath.ToSlash(rel)
		if shouldSkipWorkspacePath(relSlash) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		_, _ = hash.Write([]byte(relSlash))
		_, _ = hash.Write([]byte("\n"))
		_, _ = hash.Write([]byte(info.Mode().String()))
		_, _ = hash.Write([]byte("\n"))
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			_, _ = hash.Write([]byte(target))
			_, _ = hash.Write([]byte("\n"))
			return nil
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		if _, err := io.Copy(hash, file); err != nil {
			return err
		}
		_, _ = hash.Write([]byte("\n"))
		return nil
	})
}

func shouldSkipWorkspacePath(relSlash string) bool {
	ignoredRoots := []string{".ai-flow", ".git", ".cursor", ".claude", ".codex", "CLAUDE.md", "AGENTS.md"}
	for _, ignored := range ignoredRoots {
		if relSlash == ignored || strings.HasPrefix(relSlash, ignored+"/") {
			return true
		}
	}
	return false
}

func isNotGitRepository(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "not a git repository") || strings.Contains(err.Error(), "ambiguous argument"))
}

func isNotGitRepositoryOutput(err error, output []byte) bool {
	if err == nil {
		return false
	}
	if isNotGitRepository(err) {
		return true
	}
	text := strings.ToLower(string(output))
	return strings.Contains(text, "not a git repository") || strings.Contains(text, "fatal: not a git repository")
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func workItemPath(root, id string) string {
	return filepath.Join(root, ".ai-flow", "work-items", id+".json")
}

func runPath(root, id string) string {
	return filepath.Join(root, ".ai-flow", "runs", id, "run.json")
}

func checkpointPath(root, runID, checkpointID string) string {
	return filepath.Join(root, ".ai-flow", "runs", runID, "checkpoints", checkpointID+".json")
}

func evidencePath(root, id string) string {
	return filepath.Join(root, ".ai-flow", "evidence", id+".json")
}

func leasePath(root, workID string) string {
	return filepath.Join(root, ".ai-flow", "locks", workID+".json")
}

func readWorkItem(root, id string) (WorkItem, error) {
	var item WorkItem
	if err := requireObjectID(id, "WI"); err != nil {
		return item, err
	}
	if err := readJSON(workItemPath(root, id), &item); err != nil {
		return item, err
	}
	return item, nil
}

func readRun(root, id string) (HarnessRun, error) {
	var run HarnessRun
	if err := requireObjectID(id, "RUN"); err != nil {
		return run, err
	}
	if err := readJSON(runPath(root, id), &run); err != nil {
		return run, err
	}
	return run, nil
}

func readEvidence(root, id string) (Evidence, error) {
	var evidence Evidence
	if err := requireObjectID(id, "EV"); err != nil {
		return evidence, err
	}
	if err := readJSON(evidencePath(root, id), &evidence); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func checkExpectedRevision(actual, expected int) error {
	if expected > 0 && actual != expected {
		return fmt.Errorf("revision conflict: expected %d, found %d", expected, actual)
	}
	return nil
}

func uniqueAppend(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
