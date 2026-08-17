---
name: profile-project-engineering
description: Detect a repository's real languages, frameworks, architecture, commands, test systems, visual-test needs, and compatible installed community Skills. Use after initialization or adoption, before technical design, test specification, implementation, diagnosis, or review, and whenever manifests, lockfiles, framework configuration, CI, or test tooling changes.
---

# Profile Project Engineering

Create the evidence-backed engineering profile that downstream Skills must follow.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md). Tell the user which technologies and project commands were found, what development/testing approach fits them, and what remains uncertain; do not present profile fields, playbook names, or community Skill routing as user-facing concepts.

## Procedure

1. Locate the project root and read `.ai-flow/manifest.yaml`, the adoption baseline, workspace structure inventory, accepted decisions, and current Work Item.
2. Inspect manifests, lockfiles, build and framework configuration, CI, scripts, source layout, generated-code markers, and existing tests. Do not infer a stack from file extensions alone.
3. Detect languages, runtimes, frameworks, package managers, build systems, module boundaries, public APIs, generated roots, and the project's actual quality commands using [references/stack-detection.md](references/stack-detection.md).
4. Determine whether the profile is new, unchanged, or stale relative to the current Git revision, workspace component graph, and configuration evidence. A completed cleanup that moves component roots, manifests, generated roots, or commands makes the profile stale.
5. Inventory only Skills already available to the active IDE/agent. Select relevant community Skills using [references/community-skill-routing.md](references/community-skill-routing.md) and the reviewed source catalog in [references/recommended-sources.md](references/recommended-sources.md); project rules and accepted decisions always take precedence.
6. Select the smallest matching development and test playbooks. For browser UI, mark visual verification required unless the change is provably non-visual.
7. Write or update `.ai-flow/baseline/engineering-profile.json` using [references/profile-contract.md](references/profile-contract.md), then run `flowctl validate`.
8. Save a Checkpoint that records profile changes, unknowns, selected playbooks, and community Skill provenance.

## Guardrails

- Do not silently install, update, or execute untrusted third-party Skills.
- Do not replace established project commands or frameworks merely because another tool is preferred.
- Record uncertain versions or architecture boundaries under `unknowns`; never manufacture certainty.
- Re-profile only evidence affected by a change, but keep the resulting document complete.
