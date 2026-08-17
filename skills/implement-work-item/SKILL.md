---
name: implement-work-item
description: Implement one approved and test-specified AI Flow work item. Use when requirements, design decisions, scope, and tests are ready and the user wants production code changed, a feature built, or a bounded refactor completed.
---

# Implement Work Item

Deliver the smallest correct change within the approved scope.

## Procedure

1. Read the Work Item, linked Requirements, accepted decisions, test specification, and current checkpoint.
2. Confirm the working tree and target branch/worktree, then run `flowctl work start` to acquire the allowed-path writing lease. If a Run already exists, resume its latest Checkpoint instead.
3. Run the relevant baseline checks before editing.
4. Implement one coherent increment that satisfies the next failing test or acceptance criterion.
5. Follow existing project architecture and conventions unless an accepted decision changes them.
6. Run focused tests after each meaningful increment; keep failures visible.
7. Refactor only after behavior is correct and tests remain green.
8. Record changed files, design deviations, commands run, and remaining work with `flowctl checkpoint save`.
9. Hand off to `diagnose-and-verify`; do not declare completion from implementation alone.

## Guardrails

- Do not expand scope silently or opportunistically rewrite unrelated code.
- Do not overwrite user changes or another Work Item's leased scope.
- Do not add production dependencies, change public APIs, or migrate data without the required decision and approval.
- Do not commit, push, merge, tag, or publish unless explicitly routed through the Git/release Skills.

Read [references/implementation-contract.md](references/implementation-contract.md) before handing off.
