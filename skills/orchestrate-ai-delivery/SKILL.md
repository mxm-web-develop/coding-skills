---
name: orchestrate-ai-delivery
description: Route and coordinate AI Flow work for an initialized project. Use for project status, versions, plans, feature work, bug fixes, refactors, tests, reviews, releases, documentation updates, long-running tasks, or any software change in a repository containing `.ai-flow/manifest.yaml`.
---

# Orchestrate AI Delivery

Treat repository state as the source of truth and dispatch the smallest complete workflow.

## Entry sequence

1. Find the project root and read `.ai-flow/manifest.yaml`.
2. If it is absent, invoke `initialize-ai-project`.
3. Run `.ai-flow/bin/flowctl status --root <root> --json` and inspect Git status.
4. Classify the request using [references/routing.md](references/routing.md).
5. Bind every mutation to an existing Goal/Requirement/Work Item. When none exists, create the smallest traceable item with `flowctl work create`.
6. Dispatch one Skill at a time. Require its declared output before continuing.
7. Use `flowctl checkpoint save` before a phase transition, external wait, user pause, context compaction, or retry.
8. Stop at an approval gate rather than assuming permission to push, merge, tag, deploy, delete, or publish.

## Coordination rules

- Answer read-only status questions by reading machine state; do not start the full delivery chain.
- Keep one active writing owner per Work Item and scope. Serialize overlapping changes.
- Start writing with `flowctl work start`; resume interrupted work with `flowctl checkpoint latest` and `flowctl checkpoint resume`.
- On verification failure, return to diagnosis or design; never mark work complete from prose alone.
- Prefer a recoverable next action over a long speculative plan.
- Finish by invoking `sync-project-knowledge` whenever machine state changed.

## Completion response

Report the current version, active goal, completed change, verification evidence, unresolved risks, next work item, and whether approval is required.
