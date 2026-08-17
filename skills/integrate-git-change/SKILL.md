---
name: integrate-git-change
description: Prepare and validate Git branches, atomic commits, commit messages, pull requests, and merge gates for an approved AI Flow change. Use when the user asks to commit, prepare a PR, integrate, merge, or inspect whether a change is ready for source control.
---

# Integrate Git Change

Make every integrated change traceable, reviewable, and reversible.

## Procedure

1. Read the Work Item, verification evidence, review disposition, Git policy, and current working tree.
2. Stop if required tests or blocking review findings are unresolved.
3. Separate unrelated user changes and generated noise from the intended change.
4. Group the diff into atomic, independently reversible commits.
5. Use Conventional Commits and include traceability trailers from [references/git-contract.md](references/git-contract.md).
6. Run the configured pre-commit gate on the exact staged content.
7. Show the proposed commit set and request approval when project policy requires it.
8. Create commits only when authorized. Push, PR creation, merge, and branch deletion each require their own policy permission.
9. Record commit SHAs and integration state, then hand off to `manage-release` when the work item changes a releasable version.

## Guardrails

- Never use destructive Git cleanup to make the tree appear clean.
- Never stage secrets, unrelated files, or another Work Item's changes.
- Never rewrite shared history without explicit authorization and policy support.
- Do not claim CI passed until the remote check result is observed.
