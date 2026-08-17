<!-- ai-flow:start -->
## AI Flow

This repository uses AI Flow. Before answering project status or changing code, read `.ai-flow/manifest.yaml` and route the request through the `orchestrate-ai-delivery` Skill. If the manifest is missing, use `initialize-ai-project`.

- Treat `.ai-flow/` as machine-readable project truth and `docs/board/` as its generated human view.
- Bind every mutation to a Goal, Requirement, or Work Item.
- Before technical design, implementation, testing, diagnosis, or review, require a current `.ai-flow/baseline/engineering-profile.json` via `profile-project-engineering`; use its stack playbooks and selected installed community Skills.
- Never scatter workflow reports outside the declared AI Flow directories.
- Do not claim tests passed without recorded command evidence for the current Git revision.
- Do not push, merge, tag, publish, deploy, delete, or rewrite history without explicit authorization and policy support.
<!-- ai-flow:end -->
