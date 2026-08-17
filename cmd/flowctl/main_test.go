package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadFlatYAML(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "state.yaml")
	content := "schema_version: 1\nname: \"Demo Project\"\nphase: goal_alignment\n  nested: ignored\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	values, err := readFlatYAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if values["name"] != "Demo Project" {
		t.Fatalf("unexpected name: %q", values["name"])
	}
	if _, exists := values["nested"]; exists {
		t.Fatal("nested YAML key should not be read as a top-level scalar")
	}
}

func TestYAMLScalar(t *testing.T) {
	got := yamlScalar("A \"quoted\" project")
	want := "\"A \\\"quoted\\\" project\""
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNextActionForMode(t *testing.T) {
	if got := nextActionForMode("existing"); got != "adopt_existing_project" {
		t.Fatalf("unexpected action: %s", got)
	}
	if got := nextActionForMode("greenfield"); got != "discover_product_goal" {
		t.Fatalf("unexpected action: %s", got)
	}
}

func TestInstalledPlatformsReadsSelectionAndWindowsBOM(t *testing.T) {
	root := t.TempDir()
	installDir := filepath.Join(root, ".ai-flow", "install")
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installDir, "platforms"), []byte("\ufeffcursor\r\ncodex\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := []string{"codex", "cursor"}
	if got := installedPlatforms(root); !reflect.DeepEqual(got, want) {
		t.Fatalf("installedPlatforms() = %#v, want %#v", got, want)
	}
}
