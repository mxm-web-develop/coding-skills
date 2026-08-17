package main

import (
	"path/filepath"
	"testing"
)

func TestCompileAllSchemas(t *testing.T) {
	schemaRoot := filepath.Join("..", "..", "schemas")
	compiled, err := compileSchemas(schemaRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(compiled) != 12 {
		t.Fatalf("compiled %d schemas, want 12 object schemas plus common definitions", len(compiled))
	}
	for _, required := range []string{
		"work-item.schema.json",
		"run.schema.json",
		"checkpoint.schema.json",
		"evidence.schema.json",
		"event.schema.json",
	} {
		if compiled[required] == nil {
			t.Fatalf("missing compiled schema %s", required)
		}
	}
}

func TestObjectIDFormat(t *testing.T) {
	id, err := newObjectID("WI")
	if err != nil {
		t.Fatal(err)
	}
	if err := requireObjectID(id, "WI"); err != nil {
		t.Fatalf("generated invalid id %s: %v", id, err)
	}
	if err := requireObjectID(id, "EV"); err == nil {
		t.Fatal("accepted an ID with the wrong prefix")
	}
}
