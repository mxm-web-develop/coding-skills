---
name: research-and-design-solution
description: Research and design a technical solution for an accepted requirement or work item. Use when implementation choices are uncertain, external APIs or current technical facts must be verified, architecture changes are proposed, or a decision and rollback strategy are required before coding.
---

# Research and Design Solution

Ground design decisions in current project constraints and reproducible evidence.

## Procedure

1. Read the Requirement, Work Item, baseline, existing architecture, and relevant accepted decisions.
2. State the decision to make, constraints, and evaluation criteria.
3. Inspect the local code before external research.
4. Use primary sources for unstable technical facts and record source URLs and access dates.
5. Compare viable options, including keeping the current design when applicable.
6. Evaluate correctness, complexity, compatibility, security, performance, migration, operability, testing, and rollback.
7. Prototype only when it reduces a named uncertainty; keep prototype code separate from production changes.
8. Recommend one option with consequences and rejected alternatives.
9. Obtain approval for high-impact architecture, dependency, data, API, or migration decisions.
10. Persist the decision and hand off to `specify-tests`.

## Guardrails

- Do not cite search summaries as authoritative sources.
- Do not add a dependency without comparing maintenance and supply-chain cost.
- Do not overwrite an accepted decision; create a new decision that supersedes it.

Read [references/decision-contract.md](references/decision-contract.md) before writing a decision.
