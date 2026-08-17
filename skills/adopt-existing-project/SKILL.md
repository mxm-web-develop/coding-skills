---
name: adopt-existing-project
description: Build an evidence-based AI Flow baseline for an existing, possibly multilingual or multi-project codebase; optionally inventory and version-archive legacy documents while only marking non-document cleanup candidates. Use when onboarding a mature repository, when current features or versions are unclear, when old documents conflict, or before planning work in a project that already contains code, Git history, tests, CI, or release artifacts.
---

# Adopt Existing Project

Establish what exists before changing what should exist.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). Describe the project as the user experiences it; do not present baseline classifications, inventory fields, modes, or Skill handoffs as user choices.

## Procedure

1. Inspect Git root, branch, tags, recent commits, remotes, and uncommitted changes.
2. Map code packages, entry points, dependencies, build commands, tests, CI, deployment files, version sources, and existing documentation. Detect monorepo/workspace manifests, subprojects, nested repositories, shared roots, languages, package managers, build graphs, generated roots, and deployment units.
3. Run only safe read-only commands and existing non-destructive discovery checks.
4. Classify every statement as:
   - `observed`: directly supported by code, command output, or Git.
   - `inferred`: a reasoned interpretation that needs confirmation.
   - `user-confirmed`: explicitly accepted by the user.
5. Identify conflicting documentation, stale plans, duplicates, failing checks, unknown ownership, and version ambiguity.
6. When documentation mode is not `keep`, follow [references/document-cleanup-contract.md](references/document-cleanup-contract.md): inventory and classify documents, derive version buckets from evidence, and write a proposed `.ai-flow/baseline/workspace-document-inventory.json` without moving anything.
7. Follow [references/workspace-inventory-contract.md](references/workspace-inventory-contract.md) and write `.ai-flow/baseline/workspace-structure-inventory.json` for non-document structure and possible cleanup candidates. Every candidate must use `initialization_action: mark-only`; do not create executable cleanup mappings during adoption.
8. Present a compact natural-language account of the current version, working features, test situation, conflicting or uncertain documents, and suggested document moves. Show each proposed source and destination in understandable terms. Ask whether to apply exactly that list; retain mode, revision, hashes, and classifications internally.
9. Apply only approved document mappings, verify target hashes, update the document inventory to `applied` or `partial`, and preserve a reversible path map. Never delete historical content.
10. Write the confirmed current baseline under `.ai-flow/baseline/`, update current state, and invoke `sync-project-knowledge`. Never handwrite or patch files in `docs/board/`.
11. Hand off to `profile-project-engineering`, then `discover-product-goal` and `plan-product-delivery`. Base new requirements on the user-confirmed current version, implemented capabilities, unresolved risks, and archived-history summary.

## Guardrails

- Do not modify business code during adoption.
- Do not move or delete non-document candidates during adoption, even when their names imply `old`, `legacy`, `backup`, `new`, or a version. They require a later explicit request routed to `clean-project-workspace` and fresh dependency evidence.
- Do not enter, classify, or clean nested repositories as though they belonged to the parent repository.
- Preserve uncommitted user work.
- Do not move a file because it merely looks old. Require version/status evidence and user approval.
- Never archive operational, legal, security, current-authority, generated, vendored, or tool-referenced documents by default.
- Do not call an old document current merely because it is prominent.
- Prefer Git tags and executable configuration over narrative claims, while recording disagreements.

Read [references/baseline-contract.md](references/baseline-contract.md) before writing the baseline.
