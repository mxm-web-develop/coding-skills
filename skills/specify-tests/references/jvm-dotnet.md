# JVM and .NET tests

- Use the configured JUnit/TestNG or xUnit/NUnit/MSTest-style runner and framework test facilities already present.
- Keep domain tests framework-light; use container/application context tests only for boundaries that need them.
- Cover dependency injection wiring, serialization, persistence transactions, configuration, async/cancellation, and compatibility at the appropriate integration level.
- Avoid broad context-heavy tests for logic that can be proven with a small deterministic unit.
