---
name: initialize-ai-project
description: Initialize AI Flow in a blank or existing software project, optionally inventorying and consolidating scattered legacy documentation with explicit approval while only marking non-document cleanup candidates. Use when a user asks to set up, initialize, install into, summarize, or bring a repository under the AI development workflow, or when `.ai-flow/manifest.yaml` is missing. Post-initialization code and directory cleanup belongs to `clean-project-workspace`.
---

# Initialize AI Project

Establish the project state without guessing its history or creating business code prematurely.

Follow the [user communication contract](../orchestrate-ai-delivery/references/user-communication-contract.md) for every question and report. Keep initialization modes, state names, file formats, and Skill routing out of the default conversation.

## Procedure

1. Locate the repository root and inspect the working tree without modifying it.
2. Run `.ai-flow/bin/flowctl doctor --root <root>` when the binary exists.
3. Classify the workspace:
   - Treat it as `greenfield` only when no meaningful code, build metadata, or Git history exists.
   - Treat it as `existing` when code, dependencies, releases, tests, substantial documentation, or project history exists.
   - Ask one concise question when the evidence is ambiguous.
4. When an existing or documentation-heavy workspace contains scattered project documents, ask in natural language whether the user wants to keep everything in place, receive a read-only organization list, or review and archive old documents. Map the answer internally to `keep`, `audit-only`, or `summarize-and-archive`; never require the user to understand those values. Explain that no document moves until the user sees and approves the exact list.
5. After the user confirms the mode and project name, run:

   ```sh
   .ai-flow/bin/flowctl project init --root <root> --mode <greenfield|existing> --name <name>
   ```

6. For `existing`, invoke `adopt-existing-project` with the selected documentation mode. Map languages, subprojects, dependency boundaries, generated roots, and suspected non-document leftovers into a read-only workspace structure inventory. Finish baseline confirmation and any separately approved document archive before discussing new requirements.
7. For `greenfield`, align the first product goal with `discover-product-goal` before choosing a stack or scaffolding code.
8. After an existing baseline is accepted, invoke `profile-project-engineering`, then `discover-product-goal` and `plan-product-delivery` using the confirmed current version and capabilities.
9. Re-run `flowctl doctor` and report whether setup succeeded, what was learned about the project, whether any approved documents moved, where the readable project overview is available, and what happens next. Mention machine paths only when the user needs them.

## Guardrails

- Preserve pre-existing `AGENTS.md`, `CLAUDE.md`, `.cursor/`, and project documentation.
- Never interpret the initial cleanup preference as approval to move files. Require an inventory and explicit approval of source-to-target mappings first.
- During initialization, never move, delete, rename, merge, or rewrite code, directories, generated outputs, caches, binaries, configuration, infrastructure, migrations, or other non-document content. Mark candidates with `initialization_action: mark-only` for later revalidation.
- Do not invoke `clean-project-workspace` until initialization is complete and the user later makes an explicit workspace-cleanup request.
- Do not infer a current version for an existing project without evidence or user confirmation.
- Do not create workflow documents outside `.ai-flow/` and `docs/board/`.
- Stop before project initialization if the working tree contains overlapping user changes that the operation would overwrite.

Read [references/output-contract.md](references/output-contract.md) before writing or validating initialization output.
