# TypeScript and Node.js tests

- Use the detected runner and its existing configuration, module system, environment, and coverage rules.
- Unit-test pure transformations and domain decisions with table-driven cases where useful.
- Test adapters at their real boundary or with faithful fakes; avoid mocking every internal call.
- For HTTP, queues, databases, files, or external APIs, cover serialization, validation, errors, timeouts, and retries at integration/contract level.
- Keep fixtures typed, minimal, and local to the behavior they explain.
