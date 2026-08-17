---
name: specify-tests
description: Define executable acceptance and regression tests before implementation. Use for new features, bug reproduction, behavior changes, refactors with risk, or whenever a requirement lacks test coverage and traceability.
---

# Specify Tests

Translate accepted behavior into evidence that can fail before implementation and pass afterward.

## Procedure

1. Read the Requirement, Work Item, decision, current implementation, and existing test conventions.
2. List observable acceptance criteria and failure modes.
3. Select the smallest effective mix of unit, integration, contract, end-to-end, migration, performance, and security tests.
4. For a bug, create or identify a reproduction that fails for the reported reason.
5. Define fixtures, environment, commands, expected results, and evidence to capture.
6. Map every test to Requirement and Work Item IDs.
7. Write tests before production behavior when safe and practical; otherwise document the reason and create an executable test specification.
8. Run the relevant pre-implementation tests and record their actual result.
9. Hand off to `implement-work-item` for features or `diagnose-and-verify` for defects.

## Guardrails

- Do not weaken existing assertions to make implementation easier.
- Do not claim a red test is valid until its failure matches the intended missing behavior or bug.
- Keep unrelated flaky failures separate and visible.

Read [references/test-contract.md](references/test-contract.md) before finalizing test coverage.
