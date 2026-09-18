---
name: upgrade-ai-project
description: Reconcile an installed AI Flow project after a tool upgrade while preserving its original goals, plan, tasks, progress and pending choices. Use after updating AI Flow, on a project/runtime version mismatch, or when the user asks to continue an existing project with the new version.
---

# Upgrade and Continue

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). Explain what remains valid, what needs reconciliation, and exactly where development resumes. A tool upgrade is not a new project, new plan, approval, release or completion.

## Recover the current project

1. Read `flowctl project upgrade --mode check`, Git status, the current task and latest checkpoint. If a migration exists, resume it instead of preparing another. Preserve uncommitted code and unresolved user choices.
2. If needed, run `flowctl project upgrade --mode prepare --root <root>`. This stores the original managed records and a path/hash inventory before replacing anything. Read `.ai-flow/state/upgrade.json` and its preserved originals; never infer user intent from filenames alone.
3. Run `flowctl project upgrade --mode apply --root <root>`. Conversion preserves identities and progress; unfamiliar fields and invalid records remain in the original archive. The engineering scan records observed manifests and commands, not passing tests. The project stays unavailable for new implementation until reconciliation finishes.
4. For each `review_required` source, read its preserved original. Extract still-current requirements, acceptance criteria, technical decisions, pending approvals and the next unfinished action into their corresponding current records. Link back to the original. Preserve the original task/goal IDs whenever recoverable; do not create a competing plan. Restore referenced prototype paths if a current decision still needs them, or update links to their archived locations.
5. Compare extracted facts with the repository and the existing plan. Mark uncertain or conflicting statements as questions, not accepted scope. Ask the user only when a genuine product decision cannot be recovered. Fill missing engineering and workspace facts using `profile-project-engineering`; preserve existing community skill choices and custom test commands when still valid. Archive old generated process material; do not archive a current requirement or decision merely because it predates the upgrade.
6. Run validation, then `flowctl project upgrade --mode finish --root <root> --summary '<retained plan, recovered facts, open questions and next action>' --resolved '<source path>' ...`. List only materials actually reconciled. Acknowledge all listed sources only after reading and extracting them. The command verifies links and regenerates current views before marking compatibility restored.
7. Show one concise continuation summary: unchanged goal, completed capabilities, current task, pending user choice/acceptance, next action. Resume the same run and ownership. If a checkpoint detects code drift, compare the diff and migration inventory before using an explicit drift override; never rewrite the old checkpoint to hide drift.

## Recovery and scope

- `--mode prepare` and `--mode apply` are retryable. Changed records stop application rather than overwrite concurrent work.
- `--mode restore` restores an unfinished migration's originals and preserves displaced reconciliation records. Reinstall the previous tool version before developing against restored records. A completed migration is not blindly rolled back after new development.
- History under `.ai-flow/archive/upgrades/` is evidence, not current instructions. Default development context comes from `flowctl context`, the current task and current requirements.
- Installation updates managed tool files; reconciliation updates managed project records. Record both versions; do not claim migration finished merely because installation succeeded.
- Unmanaged workspace documents require a read-only inventory and explicit source-to-target approval through the existing adoption workflow. Never interpret upgrade permission as permission to move arbitrary user documents or source code.
- Do not turn legacy tests or old completed tasks into newly verified tests or fabricated human acceptance. Retest active work against the current content before presenting acceptance.
