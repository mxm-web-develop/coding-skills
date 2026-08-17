# Plan contract

For each milestone record outcome, included Requirements, exit gates, and target release.

For each Work Item record:

- Work Item ID, title, type, priority, and status.
- Parent Goal and Requirement IDs.
- User-observable acceptance criteria.
- Dependencies and blockers.
- Allowed change scope and protected paths.
- Required research, decisions, and tests.
- Risk and approval requirements.
- Parallelization group or serialization reason.

Reject a plan with orphan requirements or work items that cannot be independently verified.

## User presentation

The machine record keeps milestone, Work Item, Requirement, gate, ID and dependency fields. The user-facing proposal must translate them into:

- numbered development stages;
- plain-language tasks;
- what the user can see or check after each stage;
- why the order matters;
- work that can safely happen together;
- risks or decisions that still need attention.

Do not show object counts with internal nouns, abbreviations, IDs, schema fields or plan mechanics. Ask “是否按这个安排继续？” rather than asking the user to approve a plan object.
