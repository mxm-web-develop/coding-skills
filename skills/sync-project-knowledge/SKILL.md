---
name: sync-project-knowledge
description: Synchronize AI Flow machine state and concise natural-language dashboards covering major/minor versions, development tasks, architecture decisions, tests, releases, workspace cleanup, and approved archives. Use after workflow state changes, when dashboards or documentation are stale, when old plans cause confusion, or when a user asks for project status, reports, decisions, test results, cleanup results, or current progress.
---

# Sync Project Knowledge

Keep one current machine truth and generate human views from it.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md) for both conversation and generated boards. Never make users learn AI Flow's object names, IDs, directories, state values, or abbreviations to understand project status.

## Procedure

1. Read current machine state, events, active objects, releases, evidence, workspace document inventory, workspace structure inventory, workspace cleanup plans, and archive index.
2. Validate IDs, revisions, links, statuses, and `supersedes` relationships.
3. Mark replaced objects `superseded` and add reciprocal replacement links.
4. Move superseded AI Flow snapshots to `.ai-flow/archive/<type>/<version>/` and remove them from active indexes. For pre-AI-Flow documents, apply only mappings already approved in `workspace-document-inventory.json` and follow the [legacy document cleanup contract](../adopt-existing-project/references/document-cleanup-contract.md).
   Treat legacy code/files as already authorized only when an approved `workspace-cleanup/PLAN-*.json` records the exact path, fingerprint, target or removal action, recovery method, and verified result. Synchronization never invents cleanup actions.
5. Update current state using expected revision; stop and re-read on a conflict.
6. Run `.ai-flow/bin/flowctl validate --root <root>` and resolve Schema or link errors.
7. Read [references/board-contract.md](references/board-contract.md), then run `.ai-flow/bin/flowctl render-board --root <root>`.
8. Verify that the four files in `docs/board/` use natural-language summaries and compact tables for versions, tasks, decisions, tests, and releases while remaining fully traceable to machine objects. Summarize legacy history by version without expanding archived prose back into active context.
9. Verify every applied legacy-document and workspace-cleanup mapping against its recorded fingerprint and retain the recovery map.
10. Report what changed, what was archived or removed, how it can be restored, what stayed protected, and any unresolved documentation or cleanup conflict.

## Guardrails

- Do not hand-edit generated board facts when the machine state is wrong; correct the source object first.
- Do not delete historical decisions or evidence.
- Do not discover and archive arbitrary workspace documents during synchronization; require the separately reviewed adoption inventory.
- Do not execute, expand, or approve workspace cleanup during synchronization; require the explicit `clean-project-workspace` workflow.
- Do not archive an object without a replacement link or explicit rejected/cancelled reason.
- Keep the status board concise; link to IDs instead of copying full reports.
- Never expose raw status codes as the primary user explanation when a plain-language label is available.

Read [references/knowledge-contract.md](references/knowledge-contract.md) before synchronizing.
