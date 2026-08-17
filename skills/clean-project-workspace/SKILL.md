---
name: clean-project-workspace
description: Audit and safely organize code, directories, generated artifacts, caches, and other non-document workspace content in an initialized AI Flow project. Use only after `.ai-flow/manifest.yaml` exists and the user explicitly asks to clean, organize, remove obsolete code, archive deprecated modules, or tidy the project workspace. Supports multilingual repositories and monorepos; never treat initialization-time candidate labels as cleanup approval.
---

# Clean Project Workspace

Turn an explicit cleanup request into an evidence-backed, reversible change without confusing historical candidates with proven dead code.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). The user sees proposed paths, effects, risks, tests, and recovery options—not plan IDs, revisions, digests, fingerprints, classifications, or internal state names.

## Procedure

1. Confirm `.ai-flow/manifest.yaml` exists. If it does not, invoke `initialize-ai-project`; initialization may mark non-document candidates but must not clean them.
2. Read the current Git status, `.ai-flow/baseline/workspace-structure-inventory.json`, engineering profile, active Work Items, accepted decisions, CI, deployment configuration, and current human boards.
3. Create the smallest cleanup Work Item when none exists. Do not reuse an unrelated feature or bug Work Item.
4. Re-scan the repository using [references/repository-discovery.md](references/repository-discovery.md). Treat each detected subproject, nested repository, language, package manager, build graph, deployment unit, and shared root as an independent evidence boundary.
5. Revalidate every initialization-time candidate against the current Git SHA. Add newly discovered candidates, but never promote a candidate to confirmed-deprecated from its name, age, directory suffix, or lack of obvious imports alone.
6. Follow [references/cleanup-contract.md](references/cleanup-contract.md), write a proposed `.ai-flow/workspace-cleanup/PLAN-<id>.json`, and calculate its stable approval digest with `flowctl cleanup digest`. The proposal must contain exact paths, fingerprints, owning components, dependency evidence, action, risk, verification commands, recovery method, and execution batches. Do not mutate the workspace yet.
7. Present a concise proposal grouped by part of the project and intended effect: content to preserve as history, safe-to-recreate output to remove, ignore rules to improve, and content that will remain untouched. For each changing path, explain why, risk, verification, and recovery in ordinary language. Ask whether to execute exactly the displayed list; bind that answer internally to the plan revision and digest.
8. After approval, save progress and verify internally that the project and approved list are unchanged. If anything drifted, say that the project changed since confirmation and a fresh safety check is required; do not ask the user to interpret hashes or revision numbers.
9. Execute one dependency-aware batch at a time. Prefer `git mv` for tracked historical content. Never cross a nested repository boundary, modify secrets, or remove generated content without the declared recovery method.
10. Run the batch-specific tests, builds, type checks, packaging checks, and deployment/configuration validation selected for every affected component. On failure, stop, restore the batch, record evidence, and return to diagnosis.
11. Mark results in the plan, record Evidence, refresh the workspace structure inventory and engineering profile, then invoke `sync-project-knowledge`. Report what moved or was removed, how it can be restored, and what remains uncertain.

## Guardrails

- Require an explicit post-initialization cleanup request. Initialization, adoption, candidate labels, or a general documentation archive preference never authorizes code or directory mutation.
- Default `active`, `protected`, `unknown`, cross-component shared, nested-repository, migration, infrastructure, security, legal, and secret-bearing paths to `keep`.
- Treat a monorepo as a graph, not a folder list. A path is unused only after checking all consumers, build/test/package tasks, CI, deployment, runtime configuration, code generation, and plugin discovery.
- Separate source history from disposable outputs. Archive confirmed historical source; remove generated/cache content only when reproducible or recoverable; never archive dependency/vendor trees merely to make the root look clean.
- Preserve uncommitted work. Any overlapping change blocks the affected batch until the user resolves or explicitly scopes it.
- Never delete Git history, rewrite commits, clean an entire repository recursively, or infer approval for a parent directory from approval of a child path.
- Do not mark the cleanup Work Item complete until verification evidence matches the final Git revision.
