---
name: implement-work-item
description: Implement one approved and test-specified AI Flow work item using the repository's detected language and framework playbooks. Use when requirements, design decisions, scope, tests, and an engineering profile are ready and the user wants production code changed, a feature built, or a bounded refactor completed.
---

# Implement Work Item

Deliver the smallest correct, maintainable change within the approved scope.

## Procedure

1. Read the Work Item, Requirements, accepted decisions, test specification, current Checkpoint, and `.ai-flow/baseline/engineering-profile.json`. Invoke `profile-project-engineering` when the profile is absent or stale.
2. Read [references/engineering-quality-baseline.md](references/engineering-quality-baseline.md), then select only the matching implementation playbook through [references/stack-router.md](references/stack-router.md).
3. Read any installed community Skills selected in the engineering profile. Project conventions and accepted decisions take precedence.
4. Confirm the working tree and target branch/worktree, then run `flowctl work start` to acquire the allowed-path writing lease. Resume the latest Checkpoint when a Run already exists.
5. Run relevant baseline checks and write a short module/file plan before editing. Name responsibilities, pure core logic, side-effect boundaries, public interfaces, and tests.
6. Implement one coherent increment that satisfies the next failing test or acceptance criterion. Split code by responsibility as it grows; do not build a large multipurpose file.
7. Run focused tests after each meaningful increment. Refactor only while behavior remains green.
8. Record files, design deviations, selected playbooks/Skills, commands, and remaining work with `flowctl checkpoint save`.
9. Hand off to `diagnose-and-verify`; implementation alone never proves completion.

## Guardrails

- Do not expand scope, overwrite another owner's changes, or rewrite unrelated code.
- Prefer pure functions for domain transformations and decisions; isolate I/O, clocks, randomness, network, storage, UI, and framework lifecycle at explicit boundaries.
- Comment invariants, intent, tradeoffs, and surprising constraints. Do not narrate obvious syntax or compensate for unclear structure with comments.
- Follow existing project thresholds and tooling. When none exist, split proactively by responsibility rather than imposing an arbitrary universal line limit.
- Do not add dependencies, change public APIs, migrate data, commit, push, merge, tag, or publish without the required decision and routing.

Read [references/implementation-contract.md](references/implementation-contract.md) before handing off.
