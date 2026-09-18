# Completion, acceptance and working memory

## Delivery boundary

Automated checks, review, user acceptance and publication are separate facts. A passing command does not mean the user has verified the result; a request to continue or commit is not acceptance. Local commits may be explicitly authorized while acceptance remains pending. Publication requires its own authorization and project gates.

1. Define observable acceptance criteria and register required tests on the task (`--required-test`) or in its linked test specifications. Keep command evidence for the current code including uncommitted changes. Re-run all required checks after a source change; imported claims and historical passes cannot certify new code.
2. Review the actual change and record `flowctl work review --id <id> --reviewer <identity> --decision approved|changes_required --summary <conclusion>`. Resolve blocking findings before approval. This is a review record, never human acceptance.
3. Present a short verification card: where/how to open the result, prerequisites, numbered user actions and expected visible outcomes, and known limitations. Persist the exact card using `flowctl work acceptance --id <id> --instructions <setup> --step <action> --expected <result>` (repeat pairs; `--limitation` optional). Tell the user that automated checks passed and their verification is pending. Save progress before waiting.
4. Record only actual user feedback using `flowctl work accept --id <id> --by <person> --feedback <actual feedback> --result passed|failed`. An ambiguous “OK” is not sufficient if the conversation also contains another pending choice; ask which result they verified. Failure retains the same task and returns to diagnosis. A code change after the card requires `work reopen --reason <change>`, fresh checks/review and a new card. Never invent acceptance on behalf of the user.
5. After a pass, complete the already-authorized closeout without asking again: update current capabilities, accepted decisions, requirement outcomes, plan progress and the next actionable step; then `flowctl work complete --id <id> --summary <delivered capability and remaining limits>`. This requires current tests, review and acceptance and archives the task, run, saved progress, tests, evidence/logs and task-scoped reports. Retry closeout after interruption; do not start a replacement task.
6. Regenerate and validate the boards. Tell the user what is now usable and what is next; do not claim publication unless it happened.

## Working memory

- Default reads: current snapshot, current goal/plan, relevant active requirements/accepted decisions, active task, latest checkpoint and pending approvals. Start with `flowctl context`; use `--work <id>` for the verification card. Historical detail is opt-in (`--history` or a specific archived ID), never bulk-loaded at every turn.
- Keep user status to one short screen: target/outcome, recently delivered capability, current work, blockers, next action. Link optional detail. Do not enumerate every past retry or completed task.
- Store new transient reports/prototypes beneath `.ai-flow/reports/<work-id>/` and `.ai-flow/prototypes/<work-id>/` so closeout can collect them. Before closeout promote lasting requirements, decisions and capabilities to their canonical objects. If a current decision still needs a prototype, update its link to the archive before final validation.
- `.ai-flow/archive/catalog.json` maps stable record paths to preserved history. IDs and historical evidence remain resolvable. Do not delete the archive to reduce context.
- Current plan selection excludes superseded/cancelled/archived plans; planned releases are not released versions. Generated obsolete plan pages are archived when rendering. Only generator-owned pages can be reconciled automatically; arbitrary user documents require a reviewed path map.
- Existing unscoped reports require inventory and reconciliation during upgrade. Do not silently treat them as current or move unrelated project documentation.

## Recovery and budgets

Save a checkpoint when a time/retry/file budget is reached. Diagnose or revise scope and use `flowctl work budget --id <id> --reason <specific justification>` with changed limits; do not repeatedly bypass the limit. Resume the same run after a pause or editor change. Serialization and dependencies apply to resumed tasks as well as new tasks.
