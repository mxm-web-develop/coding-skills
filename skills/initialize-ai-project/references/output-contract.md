# Initialization output contract

Produce or confirm:

- `.ai-flow/manifest.yaml` with schema and installed pack versions.
- `.ai-flow/project.yaml` with project name, mode, profile, version policy, and commands when known.
- `.ai-flow/state/current.yaml` with `revision: 1` and `phase` set to `goal_alignment` or `baselining`.
- `.ai-flow/capabilities.yaml` from the installer/doctor result.
- `.ai-flow/baseline/workspace-document-inventory.json` when documentation audit or cleanup was selected.
- `.ai-flow/baseline/workspace-structure-inventory.json` for existing projects, recording languages, subprojects, shared roots, nested repositories, generated roots, and non-document cleanup candidates as `mark-only`.
- `docs/board/STATUS.md`, `ROADMAP.md`, `CURRENT_STATE.md`, and `RELEASES.md` generated from machine state.

Report separately:

- What is already confirmed about the project.
- What still needs the user's confirmation.
- Questions that block the next useful step.
- The user's document-handling choice in natural language and, when applicable, the exact files moved from one readable location to another.
- Possible old code or other clutter grouped by part of the project, with a clear statement that initialization only marked it for later review and did not move or delete it.
- What will happen next, described as an activity rather than a Skill name.

Keep mode values, object status, revisions, hashes, schema names and machine paths in the internal record unless the user explicitly requests technical details.
