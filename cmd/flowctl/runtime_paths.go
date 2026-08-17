package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func readFlatYAML(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		values[strings.TrimSpace(parts[0])] = strings.Trim(strings.TrimSpace(parts[1]), "\"'")
	}
	return values, nil
}

func resolveRoot(rootArg string, requireInstall bool) (string, error) {
	start := rootArg
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = cwd
	}
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("root is not a directory: %s", abs)
	}
	if rootArg != "" {
		if requireInstall {
			if err := ensureInstalled(abs); err != nil {
				return "", err
			}
		}
		return filepath.Clean(abs), nil
	}

	current := filepath.Clean(abs)
	for {
		if _, err := os.Stat(filepath.Join(current, ".ai-flow")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	if requireInstall {
		return "", errors.New("no .ai-flow directory found; run the installer first")
	}
	return filepath.Clean(abs), nil
}

func ensureInstalled(root string) error {
	path := filepath.Join(root, ".ai-flow", "bin", executableName("flowctl"))
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("AI Flow runtime is not installed at %s", path)
	}
	return nil
}

func writeIfMissing(path, content string, mode os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), mode)
}

func yamlScalar(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", " ", "\r", " ")
	return "\"" + replacer.Replace(strings.TrimSpace(value)) + "\""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func executableName(name string) string {
	if filepath.Separator == '\\' {
		return name + ".exe"
	}
	return name
}

func missingMessage(items []string) string {
	if len(items) == 0 {
		return "all required files are present"
	}
	return "missing: " + strings.Join(items, ", ")
}

func relativeDisplay(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
