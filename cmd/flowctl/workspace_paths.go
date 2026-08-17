package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readSemanticJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func canonicalJSONFileDigest(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var value any
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func pathCoveredByScope(path string, scopes []string) bool {
	path = filepath.Clean(filepath.FromSlash(path))
	for _, scope := range scopes {
		scope = filepath.Clean(filepath.FromSlash(scope))
		if scope == "." {
			return true
		}
		if path == scope || strings.HasPrefix(path, scope+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func pathsOverlap(left, right string) bool {
	left = filepath.Clean(filepath.FromSlash(strings.ReplaceAll(left, "\\", "/")))
	right = filepath.Clean(filepath.FromSlash(strings.ReplaceAll(right, "\\", "/")))
	if left == right {
		return true
	}
	separator := string(filepath.Separator)
	return strings.HasPrefix(left, right+separator) || strings.HasPrefix(right, left+separator)
}

func archiveTargetPreservesPath(action, sourcePath, targetPath string) bool {
	prefix := ".ai-flow/archive/legacy-files/"
	if action == "archive-code" {
		prefix = ".ai-flow/archive/legacy-code/"
	}
	source := strings.TrimPrefix(strings.ReplaceAll(sourcePath, "\\", "/"), "./")
	target := strings.ReplaceAll(targetPath, "\\", "/")
	if !strings.HasPrefix(target, prefix) || !strings.HasSuffix(target, "/"+source) {
		return false
	}
	versionBucket := strings.TrimSuffix(strings.TrimPrefix(target, prefix), "/"+source)
	return versionBucket != "" && !strings.Contains(versionBucket, "/")
}

func ensurePathInsideRepository(root, relativePath string) error {
	rootAbsolute, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	rootAbsolute, err = filepath.EvalSymlinks(rootAbsolute)
	if err != nil {
		return err
	}
	relativePath = strings.ReplaceAll(relativePath, "\\", "/")
	candidate := filepath.Join(rootAbsolute, filepath.FromSlash(relativePath))
	ancestor := candidate
	for {
		if _, err := os.Lstat(ancestor); err == nil {
			break
		} else if !os.IsNotExist(err) {
			return err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return errors.New("cannot resolve an existing ancestor")
		}
		ancestor = parent
	}
	resolved, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(rootAbsolute, resolved)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return errors.New("resolved path leaves the repository")
	}
	return nil
}

func sha256RepositoryPath(root, relativePath string) (string, error) {
	if err := ensurePathInsideRepository(root, relativePath); err != nil {
		return "", err
	}
	relativePath = strings.ReplaceAll(relativePath, "\\", "/")
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode().IsRegular() {
		return sha256File(path)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(path)
		if err != nil {
			return "", err
		}
		digest := sha256.Sum256([]byte(target))
		return hex.EncodeToString(digest[:]), nil
	}
	if !info.IsDir() {
		return "", errors.New("unsupported path type")
	}
	hash := sha256.New()
	err = filepath.WalkDir(path, func(entryPath string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entryPath == path {
			return nil
		}
		relativeEntry, err := filepath.Rel(path, entryPath)
		if err != nil {
			return err
		}
		relativeEntry = filepath.ToSlash(relativeEntry)
		kind := "file"
		payload := ""
		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}
		if entryInfo.Mode()&os.ModeSymlink != 0 {
			kind = "symlink"
			payload, err = os.Readlink(entryPath)
			if err != nil {
				return err
			}
		} else if entry.IsDir() {
			kind = "directory"
		} else if entryInfo.Mode().IsRegular() {
			payload, err = sha256File(entryPath)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unsupported entry type: %s", relativeEntry)
		}
		_, err = fmt.Fprintf(hash, "%s\x00%s\x00%s\n", relativeEntry, kind, payload)
		return err
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
