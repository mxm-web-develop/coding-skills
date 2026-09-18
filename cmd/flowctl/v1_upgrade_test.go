package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func legacyFixture(t *testing.T) (string, WorkItem, HarnessRun) {
	root, w, r := v1Fixture(t)
	writeBoardTextFixture(t, root, ".ai-flow/manifest.yaml", "schema_version: 1\npack_name: mxm-ai-flow\npack_version: 0.4.5\n")
	return root, w, r
}

func TestUpgradePreservesTaskAndRequiresHistoricalExtraction(t *testing.T) {
	root, w, r := legacyFixture(t)
	original := []byte("# 原计划补充\n用户尚未确认自动发送报告；先完成导出。\n")
	writeBoardTextFixture(t, root, ".ai-flow/reports/old-plan.md", string(original))
	if err := prepareUpgrade(root, filepath.Join(root, "schemas")); err != nil {
		t.Fatal(err)
	}
	u, _ := upgradeRecord(root)
	if err := applyUpgrade(root); err != nil {
		t.Fatal(err)
	}
	if err := applyUpgrade(root); err != nil {
		t.Fatal("retry", err)
	}
	item, err := readWorkItem(root, w.ID)
	if err != nil || item.RunID == nil || *item.RunID != r.ID || item.Status != w.Status {
		t.Fatal("progress was reset", item, err)
	}
	if err := requireProjectCompatible(root); err == nil {
		t.Fatal("development allowed before review")
	}
	if err := finishUpgrade(root, "继续导出", nil); err == nil {
		t.Fatal("historical material not reviewed")
	}
	backup, err := os.ReadFile(filepath.Join(root, ".ai-flow/archive/upgrades", u.ID, "originals/.ai-flow/reports/old-plan.md"))
	if err != nil || string(backup) != string(original) {
		t.Fatal("original lost")
	}
	u, _ = upgradeRecord(root)
	if err := finishUpgrade(root, "原计划继续导出；自动发送仍待用户确认", u.ReviewRequired); err != nil {
		t.Fatal(err)
	}
	if err := requireProjectCompatible(root); err != nil {
		t.Fatal(err)
	}
	if err := prepareUpgrade(root, filepath.Join(root, "schemas")); err != nil {
		t.Fatal("repeat upgrade", err)
	}
	latest, _ := upgradeRecord(root)
	if latest.ID != u.ID {
		t.Fatal("repeat upgrade duplicated migration")
	}
}

func TestUpgradeRejectsConcurrentRecordChanges(t *testing.T) {
	root, w, _ := legacyFixture(t)
	if err := prepareUpgrade(root, filepath.Join(root, "schemas")); err != nil {
		t.Fatal(err)
	}
	w.Title = "用户刚补充的新要求"
	if err := writeJSONAtomic(workItemPath(root, w.ID), w); err != nil {
		t.Fatal(err)
	}
	if err := applyUpgrade(root); err == nil {
		t.Fatal("overwrote concurrent work")
	}
	item, _ := readWorkItem(root, w.ID)
	if item.Title != w.Title {
		t.Fatal("new requirement lost")
	}
}

func TestUpgradePreservesUnknownFieldsAndRestoresRecords(t *testing.T) {
	root, w, _ := legacyFixture(t)
	path := workItemPath(root, w.ID)
	data, _ := os.ReadFile(path)
	var obj map[string]any
	_ = json.Unmarshal(data, &obj)
	obj["legacy_decision"] = "保留用户已确认的批量导出约束"
	if err := writeJSONAtomic(path, obj); err != nil {
		t.Fatal(err)
	}
	if err := prepareUpgrade(root, filepath.Join(root, "schemas")); err != nil {
		t.Fatal(err)
	}
	if err := applyUpgrade(root); err != nil {
		t.Fatal(err)
	}
	u, _ := upgradeRecord(root)
	if len(u.ReviewRequired) != 1 {
		t.Fatalf("expected explicit review for unknown field, got %v", u.ReviewRequired)
	}
	if err := restoreUpgrade(root); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if !strings.Contains(string(data), "legacy_decision") {
		t.Fatal("restore lost unknown field")
	}
	m, _ := readFlatYAML(filepath.Join(root, ".ai-flow/manifest.yaml"))
	if m["pack_version"] != "0.4.5" {
		t.Fatal("restore changed original version")
	}
}

func TestRoutingCommonChineseRequests(t *testing.T) {
	for message, want := range map[string]string{"汇报下当前开发计划": "status", "调研下消息队列方案，看看是否适合我们项目": "research-only", "添加个新功能，批量导出": "feature", "升级后继续之前的计划": "upgrade", "验收通过，继续下一项": "acceptance-feedback"} {
		route := classifyIntent(message)
		if route.Intent != want {
			t.Errorf("%s routed to %s", message, route.Intent)
		}
		if want == "research-only" && route.Implement {
			t.Fatal("research automatically implements")
		}
		if want == "status" && !route.ReadOnly {
			t.Fatal("status request mutates")
		}
	}
}

func TestUpgradeResumeAfterPartialApplyAndInterruptedFinish(t *testing.T) {
	root, _, _ := legacyFixture(t)
	writeBoardTextFixture(t, root, ".ai-flow/reports/old.md", "保留原计划")
	if err := prepareUpgrade(root, filepath.Join(root, "schemas")); err != nil {
		t.Fatal(err)
	}
	u, _ := upgradeRecord(root)
	u.Status = "applying"
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/state/upgrade.json"), u); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, ".ai-flow/reports/old.md")); err != nil {
		t.Fatal(err)
	}
	if err := applyUpgrade(root); err != nil {
		t.Fatal("partial apply could not resume", err)
	}
	u, _ = upgradeRecord(root)
	u.Status = "finishing"
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/state/upgrade.json"), u); err != nil {
		t.Fatal(err)
	}
	if err := finishUpgrade(root, "继续原计划", u.ReviewRequired); err != nil {
		t.Fatal("interrupted finish", err)
	}
	if err := requireProjectCompatible(root); err != nil {
		t.Fatal(err)
	}
}

func TestUpgradeRestorationChecksAllBackupsBeforeWriting(t *testing.T) {
	root, w, _ := legacyFixture(t)
	if err := prepareUpgrade(root, filepath.Join(root, "schemas")); err != nil {
		t.Fatal(err)
	}
	if err := applyUpgrade(root); err != nil {
		t.Fatal(err)
	}
	u, _ := upgradeRecord(root)
	path := workItemPath(root, w.ID)
	before, _ := os.ReadFile(path)
	last := u.Files[len(u.Files)-1]
	writeBoardTextFixture(t, root, filepath.Join(".ai-flow/archive/upgrades", u.ID, "originals", last.Path), "corrupt")
	if err := restoreUpgrade(root); err == nil {
		t.Fatal("corrupt backup restored")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Fatal("partial restore before integrity check")
	}
}
