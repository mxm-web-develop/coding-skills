---
name: adopt-existing-project
description: Build an evidence-based AI Flow baseline for an existing codebase and optionally inventory, summarize, and version-archive scattered legacy documents. Use when onboarding a mature or documentation-heavy repository, when current features or versions are unclear, when old documents conflict, or before planning new work in a project that already contains code, Git history, tests, CI, or release artifacts.
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
5. Identify conflicting documentation, stale plans, duplicates, failing checks, unknown ownership, and version ambiguity.
6. When documentation mode is not `keep`, follow [references/document-cleanup-contract.md](references/document-cleanup-contract.md): inventory and classify documents, derive version buckets from evidence, and write a proposed `.ai-flow/baseline/workspace-document-inventory.json` without moving anything.
7. Present a compact baseline and cleanup proposal for confirmation. For `summarize-and-archive`, require explicit approval of the exact source-to-target mappings; unapproved, unknown, active, or protected files stay in place.
8. Apply only approved mappings, verify target hashes, update the inventory to `applied` or `partial`, and preserve a reversible path map. Never delete historical content.
9. Write the confirmed current baseline under `.ai-flow/baseline/`, update current state, and invoke `sync-project-knowledge`.
10. Hand off to `profile-project-engineering`, then `discover-product-goal` and `plan-product-delivery`. Base new requirements on the user-confirmed current version, implemented capabilities, unresolved risks, and archived-history summary.

## Guardrails

- Do not modify business code during adoption.
- Preserve uncommitted user work.
- Do not move a file because it merely looks old. Require version/status evidence and user approval.
- Never archive operational, legal, security, current-authority, generated, vendored, or tool-referenced documents by default.
- Do not call an old document current merely because it is prominent.
- Prefer Git tags and executable configuration over narrative claims, while recording disagreements.

Read [references/baseline-contract.md](references/baseline-contract.md) before writing the baseline.
