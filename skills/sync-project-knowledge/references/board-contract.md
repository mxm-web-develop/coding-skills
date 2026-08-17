# Human board contract

The human board is a generated management view, not a second fact source. Keep exactly four files and lead each with a short natural-language explanation before tables.

## STATUS.md

Show:

- one paragraph explaining the current major version, active development version, product goal, completed/active/blocked task counts, verification result, and next action;
- a compact overview table;
- child-version progress under the current major version;
- current-major or still-active Work Items with owner and test result;
- test specifications with level, purpose, linked task, status, and Evidence summary;
- blockers and the next action.

Do not flood this page with completed tasks from older major versions; those belong in release history.

## ROADMAP.md

Explain the active goal in product language. Show target versions, milestones, expected outcomes, task completion, state, and exit gates. Avoid implementation detail that belongs in Work Items or decisions.

## CURRENT_STATE.md

Show the current version and phase, accepted requirements for the active goal, current architecture/technical decisions, detected language/framework/architecture Playbooks, visual-test expectations, boundaries, risks, and unresolved engineering facts.

## RELEASES.md

Render actual Release objects in descending semantic-version order. For each release show status, change summary, linked task count, Evidence result, known issues, migration, and rollback. Never print “no releases” when a Release object exists.

## Language and accuracy

- Use short natural-language labels for machine states and phases.
- Use “未记录”“待确认” or “尚无证据” when facts are absent; never infer a positive result.
- Escape table delimiters and collapse multiline text inside cells.
- Keep archived/superseded objects out of current tables.
- Write all four files atomically from validated `.ai-flow/` objects.
