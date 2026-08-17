# Test strategy router

Select tests from risk and detected stack, not a fixed pyramid.

| Detected work | Playbook |
|---|---|
| Browser-rendered UI or web journey | [web-and-visual.md](web-and-visual.md) |
| TypeScript/JavaScript library or Node service | [typescript-node.md](typescript-node.md) |
| Python | [python.md](python.md) |
| Go or Rust | [go-rust.md](go-rust.md) |
| Java/Kotlin or .NET | [jvm-dotnet.md](jvm-dotnet.md) |
| Native/cross-platform mobile | [mobile.md](mobile.md) |

Start with fast deterministic tests around pure core rules, then test integration boundaries, and reserve E2E for critical user/system journeys. Add contract, migration, concurrency, performance, or security coverage when the Work Item changes that risk.

Use test commands and installed community Skills recorded in the engineering profile. If a selected Skill conflicts with current project configuration, the project wins and the conflict is recorded.
