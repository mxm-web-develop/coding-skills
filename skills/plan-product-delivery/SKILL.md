---
name: plan-product-delivery
description: Decompose an accepted product goal into milestones and independently verifiable work items. Use after goal alignment, when creating a delivery plan or roadmap, when work is too large for one change, or when dependencies and safe parallel boundaries must be established.
---

# Plan Product Delivery

Create vertical slices that deliver observable value and can pass gates independently.

## Procedure

1. Read the accepted Goal, Requirements, baseline, and active decisions.
2. Map requirement dependencies, risk concentrations, external prerequisites, and unknowns.
3. Define milestone outcomes rather than activity lists.
4. Split milestones into small vertical Work Items containing behavior, tests, implementation, and documentation where applicable.
5. Assign each Work Item a scope, inputs, outputs, acceptance criteria, dependencies, risk, and likely change areas.
6. Identify items safe to run in parallel and declare non-overlapping file or component boundaries.
7. Put research spikes before decisions they unblock; do not disguise uncertain research as implementation.
8. Validate that every Requirement maps to at least one Work Item and planned test.
9. Present sequencing and tradeoffs for confirmation, then persist the plan and create accepted slices with `flowctl work create`.

## Guardrails

- Avoid horizontal slices such as “build all models” with no user-observable result.
- Do not schedule implementation before required decisions or test design.
- Keep one Work Item small enough for a bounded Harness run and atomic review.

Read [references/plan-contract.md](references/plan-contract.md) before persisting the plan.
