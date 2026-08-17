---
name: review-change
description: Independently review a proposed code or documentation change against its requirements, design, tests, and project policies. Use before commit, merge, or release; for explicit review requests; or when verification passed but release readiness is not yet established.
---

# Review Change

Review the actual diff and evidence, not the implementer's summary.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). Lead with whether the change is safe to continue, then explain concrete problems and impact; keep severity codes, object IDs, and workflow routing secondary unless the user asks for review detail.

## Procedure

1. Fix the review base and inspect the complete diff, untracked files, linked Work Item, Requirements, decisions, evidence, and current engineering profile.
2. Check requirement compliance and missing acceptance behavior.
3. Check correctness, error handling, compatibility, data integrity, concurrency, security, performance, and operations as applicable.
4. Check architectural fit, module/file responsibility, pure-core/effect boundaries, public API scope, generated code, whether material technology/UX choices were confirmed, and whether any decision changed without a superseding record. Flag multipurpose monoliths and comments that obscure rather than explain intent.
5. Check stack-appropriate test quality, regression coverage, selected community Skill provenance, and evidence revision freshness. Require functional and visual evidence for applicable UI changes.
6. Run focused independent checks when needed and safe.
7. Report findings by severity with precise file/line evidence and a concrete failure scenario.
8. Mark the review `approved`, `approved_with_nonblocking_findings`, or `changes_required`.
9. Route blocking findings back to the appropriate Skill. On approval, use `flowctl work review-ready` and route to `integrate-git-change`.

## Guardrails

- Do not modify the reviewed change during an independent review unless the user explicitly asks for fixes.
- Do not report style-only noise already enforced by tools.
- Do not approve when evidence was produced for a different Git revision.
- Do not approve UI code merely because it resembles a confirmed HTML exploration; require framework-appropriate accessibility, responsiveness, interaction, and visual evidence.

Read [references/review-contract.md](references/review-contract.md) before finalizing the review.
