# Workspace cleanup contract

This file defines internal safety records. Translate every question, approval and result through the [user communication contract](../../orchestrate-ai-delivery/references/user-communication-contract.md); never require the user to interpret the names below.

## Two gates

Discovery and mutation are separate phases.

1. Write a proposal for the current Git SHA and inventory revision.
2. Run `flowctl cleanup digest --root <root> --plan <plan-path>` and obtain approval for the exact plan revision and returned SHA-256 digest.
3. Before execution, reject stale plans, changed fingerprints, new overlapping worktree changes, or changed component boundaries.

Bind the plan to a canonical digest of the confirmed workspace structure inventory and to SHA-256 fingerprints for workspace manifests plus every scoped component's manifests, CI and deployment files. Recheck them while the plan is proposed/approved and immediately before execution; an uncommitted boundary edit invalidates approval even when `HEAD` is unchanged.

Initialization-time labels are leads, not evidence of deprecation and not approval.

## Classification and actions

Use these defaults:

| Classification | Allowed default action |
| --- | --- |
| `active`, `protected`, `unknown`, `shared` | `keep` |
| `deprecated-confirmed` | `archive-code` or `archive-file` |
| `duplicate-confirmed` | `archive-code` or `archive-file` after proving the retained authority |
| `generated`, `cache`, `obsolete-artifact` | `remove-generated` only with a recovery method |

An ignore-rule change is a separate `add-ignore-rule` item with the exact `ignore_pattern`, whether the target ignore file existed, its approved before fingerprint when present, and its expected final fingerprint. It does not authorize removing existing files. Any concurrent target edit invalidates the approved plan.

## Archive and recovery targets

Preserve original relative paths beneath a version bucket:

```text
.ai-flow/archive/legacy-code/<version-bucket>/<original-relative-path>
.ai-flow/archive/legacy-files/<version-bucket>/<original-relative-path>
```

Reject absolute paths, `..`, symlink escapes, target conflicts, moves across nested repositories, and paths outside the inspected VCS root. Archived content is historical and must be excluded from active Agent context, compilation, packaging, testing and deployment.

Every mutating item declares one recovery method:

- `git`: tracked bytes and history can be restored from the recorded commit plus reverse mapping.
- `archive`: target fingerprint matches the source fingerprint and the reverse path is free.
- `regenerate`: an already-verified deterministic command recreates disposable output.
- `quarantine`: bytes are first moved to a plan-specific recovery location.

## Multi-component execution

Order batches by the component dependency graph. Process leaf consumers before shared providers unless an accepted migration plan requires another order. Strongly connected components and shared roots form one atomic batch.

Each batch records:

- exact cleanup item IDs;
- affected components and languages;
- preconditions and working-tree scope;
- before/after verification commands;
- verification record IDs for every declared before/after command, with recovery records when a batch is restored;
- a mutation start/end window; before and after/recovery records must be distinct and fall on the correct side of that window. A restored batch still requires passing before checks, a trusted failing after check, and passing recovery checks;
- stop and recovery behavior.

Do not claim repository-wide success when a language or component could not be verified. Record it as an explicit residual risk.

## Completion

After all approved batches:

1. Verify fingerprints and archive targets.
2. Run root-level integration, packaging and deployment/configuration checks when available.
3. Record command Evidence at the final Git SHA.
4. Refresh workspace inventory and engineering profile.
5. Render the human board with a short cleanup summary, recovery route and remaining candidates.

The final plan records its cleanup Work Item, `completed_at`, `completed_git_sha`, and at least one `verification_evidence_id`. Each verified batch must link every declared command to a matching verification record; final records must belong to that Work Item, be trusted and passing, match the expected pre/post Git revision, and retain an unchanged log. Every mutating item must have its own approved path mapping, verified recovery reference, final result, current content fingerprint, and exactly one execution batch; aggregate approval never substitutes for path approval.
