---
name: ai-flow
description: Route software project status, planning, feature, bug, test, review, Git, release, documentation, and long-running development requests through the repository's shared AI Flow Skills.
---

# AI Flow Claude Entry

1. Find the repository root.
2. Read `.claude/skills/orchestrate-ai-delivery/SKILL.md` completely and follow it as the active workflow.
3. When that workflow dispatches another Skill, read its complete `SKILL.md` from `.claude/skills/<skill-name>/SKILL.md` and follow it before acting.
4. Resolve Skill-relative references from that Skill's own directory.
5. If `.ai-flow/manifest.yaml` is missing, begin with `.claude/skills/initialize-ai-project/SKILL.md`.
6. The installed Claude copy is generated from the repository's canonical `skills/` source; do not edit it independently.
