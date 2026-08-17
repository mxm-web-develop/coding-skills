# Knowledge synchronization contract

Machine truth:

- `.ai-flow/state/current.yaml` is the current snapshot.
- `.ai-flow/events/` is append-only history.
- `.ai-flow/evidence/` is immutable verification metadata.
- `.ai-flow/archive/` holds inactive historical objects.
- `.ai-flow/baseline/workspace-document-inventory.json` is the approved source-to-archive map for legacy project documents.
- `.ai-flow/archive/legacy-documents/<version>/` preserves approved historical documents beneath their original relative paths.

Human board:

- `STATUS.md`: version, active goal, completed/in-progress/blocked, verification summary, next action.
- `ROADMAP.md`: active goals and near milestones only.
- `CURRENT_STATE.md`: current capabilities, boundaries, and accepted decisions.
- `RELEASES.md`: concise observed release history.

Legacy history appears only as a short version summary, conflict/unknown count, and links to the archive inventory. It must not be treated as current authority.

Reject active objects that depend on superseded objects as their current authority.
