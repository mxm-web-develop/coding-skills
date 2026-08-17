# Git integration contract

Branch patterns:

- `feat/WI-<id>-<slug>`
- `fix/WI-<id>-<slug>`
- `chore/WI-<id>-<slug>`

Commit example:

```text
feat(profile): add seller persona scoring

Work-Item: WI-2026-0012
Requirement: REQ-2026-0007
Goal: GOAL-2026-0001
Evidence: EV-2026-0042
```

Record base/head SHA, included paths, excluded user changes, local gates, remote checks, approvals, PR URL when present, and merge SHA when observed.
