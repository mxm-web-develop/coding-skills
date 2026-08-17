# Engineering profile contract

Write `.ai-flow/baseline/engineering-profile.json` and validate it against `engineering-profile.schema.json`.

The profile must contain:

- the Git revision and detection time;
- evidence-backed languages and frameworks;
- package managers and build systems;
- module, generated-code, and public-API roots;
- exact install, build, format, lint, type-check, unit, integration, E2E, and visual commands that exist;
- selected Core playbook names;
- selected installed community Skills with provenance;
- browser visual-test requirements, browsers, and viewports;
- unresolved facts under `unknowns`.

Use repository-relative POSIX paths. Keep commands as complete shell strings for human/agent execution, but never place credentials in the profile. Increment `revision` when semantic content changes.
