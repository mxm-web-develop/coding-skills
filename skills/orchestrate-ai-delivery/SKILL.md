---
name: orchestrate-ai-delivery
description: Route and coordinate AI Flow work for an initialized project. Use for project status, versions, plans, feature work, bug fixes, refactors, workspace cleanup, tests, reviews, releases, documentation updates, long-running tasks, or any software change in a repository containing `.ai-flow/manifest.yaml`.
---

# Orchestrate AI Delivery

Treat repository state as the source of truth and dispatch the smallest complete workflow.

Before any user-facing question, progress update, choice, error, or completion response, read and follow [references/user-communication-contract.md](references/user-communication-contract.md). Internal terms remain in machine state; conversation uses the user's product and project language.

## Entry sequence

1. Find the project root and read `.ai-flow/manifest.yaml`.
2. If it is absent, invoke `initialize-ai-project`.
3. Run `.ai-flow/bin/flowctl status --root <root> --json` and inspect Git status.
4. For technical mutation, ensure `.ai-flow/baseline/engineering-profile.json` exists and is current. Invoke `profile-project-engineering` when it is absent or stale.
5. Classify the request using [references/routing.md](references/routing.md) without narrating routing or Skill names to the user.
6. Bind every mutation to an existing Goal/Requirement/Work Item. When none exists, create the smallest traceable item with `flowctl work create`.
7. Dispatch one Skill at a time. Require its declared output before continuing.
8. Use `flowctl checkpoint save` before a phase transition, external wait, user pause, context compaction, or retry.
9. Stop at an approval gate rather than assuming permission to push, merge, tag, deploy, delete, or publish.

## Coordination rules

- Answer read-only status questions by reading machine state; do not start the full delivery chain.
- Translate every delegated result through the user communication contract. Do not forward raw Skill output, object IDs, enum values, revision numbers, hashes, internal directory names, or CLI errors as the primary response.
- Route language/framework decisions through the engineering profile. Do not let an execution Skill guess a different stack or silently acquire third-party Skills.
- Treat cleanup of pre-AI-Flow documents as an adoption workflow: inventory first, obtain path-level approval, then apply only approved reversible mappings.
- Route an explicit post-initialization request to clean code, folders, generated outputs, caches, or other non-document content to `clean-project-workspace`. Initialization-time candidate labels are never sufficient evidence or approval.
- Keep one active writing owner per Work Item and scope. Serialize overlapping changes.
- Start writing with `flowctl work start`; resume interrupted work with `flowctl checkpoint latest` and `flowctl checkpoint resume`.
- On verification failure, return to diagnosis or design; never mark work complete from prose alone.
- Prefer a recoverable next action over a long speculative plan.
- Finish by invoking `sync-project-knowledge` whenever machine state changed.

## Completion response

Report the current version, what this version is trying to achieve, what changed, the test result, unresolved risks, what happens next, and whether the user needs to decide anything. Use natural labels; include internal IDs or technical records only in optional detail when useful.
