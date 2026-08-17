# Knowledge synchronization contract

Machine truth:

- `.ai-flow/state/current.yaml` is the current snapshot.
- `.ai-flow/events/` is append-only history.
- `.ai-flow/evidence/` is immutable verification metadata.
- `.ai-flow/archive/` holds inactive historical objects.
- `.ai-flow/baseline/workspace-document-inventory.json` is the approved source-to-archive map for legacy project documents.
- `.ai-flow/archive/legacy-documents/<version>/` preserves approved historical documents beneath their original relative paths.

Human board:

- `STATUS.md`: natural-language current-major summary plus child-version, Work Item, test, blocker, and next-action tables.
- `ROADMAP.md`: active goal and near milestones grouped by target version and exit gate.
- `CURRENT_STATE.md`: active requirements, current capabilities, boundaries, engineering profile, and accepted decisions.
- `RELEASES.md`: real Release objects with change, Evidence, known issue, migration, and rollback summaries.

Legacy history appears only as a short version summary, conflict/unknown count, and links to the archive inventory. It must not be treated as current authority.

Reject active objects that depend on superseded objects as their current authority.
