# Python tests

- Use the configured runner, fixtures, async plugin, coverage, type, and lint commands.
- Prefer parameterized unit tests for pure rules and scoped fixtures for effects.
- Test package/public behavior rather than private implementation details.
- Cover filesystem, database, network, process, environment, and time boundaries with temporary or controlled resources.
- Preserve exception type, message/context, and cleanup behavior where they form part of the contract.
