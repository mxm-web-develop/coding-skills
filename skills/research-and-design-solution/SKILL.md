---
name: research-and-design-solution
description: Research, compare, and confirm a technical or UX/UI solution for an accepted requirement or work item. Use when implementation choices are uncertain, backend technologies or architecture need comparison, external APIs or current technical facts must be verified, frontend style or interaction direction should be explored with HTML prototypes, or a decision and rollback strategy are required before coding.
---

# Research and Design Solution

Ground design decisions in current project constraints and reproducible evidence.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). Compare options by user-relevant outcome, cost, risk, compatibility, and rollback; keep decision-object IDs, research routing, and internal record names out of the main explanation. Read [references/interactive-exploration.md](references/interactive-exploration.md) whenever the choice changes architecture, technology, dependencies, data/API shape, visual direction, motion, or user interaction.

## Procedure

1. Read the Requirement, Work Item, baseline, `.ai-flow/baseline/engineering-profile.json`, existing architecture, and relevant accepted decisions. Invoke `profile-project-engineering` when the profile is absent or stale.
2. State the decision to make, constraints, and evaluation criteria in product language. Ask about any missing priority that could change the recommendation.
3. Inspect the local code and recorded stack constraints before external research. Apply only relevant installed community Skills selected in the engineering profile.
4. Use primary sources for unstable technical facts and record source URLs and access dates.
5. Compare at least two genuinely viable options, including keeping the current design when applicable. Do not add filler alternatives merely to reach a count.
6. For backend, infrastructure, data, or framework choices, show each option's strengths, weaknesses, project fit, migration/operating cost, testing impact, risks, and rollback path in a concise table.
7. For visible frontend UX/UI choices with unresolved direction, create two or three meaningfully different, self-contained HTML prototypes. Cover representative content, responsive states, important interactions, and motion where relevant; keep them outside production source.
8. Recommend one direction. Explain why it best matches the user's priorities, what is sacrificed, and which uncertainty remains.
9. Present one clear choice: accept the recommendation, choose another option, combine named parts, or request another exploration. Do not begin production implementation while a material choice is awaiting confirmation.
10. Persist the comparison, recommendation rationale, prototype paths when present, and the user's confirmation or feedback. Hand off to `specify-tests` only after the required choice is confirmed.

## Guardrails

- Do not cite search summaries as authoritative sources.
- Do not add a dependency without comparing maintenance and supply-chain cost.
- Do not treat an attractive prototype as production-ready code or as proof of accessibility, responsiveness, performance, or framework compatibility.
- Do not use real customer data, credentials, production endpoints, remote scripts, or analytics in an HTML exploration.
- Do not silently choose a technology or visual direction when the alternatives materially change cost, architecture, migration, or user experience.
- Do not overwrite an accepted decision; create a new decision that supersedes it.

Read [references/decision-contract.md](references/decision-contract.md) before writing a decision.
