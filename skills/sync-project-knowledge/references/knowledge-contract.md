# Knowledge synchronization contract

Machine truth:

- `.ai-flow/state/current.yaml` is the current snapshot.
- `.ai-flow/events/` is append-only history.
- `.ai-flow/evidence/` is immutable verification metadata.
- `.ai-flow/archive/` holds inactive historical objects.

Human board:

- `STATUS.md`: version, active goal, completed/in-progress/blocked, verification summary, next action.
- `ROADMAP.md`: active goals and near milestones only.
- `CURRENT_STATE.md`: current capabilities, boundaries, and accepted decisions.
- `RELEASES.md`: concise observed release history.

Reject active objects that depend on superseded objects as their current authority.
