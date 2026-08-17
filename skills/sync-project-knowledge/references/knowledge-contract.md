# Knowledge synchronization contract

Machine truth:

- `.ai-flow/state/current.yaml` is the current snapshot.
- `.ai-flow/events/` is append-only history.
- `.ai-flow/evidence/` is immutable verification metadata.
- `.ai-flow/archive/` holds inactive historical objects.
- `.ai-flow/baseline/workspace-document-inventory.json` is the approved source-to-archive map for legacy project documents.
- `.ai-flow/baseline/workspace-structure-inventory.json` records the current multilingual component graph and mark-only cleanup candidates.
- `.ai-flow/workspace-cleanup/PLAN-*.json` records explicit post-initialization cleanup approvals, batches, results, verification and recovery.
- `.ai-flow/archive/legacy-documents/<version>/` preserves approved historical documents beneath their original relative paths.
- `.ai-flow/archive/legacy-code/<version>/` and `.ai-flow/archive/legacy-files/<version>/` preserve only content approved by a cleanup plan.

Human board:

- `STATUS.md`: natural-language current-major summary plus child-version, Work Item, test, blocker, and next-action tables.
- `ROADMAP.md`: active goal and near milestones grouped by target version and exit gate.
- `CURRENT_STATE.md`: active requirements, current capabilities, boundaries, engineering profile, and accepted decisions.
- `RELEASES.md`: real Release objects with change, Evidence, known issue, migration, and rollback summaries.

Legacy history appears only as a short version summary, conflict/unknown count, cleanup/recovery summary, and links to the relevant inventory or plan. It must not be treated as current authority.

Reject active objects that depend on superseded objects as their current authority.
