# Multilingual repository discovery

Build a component graph before classifying cleanup candidates.

## Boundaries

Detect and record:

- VCS root, submodules, nested repositories, worktrees, symlinks, and paths outside the root.
- Workspace definitions such as npm/yarn/pnpm workspaces, Nx, Turborepo, Lerna, Rush, `go.work`, Cargo workspaces, Maven modules, Gradle settings, .NET solutions, Bazel, Pants, Buck, CMake, and language-specific project files.
- Deployable applications, services, workers, CLIs, browser extensions, desktop/mobile apps, libraries, infrastructure, migrations, schemas, documentation sites, examples, fixtures, and code generators.
- Shared roots consumed by more than one component and root-level scripts/configuration that coordinate components.

Do not require one manifest per component. Infer a boundary only from multiple signals and record the evidence.

## Language and toolchain evidence

Inspect relevant manifests and locks rather than relying on extensions alone, including:

- JavaScript/TypeScript: `package.json`, workspace config, lockfiles, tsconfig references, framework/build/test config.
- Python: `pyproject.toml`, lockfiles, `requirements*.txt`, `setup.*`, Pixi/Conda config, import roots and test config.
- Go: `go.mod`, `go.work`, packages, generated markers and build tags.
- Rust: `Cargo.toml`, workspace membership, features and build scripts.
- JVM: Maven/Gradle module graphs, source sets, generated sources and packaging tasks.
- .NET: solution/project references, target frameworks, generated `bin/obj`, packaging and publish profiles.
- Ruby, PHP, Elixir, Swift, Kotlin, Dart/Flutter, C/C++, infrastructure and other ecosystems: their manifests, locks, build graphs, generated roots, test commands and deployment entry points.

Unknown ecosystems remain `unknown`; do not improvise deletion rules.

## Liveness evidence

For each candidate, inspect all applicable evidence:

1. Direct imports, project/package references and dynamic loading or plugin registration.
2. Root and component build, test, lint, type-check, packaging and code-generation tasks.
3. CI matrices, release workflows, deployment manifests, containers, infrastructure and operational scripts.
4. Runtime configuration, feature flags, environment-variable wiring, routes, entry points and data migrations.
5. Tests, fixtures, snapshots, examples and developer tooling.
6. Accepted decisions, current boards, ownership rules and user confirmation.
7. Git history as supporting context only. Age or low activity is not proof of deprecation.

Classify shared code conservatively. If any consumer remains active or unknown, keep the shared path.

## Generated and external content

Separate:

- reproducible outputs and caches;
- vendored dependencies or externally synchronized sources;
- checked-in generated sources required by builds or releases;
- large binaries and release artifacts;
- local secrets and environment files.

Never archive vendor/cache trees as historical source. Never touch secrets. Checked-in generated sources require their real build/release contract before any action.

