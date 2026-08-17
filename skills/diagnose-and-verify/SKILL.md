---
name: diagnose-and-verify
description: Reproduce failures, identify root causes, fix defects, and collect trustworthy verification evidence. Use for reported bugs, failing tests, regressions, broken builds, implementation validation, or when a work item must prove that its acceptance criteria pass.
---

# Diagnose and Verify

Separate symptoms, design causes, implementation causes, and verified outcomes.

## Procedure

1. Read the Requirement, Work Item, test specification, implementation handoff, and existing evidence.
2. Reproduce the failure or run the declared verification command on the exact target revision.
3. Minimize the reproduction and form explicit hypotheses.
4. Inspect and instrument before changing code when the cause is uncertain.
5. Identify whether the root cause is in requirements, design, implementation, environment, data, or the test itself.
6. Return to `research-and-design-solution` when the accepted design is invalid; otherwise implement the smallest root-cause fix.
7. Run the focused reproduction, affected suite, and required regressions.
8. Execute required checks through `flowctl evidence run --work <id> --run <id> --test <id> -- <command>` so command, exit code, timestamp, Git SHA, environment, log, and SHA-256 are captured.
9. Write the diagnosis and evidence, then hand off to `review-change`.

## Guardrails

- Do not patch a symptom while leaving a known root cause unexplained.
- Do not label an unexecuted check as passed.
- Do not delete or weaken tests merely because they fail.
- Distinguish unrelated pre-existing failures from regressions introduced by the work item.

Read [references/evidence-contract.md](references/evidence-contract.md) before marking verification complete.
