package main

import (
	"os"
	"path/filepath"
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
