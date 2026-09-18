package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type UpgradeFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Action string `json:"action"`
	Reason string `json:"reason,omitempty"`
}
type UpgradeRecord struct {
	ID             string        `json:"id"`
	From           string        `json:"from"`
	To             string        `json:"to"`
	Status         string        `json:"status"`
	CreatedAt      string        `json:"created_at"`
	GitSHA         string        `json:"git_sha"`
	ContentSHA256  string        `json:"content_sha256"`
	Files          []UpgradeFile `json:"files"`
	ReviewRequired []string      `json:"review_required"`
	Summary        string        `json:"summary,omitempty"`
}

func upgradeRecord(root string) (UpgradeRecord, error) {
	var u UpgradeRecord
	err := readJSON(filepath.Join(root, ".ai-flow", "state", "upgrade.json"), &u)
	if err == nil {
		if !strings.HasPrefix(u.ID, "upgrade-") || filepath.Base(u.ID) != u.ID || strings.ContainsAny(u.ID, "/\\") {
			return u, errors.New("invalid upgrade identifier")
		}
		for _, f := range u.Files {
			if !strings.HasPrefix(f.Path, ".ai-flow/") || strings.Contains(f.Path, "..") || strings.Contains(f.Path, "\\") || !contains([]string{"preserve", "replace", "archive"}, f.Action) {
				return u, errors.New("unsafe upgrade path or action")
			}
			if e := ensurePathInsideRepository(root, f.Path); e != nil {
				return u, e
			}
		}
	}
	return u, err
}

func requireProjectCompatible(root string) error {
	u, err := upgradeRecord(root)
	if err == nil && !contains([]string{"completed", "restored"}, u.Status) {
		return errors.New("project records are being upgraded; finish the upgrade review before continuing development")
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	m, err := readFlatYAML(filepath.Join(root, ".ai-flow", "manifest.yaml"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if m["pack_version"] != "" && m["pack_version"] != packVersion {
		return fmt.Errorf("project records use %s; run project upgrade before continuing with %s", m["pack_version"], packVersion)
	}
	return nil
}

func runProjectUpgrade(args []string) error {
	fs := flag.NewFlagSet("project upgrade", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	mode := fs.String("mode", "check", "check, prepare, apply, finish, restore")
	schemas := fs.String("schemas", "", "new runtime schema directory (installer)")
	summary := fs.String("summary", "", "reconciled requirements, preserved plan and next action")
	var resolved stringListFlag
	fs.Var(&resolved, "resolved", "reviewed source path; repeat for all recovery gaps")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, false)
	if err != nil {
		return err
	}
	if *mode == "check" {
		err := requireProjectCompatible(root)
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Println("项目记录与当前工具兼容，可以继续原计划。")
		return nil
	}
	unlock, err := projectMutationLock(root)
	if err != nil {
		return err
	}
	defer unlock()
	switch *mode {
	case "prepare":
		return prepareUpgrade(root, *schemas)
	case "apply":
		return applyUpgrade(root)
	case "finish":
		return finishUpgrade(root, *summary, resolved)
	case "restore":
		return restoreUpgrade(root)
	default:
		return errors.New("unknown upgrade mode")
	}
}

func prepareUpgrade(root, schemaRoot string) error {
	manifest, err := readFlatYAML(filepath.Join(root, ".ai-flow", "manifest.yaml"))
	if os.IsNotExist(err) {
		fmt.Println("尚未初始化，无项目记录需要升级。")
		return nil
	}
	if err != nil {
		return err
	}
	if u, err := upgradeRecord(root); err == nil && !contains([]string{"completed", "restored"}, u.Status) {
		return printJSON(u)
	}
	if manifest["pack_version"] == packVersion {
		fmt.Println("项目记录已经是当前版本。")
		return nil
	}
	if schemaRoot == "" {
		schemaRoot, err = findSchemaRoot(root)
		if err != nil {
			return err
		}
	}
	compiled, err := compileSchemas(schemaRoot)
	if err != nil {
		return err
	}
	fingerprint, err := verificationFingerprint(root)
	if err != nil {
		return err
	}
	u := UpgradeRecord{ID: "upgrade-" + time.Now().UTC().Format("20060102T150405.000000000"), From: manifest["pack_version"], To: packVersion, Status: "prepared", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), GitSHA: gitSHA(root), ContentSHA256: fingerprint, Files: []UpgradeFile{}, ReviewRequired: []string{}}
	base := filepath.Join(root, ".ai-flow", "archive", "upgrades", u.ID)
	paths := []string{}
	err = filepath.Walk(filepath.Join(root, ".ai-flow"), func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}
		rel := relativeDisplay(root, path)
		if info.IsDir() {
			for _, prefix := range []string{".ai-flow/archive/upgrades", ".ai-flow/upgrades", ".ai-flow/bin", ".ai-flow/runtime", ".ai-flow/install", ".ai-flow/locks"} {
				if rel == prefix {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("review symlink before migration: %s", rel)
		}
		if rel == ".ai-flow/state/upgrade.json" {
			return nil
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Strings(paths)
	for _, rel := range paths {
		path := filepath.Join(root, filepath.FromSlash(rel))
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		digest, err := sha256File(path)
		if err != nil {
			return err
		}
		entry := UpgradeFile{Path: rel, SHA256: digest, Action: "preserve"}
		if err := writeBytesAtomic(filepath.Join(base, "originals", filepath.FromSlash(rel)), content); err != nil {
			return err
		}
		if schemaName := upgradeSchema(rel); schemaName != "" {
			var object map[string]any
			if json.Unmarshal(content, &object) != nil {
				entry.Action = "archive"
				entry.Reason = "unreadable record; recover its intent from the original"
				u.ReviewRequired = append(u.ReviewRequired, rel)
			} else {
				converted, notes := normalizeUpgradeObject(rel, object, schemaRoot)
				encoded, _ := json.MarshalIndent(converted, "", "  ")
				if schema := compiled[schemaName]; schema == nil {
					return fmt.Errorf("missing new schema %s", schemaName)
				} else if err := schema.Validate(converted); err != nil {
					entry.Action = "archive"
					entry.Reason = "record cannot be safely converted; restore its requirements from the preserved original"
					u.ReviewRequired = append(u.ReviewRequired, rel)
				} else {
					entry.Action = "replace"
					entry.Reason = strings.Join(notes, "; ")
					if len(notes) > 0 {
						u.ReviewRequired = append(u.ReviewRequired, rel)
					}
					if err := writeBytesAtomic(filepath.Join(root, ".ai-flow", "upgrades", u.ID, "staged", filepath.FromSlash(rel)), encoded); err != nil {
						return err
					}
				}
			}
		} else if strings.HasPrefix(rel, ".ai-flow/reports/") || strings.HasPrefix(rel, ".ai-flow/prototypes/") {
			entry.Action = "archive"
			entry.Reason = "process material retained as historical source; extract still-current requirements before resuming"
			u.ReviewRequired = append(u.ReviewRequired, rel)
		}
		u.Files = append(u.Files, entry)
	}
	if err := writeJSONAtomic(filepath.Join(base, "inventory.json"), u); err != nil {
		return err
	}
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow", "state", "upgrade.json"), u); err != nil {
		return err
	}
	fmt.Printf("升级前资料已备份。原计划与任务身份保留；需要核对 %d 份材料。\n", len(u.ReviewRequired))
	return nil
}

func applyUpgrade(root string) error {
	u, err := upgradeRecord(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if u.Status == "completed" || u.Status == "needs_review" {
		return printJSON(u)
	}
	if u.Status != "prepared" && u.Status != "applying" {
		return errors.New("prepare an upgrade first")
	}
	if u.To != packVersion {
		return errors.New("upgrade proposal belongs to a different runtime version")
	}
	if u.Status == "prepared" {
		fingerprint, err := verificationFingerprint(root)
		if err != nil {
			return err
		}
		if fingerprint != u.ContentSHA256 || gitSHA(root) != u.GitSHA {
			return errors.New("project code changed after upgrade preparation; preserve the new work and prepare a fresh review")
		}
		for _, file := range u.Files {
			digest, err := sha256File(filepath.Join(root, filepath.FromSlash(file.Path)))
			if err != nil {
				return err
			}
			if digest != file.SHA256 {
				return fmt.Errorf("project record changed after upgrade preparation: %s", file.Path)
			}
		}
	}
	for _, file := range u.Files {
		if file.Action == "preserve" {
			continue
		}
		path := filepath.Join(root, filepath.FromSlash(file.Path))
		digest, currentErr := sha256File(path)
		if os.IsNotExist(currentErr) && file.Action == "archive" && u.Status == "applying" {
			continue
		}
		if currentErr != nil {
			return currentErr
		}
		if digest == file.SHA256 {
			continue
		}
		stagedDigest, stagedErr := sha256File(filepath.Join(root, ".ai-flow/upgrades", u.ID, "staged", filepath.FromSlash(file.Path)))
		if file.Action != "replace" || stagedErr != nil || digest != stagedDigest {
			return fmt.Errorf("record changed during upgrade; retain it and reconcile before retrying: %s", file.Path)
		}
	}
	u.Status = "applying"
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/state/upgrade.json"), u); err != nil {
		return err
	}
	for _, file := range u.Files {
		path := filepath.Join(root, filepath.FromSlash(file.Path))
		switch file.Action {
		case "replace":
			content, err := os.ReadFile(filepath.Join(root, ".ai-flow/upgrades", u.ID, "staged", filepath.FromSlash(file.Path)))
			if err != nil {
				return err
			}
			if err := writeBytesAtomic(path, content); err != nil {
				return err
			}
		case "archive":
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
	}
	if err := scanUpgradeEngineering(root); err != nil {
		return err
	}
	u.Status = "needs_review"
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/state/upgrade.json"), u); err != nil {
		return err
	}
	return writeUpgradeGuide(root, u)
}

func finishUpgrade(root, summary string, resolved []string) error {
	u, err := upgradeRecord(root)
	if err != nil {
		return err
	}
	if u.Status == "completed" {
		return nil
	}
	if u.Status != "needs_review" && u.Status != "finishing" {
		return errors.New("apply and review the upgrade first")
	}
	if strings.TrimSpace(summary) == "" {
		return errors.New("record the extracted requirements, retained plan, remaining questions and next action")
	}
	for _, path := range u.ReviewRequired {
		if !contains(resolved, path) {
			return fmt.Errorf("review the preserved source and reconcile its current requirements first: %s", path)
		}
	}
	if err := runValidate([]string{"--root", root, "--machine-only"}); err != nil {
		return fmt.Errorf("reconcile missing or invalid project links before finishing: %w", err)
	}
	if err := renderBoard(root); err != nil {
		return err
	}
	if err := runValidate([]string{"--root", root}); err != nil {
		return err
	}
	u.Status = "finishing"
	u.Summary = summary
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/state/upgrade.json"), u); err != nil {
		return err
	}
	manifest, err := os.ReadFile(filepath.Join(root, ".ai-flow/manifest.yaml"))
	if err != nil {
		return err
	}
	lines := strings.Split(string(manifest), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(line, "pack_version:") {
			lines[i] = "pack_version: " + packVersion
			found = true
		}
	}
	if !found {
		lines = append(lines, "pack_version: "+packVersion)
	}
	if err := writeBytesAtomic(filepath.Join(root, ".ai-flow/manifest.yaml"), []byte(strings.Join(lines, "\n"))); err != nil {
		return err
	}
	if err := writeBytesAtomic(filepath.Join(root, ".ai-flow/skill-pack.lock.yaml"), []byte("schema_version: 1\nname: "+packName+"\nversion: "+packVersion+"\nsource: installed\n")); err != nil {
		return err
	}
	u.Status = "completed"
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/state/upgrade.json"), u); err != nil {
		return err
	}
	return writeUpgradeGuide(root, u)
}

func restoreUpgrade(root string) error {
	u, err := upgradeRecord(root)
	if err != nil {
		return err
	}
	if u.Status == "restored" {
		return nil
	}
	if u.Status == "completed" {
		return errors.New("completed upgrade may have new development; use the preserved path map to review restoration instead of overwriting it")
	}
	base := filepath.Join(root, ".ai-flow/archive/upgrades", u.ID, "originals")
	// Validate every original before restoring any file.
	for _, file := range u.Files {
		digest, err := sha256File(filepath.Join(base, filepath.FromSlash(file.Path)))
		if err != nil {
			return err
		}
		if digest != file.SHA256 {
			return errors.New("backup integrity mismatch")
		}
	}

	for _, file := range u.Files {
		content, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(file.Path)))
		if err != nil {
			return err
		}
		digest, err := sha256File(filepath.Join(base, filepath.FromSlash(file.Path)))
		if err != nil {
			return err
		}
		if digest != file.SHA256 {
			return errors.New("backup integrity mismatch")
		}
		currentPath := filepath.Join(root, filepath.FromSlash(file.Path))
		if current, readErr := os.ReadFile(currentPath); readErr == nil && !sameBytes(current, content) {
			if err := writeBytesAtomic(filepath.Join(root, ".ai-flow/archive/upgrades", u.ID, "restore-displaced", filepath.FromSlash(file.Path)), current); err != nil {
				return err
			}
		}
		if err := writeBytesAtomic(filepath.Join(root, filepath.FromSlash(file.Path)), content); err != nil {
			return err
		}
	}
	u.Status = "restored"
	return writeJSONAtomic(filepath.Join(root, ".ai-flow/state/upgrade.json"), u)
}

func writeUpgradeGuide(root string, u UpgradeRecord) error {
	var out strings.Builder
	fmt.Fprintf(&out, "# 升级后继续开发\n\n原版本：%s；目标版本：%s。\n\n", u.From, u.To)
	if u.Status == "completed" {
		out.WriteString("升级检查已完成。继续原计划：" + u.Summary + "\n")
	} else {
		out.WriteString("原计划、任务、已完成进度和待确认事项已保留。先核对下列资料，提取仍然有效的需求、约束和决策，补齐新记录；不要把旧的测试通过当成当前代码已经验证。\n\n")
	}
	for _, path := range u.ReviewRequired {
		fmt.Fprintf(&out, "- [%s](../../archive/upgrades/%s/originals/%s)\n", filepath.Base(path), u.ID, path)
	}
	out.WriteString("\n资料升级不会授权发布，也不会自动批准之前尚未确认的选择。非本工具管理的旧文档只做盘点，整理前仍需精确路径清单。\n")
	return writeBytesAtomic(filepath.Join(root, ".ai-flow/upgrades", u.ID, "CONTINUE.md"), []byte(out.String()))
}
