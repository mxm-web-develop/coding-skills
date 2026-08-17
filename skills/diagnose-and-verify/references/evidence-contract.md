# Diagnosis and evidence contract

Diagnosis fields:

- Symptom and impact.
- Reproduction steps and minimal failing case.
- Investigated hypotheses and evidence.
- Root cause category and explanation.
- Design issue, implementation issue, or both.
- Fix and regression protection.

Verified evidence fields:

- Evidence ID, Work Item/Test IDs, command, exit code, start/end time, Git SHA, environment, report/log path, and SHA-256 hash.
- Result: `passed`, `failed`, `blocked`, or `unverified`.

Agent prose without executed evidence must remain `unverified`.
