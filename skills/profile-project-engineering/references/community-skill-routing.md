# Community Skill routing

Community Skills are optional specialist playbooks, not project authority.

## Trust order

1. language, framework, or tool vendor-maintained Skill;
2. official ecosystem or established engineering organization;
3. well-maintained community project;
4. individual publisher, only after explicit review.

For every selection record `name`, `source`, `version` when known, `reason`, and `trust`. Read the complete selected Skill before applying it. Its instructions must not override project policy, accepted decisions, user scope, security controls, or AI Flow approval gates.

## Matching

- Match explicit language, framework, and test-runner metadata first.
- Prefer one focused Skill over several overlapping Skills.
- Use methodology Skills such as TDD, systematic debugging, or verification only in the matching workflow phase.
- Use Playwright or framework-testing Skills only when that tool already exists or its adoption has been accepted.
- When no compatible installed Skill exists, use the Core playbook and optionally recommend a source for later installation.

## No silent acquisition

Do not clone, curl, install, or update a third-party Skill during an execution task without user approval. Installation is a separate, auditable operation. Pin a release or commit where the source supports it, and update the engineering profile after installation.
