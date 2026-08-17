# Workspace structure inventory contract

Write `.ai-flow/baseline/workspace-structure-inventory.json` for an existing project.

## Repository graph

Record the VCS root and classify topology as single-project, monorepo, multi-project, nested-repositories, or unknown. Detect workspace manifests, nested repositories and shared roots before enumerating components.

For every component record its root, role, languages, manifests, build systems, internal component dependencies, entry points, tests, deployment evidence, status and evidence. Use manifests, build graphs, CI, deployment and runtime configuration; file extensions alone are insufficient.

Support mixed ecosystems in one repository. JavaScript/TypeScript workspaces, Python projects, Go workspaces, Cargo workspaces, Maven/Gradle modules, .NET solutions, mobile/desktop apps, native builds, infrastructure and unknown toolchains may coexist. Preserve uncertainty instead of forcing them into one root toolchain.

## Initialization candidates

Non-document content may be classified only as:

- `candidate-deprecated` or `candidate-duplicate` when evidence suggests later cleanup investigation;
- `generated`, `cache`, or `vendor` when its origin is known;
- `protected`, `shared`, or `unknown` when it must remain untouched.

Every candidate must use `initialization_action: mark-only`. Do not include an archive target, delete action, approval, or execution result. A name such as `old`, `legacy`, `backup`, `new`, a date, or a version is weak evidence and never confirms deprecation.

Exclude `.git`, `.ai-flow`, installed AI Flow Skill copies and dependency/vendor trees from deep semantic scanning. Record nested repositories but do not enter them without separate scope.

## Relationship to cleanup

The inventory is a dated baseline and candidate source only. A later explicit user request invokes `clean-project-workspace`, which must re-scan the current Git revision and create a separately approved cleanup plan. It may reject or reclassify any initialization candidate.

