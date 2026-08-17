# Routing table

| Intent | Route |
| --- | --- |
| Status, version, progress, next work, test state | `flowctl status`; read-only response |
| Missing AI Flow state | `initialize-ai-project` |
| Existing code without baseline | `adopt-existing-project` |
| Scattered, stale, duplicate, or conflicting pre-AI-Flow documents | `adopt-existing-project` for read-only inventory and approval → `sync-project-knowledge` for approved mappings |
| New large goal or unclear feature | `discover-product-goal` → `plan-product-delivery` |
| Approved feature work | `profile-project-engineering` when stale → `research-and-design-solution` → `specify-tests` → `implement-work-item` |
| Bug or failing test | `profile-project-engineering` when stale → `specify-tests` when no reproduction exists → `diagnose-and-verify` |
| Review request | `review-change` |
| Commit, PR, or merge preparation | `integrate-git-change` |
| Version or release request | `manage-release` |
| State, report, dashboard, or archive drift | `sync-project-knowledge` |

Never skip verification, review, Git traceability, or knowledge synchronization for a mutation. Small changes may combine discussion and design, but must keep those gates.
