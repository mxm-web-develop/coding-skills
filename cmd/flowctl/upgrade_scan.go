package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func upgradeSchema(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 3 || !strings.HasSuffix(path, ".json") {
		return ""
	}
	flat := map[string]string{"goals": "goal", "requirements": "requirement", "plans": "plan", "work-items": "work-item", "decisions": "decision", "tests": "test-spec", "evidence": "evidence", "releases": "release"}
	if len(parts) == 3 && flat[parts[1]] != "" {
		return flat[parts[1]] + ".schema.json"
	}
	if parts[1] == "runs" && filepath.Base(path) == "run.json" {
		return "run.schema.json"
	}
	if parts[1] == "runs" && strings.Contains(path, "/checkpoints/") {
		return "checkpoint.schema.json"
	}
	return ""
}

func normalizeUpgradeObject(path string, object map[string]any, schemaRoot string) (map[string]any, []string) {
	var schema struct {
		Properties map[string]any `json:"properties"`
	}
	_ = readSemanticJSON(filepath.Join(schemaRoot, upgradeSchema(path)), &schema)
	result := map[string]any{}
	notes := []string{}
	for key, value := range object {
		if _, ok := schema.Properties[key]; ok {
			result[key] = value
		} else {
			notes = append(notes, "original field preserved in backup: "+key)
		}
	}
	// Versions retain original identity, progress, requirements, and pending
	// choices. New mandatory fields must never invent user approval or a pass.
	if strings.Contains(path, "/work-items/") {
		if result["status"] != "done" && result["status"] != "cancelled" {
			result["workflow_version"] = float64(1)
		}
		for _, key := range []string{"dependencies", "protected_paths", "risks", "required_tests", "approval_requirements"} {
			if _, ok := result[key]; !ok {
				result[key] = []any{}
			}
		}
	}
	if _, ok := result["schema_version"]; !ok {
		result["schema_version"] = float64(1)
		notes = append(notes, "confirm imported record version")
	}
	if _, ok := result["revision"]; !ok && !strings.Contains(path, "/evidence/") && !strings.Contains(path, "/checkpoints/") {
		result["revision"] = float64(1)
		notes = append(notes, "confirm imported record revision")
	}
	sort.Strings(notes)
	return result, notes
}

func scanUpgradeEngineering(root string) error {
	path := filepath.Join(root, ".ai-flow/baseline/engineering-profile.json")
	var existing map[string]any
	_ = readSemanticJSON(path, &existing)
	languages := []map[string]any{}
	managers := []string{}
	builds := []string{}
	modules := []string{}
	unknowns := []string{}
	commands := map[string][]string{}
	for _, kind := range []string{"install", "build", "format", "lint", "typecheck", "unit", "integration", "e2e", "visual"} {
		commands[kind] = []string{}
	}
	seen := map[string]bool{}
	add := func(name, rel, manager string) {
		languages = append(languages, map[string]any{"name": name, "evidence": []string{rel}})
		modules = uniqueAppend(modules, filepath.ToSlash(filepath.Dir(rel)))
		if manager != "" {
			managers = uniqueAppend(managers, manager)
		}
	}
	manifests := []string{}
	excluded := []string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		rel := relativeDisplay(root, path)
		if info.IsDir() {
			if rel != "." && contains([]string{"fixtures", "__fixtures__", "testdata", "examples", "samples"}, strings.ToLower(info.Name())) {
				excluded = append(excluded, rel)
				return filepath.SkipDir
			}
			if rel != "." && (shouldSkipWorkspacePath(rel) || contains([]string{"node_modules", "vendor", "dist", "package", ".venv", "target", "build"}, info.Name())) {
				return filepath.SkipDir
			}
			if rel != "." {
				if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
					unknowns = append(unknowns, "嵌套仓库需独立检查："+rel)
					return filepath.SkipDir
				}
			}
			return nil
		}
		name := info.Name()
		dir := filepath.ToSlash(filepath.Dir(rel))
		prefix := ""
		if dir != "." {
			prefix = "cd '" + strings.ReplaceAll(dir, "'", "'\\''") + "' && "
		}
		switch name {
		case "go.mod":
			add("Go", rel, "Go modules")
			builds = uniqueAppend(builds, "Go toolchain")
			commands["unit"] = append(commands["unit"], prefix+"go test ./...")
			commands["lint"] = append(commands["lint"], prefix+"go vet ./...")
			seen["go"] = true
		case "package.json":
			var pkg struct {
				Scripts      map[string]string `json:"scripts"`
				Dependencies map[string]string `json:"dependencies"`
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if json.Unmarshal(data, &pkg) != nil {
				unknowns = append(unknowns, "无法读取工程清单："+rel)
				break
			}
			manager := "npm"
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), "pnpm-lock.yaml")); err == nil {
				manager = "pnpm"
			}
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), "yarn.lock")); err == nil {
				manager = "yarn"
			}
			add("JavaScript/TypeScript", rel, manager)
			for script, kind := range map[string]string{"test": "unit", "build": "build", "lint": "lint", "typecheck": "typecheck", "test:e2e": "e2e"} {
				if pkg.Scripts[script] != "" {
					commands[kind] = append(commands[kind], prefix+manager+" run "+script)
				}
			}
			seen["javascript-typescript"] = true
		case "pyproject.toml", "requirements.txt":
			add("Python", rel, "")
			unknowns = append(unknowns, "请核对 Python 测试与环境命令："+rel)
			seen["python"] = true
		case "Cargo.toml":
			add("Rust", rel, "Cargo")
			commands["unit"] = append(commands["unit"], prefix+"cargo test")
			seen["rust"] = true
		case "pom.xml", "build.gradle", "build.gradle.kts":
			add("Java/Kotlin", rel, "")
			unknowns = append(unknowns, "请核对 JVM 构建与测试入口："+rel)
		default:
			return nil
		}
		manifests = append(manifests, rel)
		return nil
	})
	if err != nil {
		return err
	}
	playbooks := []string{"engineering-quality-baseline"}
	for _, name := range []string{"go", "javascript-typescript", "python", "rust"} {
		if seen[name] {
			playbooks = append(playbooks, name)
		}
	}
	if len(languages) == 0 {
		unknowns = append(unknowns, "未识别到可确认的工程清单；保留已有环境信息并人工补齐")
	}
	if existing != nil {
		unknowns = append(unknowns, "原工程画像已备份；复核新增扫描与原有自定义命令、社区技能和视觉验证要求")
	}
	profile := map[string]any{"schema_version": 1, "revision": 1, "detected_at": time.Now().UTC().Format(time.RFC3339Nano), "git_sha": gitSHA(root), "languages": languages, "frameworks": []any{}, "package_managers": managers, "build_systems": builds, "architecture": map[string]any{"style": "repository scan; preserve accepted architecture decisions", "module_roots": modules, "generated_roots": []string{"docs/board"}, "public_api_roots": []string{}, "constraints": []string{"升级不改变原计划和现行需求，未知项需核对"}}, "commands": commands, "selected_playbooks": playbooks, "community_skills": []any{}, "visual_testing": map[string]any{"required": seen["javascript-typescript"], "browsers": []string{}, "viewports": []string{}}, "unknowns": unknowns}
	// Preserve specialized knowledge; a manifest scan cannot replace confirmed
	// architecture, community skills, or a project's established test commands.
	for _, key := range []string{"frameworks", "architecture", "community_skills", "visual_testing"} {
		if value, ok := existing[key]; ok {
			profile[key] = value
		}
	}
	// Old commands stay in the confirmed profile. Do not merge unverified
	// discoveries into it: stale fixture commands otherwise live forever.
	if revision, ok := existing["revision"].(float64); ok {
		profile["revision"] = int(revision) + 1
	}
	if len(languages) == 0 {
		for _, key := range []string{"languages", "package_managers", "build_systems", "selected_playbooks"} {
			if value, ok := existing[key]; ok {
				profile[key] = value
			}
		}
	}
	candidate := filepath.Join(root, ".ai-flow/baseline/engineering-candidate.json")
	if err := writeJSONAtomic(candidate, profile); err != nil {
		return err
	}
	return writeJSONAtomic(filepath.Join(root, ".ai-flow/baseline/upgrade-scan.json"), map[string]any{"scanned_at": profile["detected_at"], "git_sha": gitSHA(root), "manifests": manifests, "unknowns": unknowns, "tests_executed": false, "status": "candidate", "excluded_components": excluded, "confirmed_profile_unchanged": true})
}
