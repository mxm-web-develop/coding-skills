# Interactive solution exploration

Use this contract to turn technical selection into a short, understandable conversation rather than a hidden agent decision.

## Choose the exploration shape

- Backend, infrastructure, data, API, framework, or dependency: compare viable technical options.
- Frontend UX/UI with uncertain visual or interaction direction: compare working HTML experiences.
- Mixed feature: confirm architectural constraints first, then explore only the visible choices those constraints leave open.
- Existing project with an accepted stack or design system: include “continue the current approach” and prefer compatible evolution unless evidence justifies migration.

## Backend and architecture comparison

Tailor four to seven criteria to the actual decision. Usually include project fit, delivery speed, maintainability, team familiarity, performance/scaling, security, operating cost, ecosystem maturity, migration effort, testing, and rollback.

Present a compact table with:

- option and plain-language summary;
- strengths;
- weaknesses and failure modes;
- fit with the current repository and team;
- adoption, migration, and operating impact;
- recommendation status.

Recommend exactly one direction when evidence supports it. Explain the recommendation from the user's stated priorities, not from popularity. State the strongest reason not to choose it and any fact that still needs a spike or benchmark.

When one missing constraint could reverse the recommendation, state a conditional recommendation and ask only for that constraint first. Do not ask the user to approve the final technology in the same turn. Update the comparison after the answer, then request the direction choice.

## Frontend UX/UI exploration

Create two or three meaningfully different directions, not cosmetic recolors of the same layout. Vary the information hierarchy, component composition, density, navigation, feedback, or interaction model according to the uncertainty being tested.

Store working artifacts under:

```text
.ai-flow/prototypes/<decision-or-topic>/<option>/index.html
```

Each option must:

- be self-contained HTML/CSS/JavaScript with no build step;
- use representative but synthetic content and include loading, empty, error, success, and disabled states when relevant;
- demonstrate desktop and mobile behavior;
- make primary interactions clickable;
- demonstrate meaningful motion when motion is part of the choice and honor `prefers-reduced-motion`;
- label itself as an exploration, not production UI;
- avoid remote scripts, trackers, production endpoints, secrets, and real user data.

Reuse the current product's known design tokens and constraints when they exist. A raw HTML exploration may simulate framework components, but it must not be copied directly into production without implementation and testing.

Give the user an easy preview path or start a local static preview when the active IDE supports it. Summarize each option in ordinary language: what it feels like, what it optimizes, and what tradeoff the user will notice. Screenshots can supplement a live prototype but do not replace clickable interaction when interaction is under evaluation.

If the user combines parts of multiple directions, create one refined confirmation prototype before production implementation. After confirmation, archive rejected explorations under `.ai-flow/archive/design-explorations/<version>/` and keep links in the decision record valid.

Treat information hierarchy, visual density, mobile behavior, motion, and interaction as one coherent experience direction when they depend on each other. Do not make the user approve each attribute separately. If the primary user outcome is still unclear, ask for that outcome first; then let the working prototypes carry the related choices together.

## Confirmation

Ask one decision at a time. Offer these natural actions without exposing internal status names:

1. Use the recommended direction.
2. Choose another named direction.
3. Combine specifically named parts.
4. Explore again because a stated concern is unresolved.

Record the selected option, feedback, confirmation time, and whether another pass is required. A material choice is ready for implementation only when the user has confirmed it. A prior project decision may count as confirmation when it clearly covers the same scope and remains current.
