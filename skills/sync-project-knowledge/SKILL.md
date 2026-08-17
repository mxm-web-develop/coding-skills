---
name: sync-project-knowledge
description: Synchronize AI Flow machine state, concise human dashboards, reports, release history, and superseded-document archives. Use after any workflow state changes, when dashboards or documentation are stale, when old plans cause confusion, or when a user asks for project reports or current progress.
---

# Sync Project Knowledge

Keep one current machine truth and generate human views from it.

## Procedure

1. Read current machine state, events, active objects, releases, evidence, workspace document inventory, and archive index.
2. Validate IDs, revisions, links, statuses, and `supersedes` relationships.
3. Mark replaced objects `superseded` and add reciprocal replacement links.
4. Move superseded AI Flow snapshots to `.ai-flow/archive/<type>/<version>/` and remove them from active indexes. For pre-AI-Flow documents, apply only mappings already approved in `workspace-document-inventory.json` and follow the [legacy document cleanup contract](../adopt-existing-project/references/document-cleanup-contract.md).
5. Update current state using expected revision; stop and re-read on a conflict.
6. Run `.ai-flow/bin/flowctl validate --root <root>` and resolve Schema or link errors.
7. Run `.ai-flow/bin/flowctl render-board --root <root>`.
8. Verify that `docs/board/` reflects the machine state and contains no untracked claims. Summarize legacy history by version without expanding archived prose back into active context.
9. Verify every applied legacy-document mapping against its recorded source/target hash and retain the recovery map.
10. Report what changed, what was archived, what stayed protected, and any unresolved documentation conflict.

## Guardrails

- Do not hand-edit generated board facts when the machine state is wrong; correct the source object first.
- Do not delete historical decisions or evidence.
- Do not discover and archive arbitrary workspace documents during synchronization; require the separately reviewed adoption inventory.
- Do not archive an object without a replacement link or explicit rejected/cancelled reason.
- Keep the status board concise; link to IDs instead of copying full reports.

Read [references/knowledge-contract.md](references/knowledge-contract.md) before synchronizing.
