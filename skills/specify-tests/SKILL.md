---
name: specify-tests
description: Define executable, stack-aware acceptance and regression tests before implementation, including browser visual verification when UI changes. Use for new features, bug reproduction, behavior changes, risky refactors, or whenever a requirement lacks test coverage and traceability.
---

# Specify Tests

Translate accepted behavior into evidence that can fail before implementation and pass afterward.

## Procedure

1. Read the Requirement, Work Item, decision, implementation, existing test conventions, and `.ai-flow/baseline/engineering-profile.json`. Invoke `profile-project-engineering` when the profile is absent or stale.
2. Use [references/test-strategy-router.md](references/test-strategy-router.md) to read only the matching stack test playbooks and any installed community Skills selected by the profile.
3. List observable acceptance criteria, failure modes, changed boundaries, and regression risks.
4. Select the smallest effective mix of pure unit, component, integration, contract, end-to-end, migration, performance, security, accessibility, and visual tests. Reuse the project's existing runners.
5. For a bug, create or identify a reproduction that fails for the reported reason.
6. Define fixtures, deterministic environment, commands, expected results, and evidence. Map every test to Requirement and Work Item IDs.
7. Write tests before production behavior when safe and practical; otherwise document why and create an executable test specification.
8. Run pre-implementation tests and record actual results. A red test is valid only when its failure proves the missing behavior or defect.
9. Hand off to `implement-work-item` for features or `diagnose-and-verify` for defects.

## Guardrails

- Do not weaken assertions, delete regressions, or change visual baselines merely to pass.
- Do not add a second test framework when the current one can express the evidence.
- Keep unrelated flaky or environment failures separate and visible.
- Functional E2E, accessibility, screenshot regression, and visual design review are complementary evidence, not substitutes for one another.

Read [references/test-contract.md](references/test-contract.md) before finalizing coverage.
