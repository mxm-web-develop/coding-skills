---
name: adopt-existing-project
description: Build an evidence-based AI Flow baseline for an existing codebase. Use when onboarding a mature repository, when current features or versions are unclear, or before planning new work in a project that already contains code, Git history, tests, CI, or release artifacts.
---

# Adopt Existing Project

Establish what exists before changing what should exist.

## Procedure

1. Inspect Git root, branch, tags, recent commits, remotes, and uncommitted changes.
2. Map code packages, entry points, dependencies, build commands, tests, CI, deployment files, version sources, and existing documentation.
3. Run only safe read-only commands and existing non-destructive discovery checks.
4. Classify every statement as:
   - `observed`: directly supported by code, command output, or Git.
   - `inferred`: a reasoned interpretation that needs confirmation.
   - `user-confirmed`: explicitly accepted by the user.
5. Identify conflicting documentation, stale plans, failing checks, unknown ownership, and version ambiguity.
6. Present a compact baseline for confirmation. Do not silently resolve material ambiguity.
7. After confirmation, write the baseline under `.ai-flow/baseline/`, update current state, and invoke `sync-project-knowledge`.

## Guardrails

- Do not modify business code during adoption.
- Preserve uncommitted user work.
- Do not call an old document current merely because it is prominent.
- Prefer Git tags and executable configuration over narrative claims, while recording disagreements.

Read [references/baseline-contract.md](references/baseline-contract.md) before writing the baseline.
