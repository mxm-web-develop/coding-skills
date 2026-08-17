# Stack detection

Use repository evidence in this order:

1. accepted architecture decisions and explicit project configuration;
2. manifests and lockfiles, including workspace declarations;
3. framework, compiler, build, test, lint, format, and package-manager configuration;
4. CI workflows and package scripts that are actually executed;
5. imports, source layout, generated-code markers, and existing tests;
6. user confirmation for unresolved or conflicting evidence.

Record every detected language and framework with at least one repository-relative evidence path. When a monorepo contains different stacks, profile each workspace and preserve its command working directory.

## Staleness

Treat the profile as stale when any of these changed after `detected_at`:

- manifests or lockfiles;
- compiler, framework, build, CI, lint, format, or test configuration;
- workspace/module roots;
- accepted architecture decisions;
- available Skill inventory when a task depends on it.

Do not re-profile for an unrelated prose or source-only change when the recorded evidence remains current.

## Command selection

Prefer commands already used by CI, then documented project commands, then package scripts or native tool defaults. Preserve working directory, environment prerequisites, and targeted variants. Never invent a passing command by omitting the failing package or suite.
