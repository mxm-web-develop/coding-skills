---
name: initialize-ai-project
description: Initialize AI Flow in a blank or existing software project, optionally inventorying and consolidating scattered legacy documentation with explicit approval. Use when a user asks to set up, initialize, clean up, summarize, install into, or bring an existing repository under the AI development workflow, or when `.ai-flow/manifest.yaml` is missing.
---

# Initialize AI Project

Establish the project state without guessing its history or creating business code prematurely.

## Procedure

1. Locate the repository root and inspect the working tree without modifying it.
2. Run `.ai-flow/bin/flowctl doctor --root <root>` when the binary exists.
3. Classify the workspace:
   - Treat it as `greenfield` only when no meaningful code, build metadata, or Git history exists.
   - Treat it as `existing` when code, dependencies, releases, tests, substantial documentation, or project history exists.
   - Ask one concise question when the evidence is ambiguous.
4. When an existing or documentation-heavy workspace contains scattered project documents, ask whether the user wants `keep`, `audit-only`, or `summarize-and-archive`. Explain that any archive action requires a second approval of exact paths.
5. After the user confirms the mode and project name, run:

   ```sh
   .ai-flow/bin/flowctl project init --root <root> --mode <greenfield|existing> --name <name>
   ```

6. For `existing`, invoke `adopt-existing-project` with the selected documentation mode. Finish baseline confirmation and any separately approved cleanup before discussing new requirements.
7. For `greenfield`, align the first product goal with `discover-product-goal` before choosing a stack or scaffolding code.
8. After an existing baseline is accepted, invoke `profile-project-engineering`, then `discover-product-goal` and `plan-product-delivery` using the confirmed current version and capabilities.
9. Re-run `flowctl doctor` and report the generated machine state, archive, and human board paths.

## Guardrails

- Preserve pre-existing `AGENTS.md`, `CLAUDE.md`, `.cursor/`, and project documentation.
- Never interpret the initial cleanup preference as approval to move files. Require an inventory and explicit approval of source-to-target mappings first.
- Do not infer a current version for an existing project without evidence or user confirmation.
- Do not create workflow documents outside `.ai-flow/` and `docs/board/`.
- Stop before project initialization if the working tree contains overlapping user changes that the operation would overwrite.

Read [references/output-contract.md](references/output-contract.md) before writing or validating initialization output.
