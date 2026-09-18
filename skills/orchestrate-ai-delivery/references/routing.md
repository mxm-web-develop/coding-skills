# Routing table

| Intent | Route |
| --- | --- |
| “汇报下当前开发计划”, status, version, progress, next work, test state | `flowctl context` + current plan; read-only response |
| “调研下 xxx 方案，看看是否适合我们项目”, compare/evaluate only | engineering profile → `research-and-design-solution` in research-only mode; stop at recommendation |
| “添加个新功能 xxx”, clear feature | reuse current goal → smallest task/acceptance criteria → existing confirmed design or necessary discussion → tests → implementation |
| New skill version, record mismatch, “升级后继续之前的计划” | `upgrade-ai-project`; preserve plan/task identities and pending choices |
| User verified the delivered result, or reports acceptance failure | acceptance workflow in `completion-and-memory.md`; pass closes and archives, failure resumes same task |
| “用强模型做计划，换模型执行”, “换个模型继续开发”, executor handoff | `cross-model-execution.md` → executable contract → bounded context → execution or evidence-backed escalation |
| Missing AI Flow state | `initialize-ai-project` |
| Existing code without baseline | `adopt-existing-project` |
| Scattered, stale, duplicate, or conflicting pre-AI-Flow documents | `adopt-existing-project` for read-only inventory and approval → `sync-project-knowledge` for approved mappings |
| Explicit post-initialization code, directory, generated-output, cache, or workspace cleanup | `clean-project-workspace` → verification/recovery evidence → `sync-project-knowledge` |
| New large goal or unclear feature | `discover-product-goal` → `plan-product-delivery` |
| Approved feature work | `profile-project-engineering` when stale → reuse approved design or confirm material new technical/UX choices in `research-and-design-solution` → `specify-tests` → `implement-work-item` |
| Bug or failing test | `profile-project-engineering` when stale → `specify-tests` when no reproduction exists → `diagnose-and-verify` |
| Review request | `review-change` |
| Commit, PR, or merge preparation | `integrate-git-change` |
| Version or release request | `manage-release` |
| State, report, dashboard, or archive drift | `sync-project-knowledge` |

Never skip verification, review, Git traceability, or knowledge synchronization for a mutation. Small changes may combine discussion and design, but must keep those gates.
