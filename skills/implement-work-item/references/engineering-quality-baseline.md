# Engineering quality baseline

## Structure

- One module should have one cohesive reason to change.
- Separate domain decisions and transformations from adapters, framework glue, persistence, transport, and rendering.
- Split large code into a named folder and small files by responsibility: model/types, pure logic, ports/interfaces, adapters, orchestration, and tests as applicable.
- Avoid catch-all `utils`, `helpers`, or `manager` modules when a domain name is available.
- Keep public APIs narrow. Hide implementation details and preserve existing compatibility unless a decision authorizes change.

## Functions and effects

- Prefer deterministic input-to-output functions for core rules.
- Pass dependencies and volatile values explicitly where the language/framework makes this practical.
- Keep mutation local and make state transitions visible.
- Handle errors at the layer that can add context or recover; do not silently swallow them.

## Readability

- Use domain terms consistently with project documents and nearby code.
- Prefer simple control flow, early validation, and explicit types/contracts.
- Comments explain why, invariants, protocol constraints, or non-obvious risk. Remove stale comments with the code they describe.
- Generated code stays in recorded generated roots and is not hand-edited.

## Change discipline

Match existing format, lint, type, build, and test commands. A new abstraction must simplify current responsibilities, not merely anticipate hypothetical reuse.
