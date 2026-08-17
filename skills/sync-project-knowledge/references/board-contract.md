# Human board contract

The human board is a generated management view, not a second fact source. Follow the [user communication contract](../../orchestrate-ai-delivery/references/user-communication-contract.md). Keep exactly four files and lead each with a short natural-language explanation before tables.

## STATUS.md

Show:

- one paragraph explaining the current major version, active development version, product goal, completed/active/blocked task counts, verification result, and next action;
- a compact overview table using ordinary project language;
- child-version progress under the current major version;
- current-major or still-active development tasks with owner and test result;
- test checks with type, purpose, linked task, preparation status, and execution result;
- blockers and the next action.

Do not flood this page with completed tasks from older major versions; those belong in release history.

## ROADMAP.md

Explain the active goal in product language. Show target versions, development stages, expected outcomes, task completion, state, and completion conditions. Avoid implementation detail that belongs in internal task records or decisions.

## CURRENT_STATE.md

Show the current version and phase, confirmed needs for the active goal, current architecture/technical decisions, detected languages/frameworks/code organization, applicable development/testing approach, visual-check expectations, boundaries, risks, and unresolved technical facts.

## RELEASES.md

Render actual releases in descending semantic-version order. For each release show status, change summary, linked task count, test result, known issues, migration, and rollback. Never print “no releases” when a release record exists.

## Language and accuracy

- Use short natural-language labels for machine states and phases.
- Do not display internal object names, raw IDs, state values, Skill names, Playbook names, abbreviations, hashes, machine directories, or storage implementation notes in the main board.
- Prefer “开发阶段”“开发任务”“测试结果”“完成条件”“技术环境”“开发与测试规范” over Milestone, Work Item, Evidence, Gate, engineering profile, or Playbook.
- Preserve exact machine-object traceability in non-rendered HTML comments generated with the board. Do not show raw identifiers or machine paths in the rendered page, and never require them to understand status.
- Use “未记录”“待确认” or “尚无证据” when facts are absent; never infer a positive result.
- Escape table delimiters and collapse multiline text inside cells.
- Keep archived/superseded objects out of current tables.
- Write all four files atomically from validated `.ai-flow/` objects.
