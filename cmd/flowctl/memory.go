package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func hashBytes(data []byte) string { s := sha256.Sum256(data); return hex.EncodeToString(s[:]) }
func hashJSON(v any) (string, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	return hashBytes(b), nil
}

type ProjectFact struct {
	Key         string           `json:"key"`
	Claim       string           `json:"claim"`
	WorkID      string           `json:"work_id"`
	ConfirmedBy string           `json:"confirmed_by"`
	ConfirmedAt string           `json:"confirmed_at"`
	Sources     []SourceSnapshot `json:"sources"`
	Supersedes  string           `json:"supersedes,omitempty"`
}

// Append-only versions keep the basis of an old decision recoverable.
type FactStore struct {
	Facts []ProjectFact `json:"facts"`
}

func loadFacts(root string) (FactStore, error) {
	s := FactStore{Facts: []ProjectFact{}}
	err := readJSON(filepath.Join(root, ".ai-flow/baseline/verified-facts.json"), &s)
	if os.IsNotExist(err) {
		err = nil
	}
	return s, err
}
func verifiedFact(root, key string) (ProjectFact, error) {
	s, err := loadFacts(root)
	if err != nil {
		return ProjectFact{}, err
	}
	for i := len(s.Facts) - 1; i >= 0; i-- {
		f := s.Facts[i]
		if f.Key == key {
			if err := checkSnapshots(root, f.Sources); err != nil {
				return f, fmt.Errorf("project fact %s is stale: %w", key, err)
			}
			return f, nil
		}
	}
	return ProjectFact{}, fmt.Errorf("project fact is not confirmed: %s", key)
}
func runMemory(args []string) error {
	if len(args) > 0 && args[0] == "confirm-profile" {
		return runProfileConfirmation(args[1:])
	}
	if len(args) == 0 {
		return errors.New("usage: memory record|check|scan --root PATH")
	}
	fs := flag.NewFlagSet("memory "+args[0], flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	work := fs.String("work", "", "originating task")
	key := fs.String("key", "", "stable fact key")
	claim := fs.String("claim", "", "observed fact; not a hypothesis")
	by := fs.String("by", "", "reviewer who inspected the sources")
	supersedes := fs.String("supersedes", "", "hash of the previous version when changing a claim")
	var sources stringListFlag
	fs.Var(&sources, "source", "supporting repository file; repeatable")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	if args[0] == "check" {
		s, err := loadFacts(root)
		if err != nil {
			return err
		}
		keys := []string{}
		if *key != "" {
			keys = append(keys, *key)
		} else {
			for _, f := range s.Facts {
				keys = uniqueAppend(keys, f.Key)
			}
		}
		result := []map[string]any{}
		var stale error
		for _, k := range keys {
			f, e := verifiedFact(root, k)
			digest, _ := hashJSON(f)
			status := "confirmed"
			if e != nil {
				status = "stale"
				stale = e
			}
			result = append(result, map[string]any{"fact": f, "sha256": digest, "status": status})
		}
		if err := printJSON(result); err != nil {
			return err
		}
		return stale
	}
	if args[0] != "record" && args[0] != "scan" {
		return errors.New("unknown memory action")
	}
	if _, err := readWorkItem(root, *work); err != nil {
		return err
	}
	unlock, err := projectMutationLock(root)
	if err != nil {
		return err
	}
	defer unlock()
	if args[0] == "scan" {
		return scanUpgradeEngineering(root)
	}
	if strings.TrimSpace(*key) == "" || strings.TrimSpace(*claim) == "" || strings.TrimSpace(*by) == "" || len(sources) == 0 {
		return errors.New("fact key, claim, reviewer and supporting sources are required")
	}
	snapshots, err := snapshotSources(root, sources)
	if err != nil {
		return err
	}
	s, err := loadFacts(root)
	if err != nil {
		return err
	}
	for i := len(s.Facts) - 1; i >= 0; i-- {
		if s.Facts[i].Key == *key {
			digest, e := hashJSON(s.Facts[i])
			if e != nil {
				return e
			}
			if *supersedes != digest {
				return errors.New("a prior fact exists; inspect memory check and explicitly supersede its hash (including stale facts)")
			}
			break
		}
	}
	f := ProjectFact{Key: *key, Claim: *claim, WorkID: *work, ConfirmedBy: *by, ConfirmedAt: time.Now().UTC().Format(time.RFC3339Nano), Sources: snapshots, Supersedes: *supersedes}
	s.Facts = append(s.Facts, f)
	if err := writeJSONAtomic(filepath.Join(root, ".ai-flow/baseline/verified-facts.json"), s); err != nil {
		return err
	}
	digest, _ := hashJSON(f)
	return printJSON(map[string]any{"fact": f, "sha256": digest})
}
