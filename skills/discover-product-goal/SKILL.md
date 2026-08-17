---
name: discover-product-goal
description: Turn an initial product idea or change request into an agreed, testable goal. Use when starting a project, defining a major version objective, clarifying a broad feature, resolving ambiguous requirements, or when success criteria and non-goals are missing.
---

# Discover Product Goal

Reach shared understanding before decomposing or implementing work.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). Use the user's product vocabulary and never ask them to approve a Goal/Requirement object or “contract.”

## Procedure

1. Read the current baseline, active goals, constraints, and unresolved decisions.
2. Restate the user's intended outcome in plain language and identify ambiguity.
3. Discuss one decision branch at a time: users, problem, desired behavior, value, boundaries, constraints, risks, data, and acceptance.
4. Separate the major goal from candidate features and implementation ideas.
5. Define measurable success and explicit non-goals.
6. Surface privacy, safety, compliance, migration, compatibility, and operational concerns when relevant.
7. Present a concise “我理解的版本目标” summary covering who it serves, the result to achieve, what is included, what is excluded, and how the user will judge success. Ask whether that understanding is correct.
8. Write the confirmed Goal and Requirements under `.ai-flow/`; leave unconfirmed items as questions, not facts.
9. Hand off to `plan-product-delivery`.

## Guardrails

- Do not choose architecture solely to make the requirement discussion easier.
- Do not turn every idea into committed scope.
- Do not invent personas, metrics, regulations, or business rules.
- Stop and request a decision when alternatives materially change scope or risk.

Read [references/goal-contract.md](references/goal-contract.md) before finalizing the goal.
