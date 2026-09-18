package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func verificationFingerprint(root string) (string, error) {
	command := exec.Command("git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	output, err := command.Output()
	if err != nil {
		return gitWorktreeFingerprint(root)
	}
	names := strings.Split(string(output), "\x00")
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		if name == "" || shouldSkipWorkspacePath(name) || name == "docs/board" || strings.HasPrefix(name, "docs/board/") || strings.HasPrefix(name, ".agents/") {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(name))
		info, err := os.Lstat(path)
		if os.IsNotExist(err) {
			fmt.Fprintln(hash, name, "deleted")
			continue
		}
		if err != nil {
			return "", err
		}
		if info.IsDir() {
			continue
		}
		fmt.Fprintln(hash, name, info.Mode())
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return "", err
			}
			fmt.Fprintln(hash, target)
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		hash.Write(data)
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func currentWorkEvidence(root string, item WorkItem) ([]Evidence, error) {
	all := []Evidence{}
	for _, id := range item.EvidenceIDs {
		e, err := readEvidence(root, id)
		if err != nil {
			return nil, err
		}
		all = append(all, e)
	}
	return latestEvidence(all), nil
}

func validateWorkEvidence(root string, item WorkItem) error {
	all, err := currentWorkEvidence(root, item)
	if err != nil {
		return err
	}
	if len(all) == 0 {
		return errors.New("no verification has run")
	}
	fingerprint, err := verificationFingerprint(root)
	if err != nil {
		return err
	}
	required := append([]string(nil), item.RequiredTests...)
	testPaths, _ := recordFiles(root, "tests", true)
	for _, path := range testPaths {
		var test traceTest
		if readSemanticJSON(path, &test) == nil && test.WorkItemID == item.ID && test.Status != "retired" {
			required = uniqueAppend(required, test.ID)
		}
	}
	seen := map[string]bool{}
	for _, e := range all {
		if e.WorkItemID != item.ID {
			return errors.New("verification belongs to another task")
		}
		if e.GitSHA != gitSHA(root) || e.ContentSHA256 == "" || e.ContentSHA256 != fingerprint {
			return fmt.Errorf("verification for %s is stale; rerun on current code", e.TestID)
		}
		if evidenceResult(e) != "passed" || e.ExitCode != 0 {
			return fmt.Errorf("latest required check did not pass: %s", e.TestID)
		}
		if err := validateEvidenceLogPath(root, e.LogPath); err != nil {
			return err
		}
		digest, err := sha256File(resolveRecordPath(root, filepath.Join(root, filepath.FromSlash(e.LogPath))))
		if err != nil {
			return err
		}
		if digest != e.LogSHA256 {
			return errors.New("verification log hash mismatch")
		}
		seen[e.TestID] = true
	}
	for _, test := range required {
		if !seen[test] {
			return fmt.Errorf("required check has no current pass: %s", test)
		}
	}
	return nil
}

func qualityMatches(root, sha, fingerprint string) bool {
	current, err := verificationFingerprint(root)
	return err == nil && sha == gitSHA(root) && fingerprint != "" && fingerprint == current
}

func checkWorkStart(root string, item WorkItem) error {
	if err := requireProjectCompatible(root); err != nil {
		return err
	}
	for _, id := range item.Dependencies {
		dependency, err := readWorkItem(root, id)
		if err != nil {
			return err
		}
		if dependency.Status != "done" {
			return fmt.Errorf("dependency is unfinished: %s", dependency.Title)
		}
	}
	files, err := listJSONFiles(filepath.Join(root, ".ai-flow", "locks"))
	if err != nil {
		return err
	}
	for _, path := range files {
		var lease WorkLease
		if err := readJSON(path, &lease); err != nil {
			return err
		}
		if lease.WorkItemID == item.ID {
			continue
		}
		expiry, err := time.Parse(time.RFC3339, lease.ExpiresAt)
		if err != nil {
			return err
		}
		if time.Now().After(expiry) {
			continue
		}
		for _, left := range item.Scope {
			for _, right := range lease.Scope {
				if scopeOverlap(left, right) {
					return fmt.Errorf("another active task owns overlapping files: %s", lease.WorkItemID)
				}
			}
		}
	}
	return nil
}

func scopeOverlap(a, b string) bool {
	base := func(s string) string {
		s = strings.ReplaceAll(s, "\\", "/")
		if i := strings.IndexAny(s, "*?["); i >= 0 {
			s = s[:i]
		}
		s = strings.TrimSuffix(s, "/")
		if s == "" {
			return "."
		}
		return s
	}
	a, b = base(a), base(b)
	return a == "." || b == "." || pathsOverlap(a, b)
}

func checkRunBudget(root string, item WorkItem, run HarnessRun) error {
	if run.ID == "" {
		return nil
	}
	if run.Budgets.MaxElapsedMinutes > 0 {
		start, err := time.Parse(time.RFC3339, run.StartedAt)
		if err != nil {
			return err
		}
		if time.Since(start) > time.Duration(run.Budgets.MaxElapsedMinutes)*time.Minute {
			return errors.New("execution time budget reached; checkpoint and explicitly revise the budget before continuing")
		}
	}
	command := exec.Command("git", "-C", root, "diff", "--name-only", "-z", run.GitSHA)
	output, err := command.Output()
	if err != nil {
		return nil
	} // Non-Git projects retain time and test budgets.
	changed := strings.Split(string(output), "\x00")
	untracked, _ := exec.Command("git", "-C", root, "ls-files", "-z", "--others", "--exclude-standard").Output()
	changed = append(changed, strings.Split(string(untracked), "\x00")...)
	count := 0
	for _, path := range changed {
		if path == "" || shouldSkipWorkspacePath(path) || strings.HasPrefix(path, "docs/board/") {
			continue
		}
		if pathCoveredByScope(path, item.ProtectedPaths) {
			return fmt.Errorf("protected file was changed: %s", path)
		}
		count++
	}
	if run.Budgets.MaxChangedFiles > 0 && count > run.Budgets.MaxChangedFiles {
		return errors.New("changed-file budget reached; checkpoint and explicitly revise the scope/budget")
	}
	return nil
}

// Retry budgets count consecutive failures of the same check. A successful
// attempt resets that check, so unrelated tests do not consume its retry budget.
func checkRetryBudget(root string, item WorkItem, run HarnessRun, testID string) error {
	if run.ID == "" {
		return nil
	}
	failures := 0
	for i := len(run.EvidenceIDs) - 1; i >= 0; i-- {
		e, err := readEvidence(root, run.EvidenceIDs[i])
		if err != nil {
			return err
		}
		if e.TestID != testID {
			continue
		}
		if e.Result == "passed" {
			break
		}
		failures++
	}
	if failures > run.Budgets.MaxRetries {
		return errors.New("test retry budget reached; diagnose and explicitly revise the budget before retrying")
	}
	return nil
}

func sameBytes(a, b []byte) bool { return bytes.Equal(a, b) }
