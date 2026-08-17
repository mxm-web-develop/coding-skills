# Rust implementation

- Follow crate/workspace boundaries and the project's ownership, error, async, and feature-flag conventions.
- Model invalid states out where practical with enums and newtypes; keep transformations deterministic.
- Isolate I/O and runtime adapters from core domain modules.
- Avoid broad `unwrap`/`expect` in production paths unless an invariant is documented and enforced.
- Split modules by domain responsibility and keep public visibility minimal.
