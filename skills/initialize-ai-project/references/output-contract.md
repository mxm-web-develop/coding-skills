# Initialization output contract

Produce or confirm:

- `.ai-flow/manifest.yaml` with schema and installed pack versions.
- `.ai-flow/project.yaml` with project name, mode, profile, version policy, and commands when known.
- `.ai-flow/state/current.yaml` with `revision: 1` and `phase` set to `goal_alignment` or `baselining`.
- `.ai-flow/capabilities.yaml` from the installer/doctor result.
- `.ai-flow/baseline/workspace-document-inventory.json` when documentation audit or cleanup was selected.
- `docs/board/STATUS.md`, `ROADMAP.md`, `CURRENT_STATE.md`, and `RELEASES.md` generated from machine state.

Report separately:

- Observed facts.
- Inferences requiring confirmation.
- Blocking questions.
- Documentation mode (`keep`, `audit-only`, or `summarize-and-archive`), approval state, and archived/restored path mapping when applicable.
- The next Skill to invoke.
