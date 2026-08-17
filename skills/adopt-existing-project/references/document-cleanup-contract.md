# Legacy document cleanup contract

## Modes and approval gates

- `keep`: record that cleanup was declined; do not create a move plan.
- `audit-only`: create the inventory, summary, conflicts, and recommendations; do not move files.
- `summarize-and-archive`: first create the same read-only proposal, then ask the user to approve exact source-to-target mappings. A general request to “clean the workspace” is not path-level approval.

Never combine discovery and mutation in one step. Re-read Git status immediately before applying an approved plan and stop if a source changed after hashing.

## Discovery scope

Inspect project-authored Markdown, MDX, text, reStructuredText, AsciiDoc, office/PDF planning artifacts, release notes, reports, ADRs, specs, and similarly named documentation. Exclude `.git`, `.ai-flow`, installed Skill/runtime directories, dependency/vendor folders, build output, caches, secrets, binaries, and generated documentation unless the user explicitly expands scope.

## Classification

Classify every candidate as one of:

- `active`: current authority needed in its existing location;
- `historical`: superseded or version-bound material suitable for archive;
- `duplicate`: redundant copy after a canonical document is identified;
- `unknown`: insufficient status, version, ownership, or reference evidence;
- `protected`: operational, legal, security, governance, entry-point, or tool-referenced documentation;
- `generated` or `vendor`: not project-authored source documentation.

Default action is `keep`. Archive only `historical` or approved `duplicate` candidates. Root README, LICENSE/NOTICE, CONTRIBUTING, SECURITY, CODEOWNERS, current CHANGELOG, active ADR/index, API schemas, runbooks, and files referenced by builds, CI, package metadata, source imports, links, or agent rules are protected until proven otherwise.

## Version assignment

Prefer, in order: explicit release/version metadata in the document; Git tag containing the document revision; accepted release records; versioned directory/name; user confirmation. Commit dates alone can order documents but cannot establish a product version. Use `unversioned` or `unknown` instead of guessing.

Archive targets use:

```text
.ai-flow/archive/legacy-documents/<version-bucket>/<original-relative-path>
```

Preserve the original relative path beneath the bucket to avoid filename collisions. Reject absolute paths, `..`, symlink escapes, archive destinations outside `.ai-flow/archive/legacy-documents/`, and any mapping whose target already contains different bytes.

## Summary and recoverability

Write `.ai-flow/baseline/workspace-document-inventory.json` before mutation. Record source path, SHA-256, classification, action, version bucket/evidence, reason, approval, target path, and later the result. For tracked files prefer `git mv` so history remains visible. For untracked files, verify the copied/moved target hash before removing the source; if recovery is not assured, keep the source and mark the entry `skipped`.

Produce a concise summary of current authority, historical themes by version, contradictions, duplicates, unknowns, and information promoted into the current baseline. Do not copy old prose wholesale into active context.
