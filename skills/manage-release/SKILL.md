---
name: manage-release
description: Plan and record an AI Flow project version after approved changes pass their gates. Use when selecting the next version, preparing release notes, tagging or publishing a release, recording a feature or fix version, or answering what is included in a release.
---

# Manage Release

Derive the version from policy and verified change impact.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). Explain the proposed version, included changes, test result, known issues, migration, and rollback in release language; do not expose internal object names or gate states.

## Procedure

1. Read the current version, versioning policy, merged Work Items, evidence, review state, and unreleased changes.
2. Confirm every included change is integrated and traceable.
3. Classify impact as breaking, compatible feature, fix, or non-release change.
4. Propose the next version; use SemVer unless the project explicitly defines another parseable policy.
5. Generate a concise human summary and a machine release record.
6. Include migrations, compatibility notes, known issues, verification, and rollback information.
7. Request approval before creating a tag, GitHub release, package publication, or deployment.
8. After an observed release action, record its immutable identifiers and invoke `sync-project-knowledge`.

## Guardrails

- Do not include unmerged or unverified work as released.
- Do not infer that a tag or deployment succeeded from a requested command.
- Do not silently change versioning policy.
- Keep release state distinct from deployment state.

Read [references/release-contract.md](references/release-contract.md) before finalizing a version.
