# Human board contract

The human board is a generated management view, not a second fact source. Follow the [user communication contract](../../orchestrate-ai-delivery/references/user-communication-contract.md). Keep the four management boards (each leads with a short natural-language explanation before tables), one version plan index, and one per-version plan document per un-released version. The plan index and plan documents use the same natural-language rules as the four boards.

## STATUS.md

Show:

- one paragraph explaining the current major version, active development version, product goal, completed/active/blocked task counts, verification result, and next action;
- a compact overview table using ordinary project language;
- child-version progress under the current major version;
- current-major or still-active development tasks with owner and test result;
- test checks with type, purpose, linked task, preparation status, and execution result;
- blockers and the next action.

Do not flood this page with completed tasks from older major versions; those belong in release history.

## ROADMAP.md

Explain the active goal in product language. Show target versions, development stages, expected outcomes, task completion, state, and completion conditions. Avoid implementation detail that belongs in internal task records or decisions.

## CURRENT_STATE.md

Show the current version and phase, confirmed needs for the active goal, current architecture/technical decisions, recommendation reasons and confirmation state, detected languages/frameworks/code organization, applicable development/testing approach, visual-check expectations, boundaries, risks, and unresolved technical facts. When an active UX/UI decision has HTML explorations, show a compact preview table with plain-language differences and safe clickable links.

## RELEASES.md

Render actual releases in descending semantic-version order. For each release show status, change summary, linked task count, test result, known issues, migration, and rollback. Never print “no releases” when a release record exists.

## PLANS.md (version plan index)

`docs/board/PLANS.md` is a single-page index that lists every plan that targets a not-yet-released version. It is a quick "what are we about to ship and how far along is it" view, written for the dev team, PM, and stakeholders. It is **not** a release history; shipped work belongs in `RELEASES.md`.

The page leads with one short paragraph naming the currently active version, the next planned version, and how many versions are listed. Then a compact table with one row per un-released plan. Columns, in plain language:

- 版本 (target version, e.g. v0.5.0)
- 本版目标 (one-line description of what this version is for, in product language)
- 负责人 (owner, or 未分配)
- 完成度 (completed task count over total, e.g. 4 / 9)
- 状态 (one of 规划中 / 开发中 / 复核中 / 已完成 / 已搁置; never raw enum)
- 方案 (clickable link to the matching per-version document under `docs/board/plans/v<version>.md`)

Skip rows for plans that cannot be associated with a target version. Skip rows for already-released versions; those live in `RELEASES.md`. Never print "no plans" when at least one un-released plan exists.

## Per-version plan documents (docs/board/plans/v<version>.md)

For every plan listed in `PLANS.md`, render one Markdown document at `docs/board/plans/v<version>.md`. These documents are written for human readers: PM, dev team, and stakeholders who do not work with AI Flow day to day. The same natural-language rules as the four boards apply — no raw IDs, no internal object names, no shorthand state values in the main text. IDs may live in hidden HTML comments for traceability but must never be required to understand the document.

Each document uses this fixed section shape, in this order. Use a short natural-language heading for every section. Skip any section that has no content; do not leave empty placeholders.

1. Lead paragraph — one sentence each, in plain language:
   - 本计划面向 (which goal / product outcome this plan is for)
   - 要解决的问题 (the user-facing problem this version is meant to solve)
   - 完成后能提供 (the concrete capability or experience that becomes available once this version is done)
2. 范围内 — bullet list of what is in scope for this version, in product language. Source: the goal's in-scope items.
3. 不在范围内 — bullet list of explicit non-goals. Source: the goal's non-goal items. Helps the team avoid scope creep.
4. 验收要点 — bullet list of acceptance criteria the team will use to call this version done. Source: the goal's acceptance criteria.
5. 阶段划分 — one subsection per stage. For each stage:
   - numbered heading (e.g. `### 1. <stage title>`)
   - 完成后能看到: one sentence describing the user-visible outcome of completing this stage
   - 完成条件: bullet list of exit conditions the team will check
   - 本阶段包含的开发任务: short bullet list (or compact table) of the development tasks assigned to this stage, each described by what it does, not by its ID
6. 开发任务清单 — one compact table covering every development task in this plan, regardless of stage. Columns: 任务 (what the task does), 类型 (one of 新功能 / 修复 / 重构 / 体验 / 验证, never raw enum), 阶段 (which stage it belongs to), 状态 (one of 规划中 / 开发中 / 复核中 / 等待 / 已完成 / 已搁置), 负责人 (or 未分配). Each row carries its machine ID in a hidden HTML comment, never in the visible text.
7. 技术选型 — compact comparison table of the major technical choices for this version. Columns: 方向 (the choice area in plain language, e.g. 前后端通信方式), 推荐方案 (one-sentence description, no raw ADR IDs), 备选 (one-sentence description of the main alternative considered), 取舍要点 (one short line on why the recommended option was chosen). Rows cover the choices that materially affect this version; do not enumerate every ADR.
8. 风险与依赖 — bullet list. Each item: 风险或依赖 (one-line description in plain language), 影响 (which stage or capability is affected), 缓解 (one short line on how the team plans to deal with it). Source: the plan's risk entries and any cross-version dependencies surfaced from related goals.
9. 相关材料 — closing footer. Plain-language links to the matching row in `PLANS.md`, the goal in product language (`docs/board/CURRENT_STATE.md`), and any UX exploration previews already approved for this version.

## Language and accuracy

- Use short natural-language labels for machine states and phases.
- Do not display internal object names, raw IDs, state values, Skill names, Playbook names, abbreviations, hashes, machine directories, or storage implementation notes in the main board.
- Prefer “开发阶段”“开发任务”“测试结果”“完成条件”“技术环境”“开发与测试规范” over Milestone, Work Item, Evidence, Gate, engineering profile, or Playbook.
- Preserve exact machine-object traceability in non-rendered HTML comments generated with the board. Do not show raw identifiers or machine paths in the rendered page, and never require them to understand status. A validated clickable preview link is allowed when the path is required for the user to open an HTML experience exploration.
- Use “未记录”“待确认” or “尚无证据” when facts are absent; never infer a positive result.
- Escape table delimiters and collapse multiline text inside cells.
- Keep archived/superseded objects out of current tables.
- Write all generated files atomically from validated `.ai-flow/` objects: the four boards, `PLANS.md`, and every per-version plan document under `docs/board/plans/`.
- Generated boards must not show bare section or phase numbers as the primary text. If a board entry needs to point at a source document (e.g. a per-version plan document, an external spec, a chapter in a requirements doc), restate the section's content in plain language and attach a clickable Markdown link to the document location; do not write `§N`, `Phase N`, `Module N`, or `doc §N` in the rendered text. Raw IDs and section numbers may stay in the non-rendered HTML trace comments for machine linkage.
