---
name: plan-product-delivery
description: Turn a confirmed product goal into understandable development stages and independently checkable tasks. Use after goal alignment, when arranging delivery, showing a roadmap, breaking down work that is too large for one change, or establishing dependencies and safe parallel boundaries.
---

# Plan Product Delivery

Create independently checkable pieces of product value.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). Keep milestone, vertical-slice, Work Item, plan-object, ID, gate, and abbreviation vocabulary inside machine records.

## Procedure

1. Read the accepted Goal, Requirements, baseline, and active decisions.
2. Map requirement dependencies, risk concentrations, external prerequisites, and unknowns.
3. Define milestone outcomes rather than activity lists.
4. Split milestones into small vertical Work Items containing behavior, tests, implementation, and documentation where applicable.
5. Assign each Work Item a scope, inputs, outputs, acceptance criteria, dependencies, risk, and likely change areas.
6. Identify items safe to run in parallel and declare non-overlapping file or component boundaries.
7. Put research spikes before decisions they unblock; do not disguise uncertain research as implementation.
8. Validate that every Requirement maps to at least one Work Item and planned test.
9. Present the arrangement as numbered stages and plain-language development tasks. Explain what the user will be able to see or verify after each stage, why the order matters, and which work can happen together. Ask whether to proceed with that arrangement, then persist the internal plan and task objects.

## Guardrails

- Avoid horizontal slices such as “build all models” with no user-observable result.
- Do not schedule implementation before required decisions or test design.
- Keep one Work Item small enough for a bounded Harness run and atomic review.
- Never say phrases such as “3 milestones / 7 work items,” “vertical slices,” “3MS+7WI,” “按 Skill 推荐,” or “确认 plan.” Say “三个阶段、七项开发任务；每项完成后都能单独检查效果.”

Read [references/plan-contract.md](references/plan-contract.md) before persisting the plan.
