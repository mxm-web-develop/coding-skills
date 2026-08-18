# AI Flow 开发计划

状态：v0.4.0 每版一份人话版开发方案 + 收紧对外表达

更新日期：2026-08-18

计划原则：先确保三套 IDE 共用一个可安装、可恢复、可验证的小闭环，再增加团队并发、远程平台和生产交付复杂度。

## 1. 已完成的 v0.1.x

### 安装与平台入口

- Shell 与 PowerShell 本地安装、更新、卸载。
- GitHub Release bootstrap、一行式远程安装与 SHA-256 校验。
- Cursor、Codex、Claude Code 路由入口；业务规则使用一个源码并安装到各自的原生发现目录。
- 14 个 Core Skill，均通过 Skill 结构校验。

### 项目与流程

- greenfield 与 existing 两种项目初始化模式。
- Orchestrator 自然语言路由、状态查询短路径和完整交付路径。
- Work Item 状态机、revision 乐观并发和单写 lease。
- Harness Run、预算、Checkpoint 保存与 Git SHA 恢复检查。
- Evidence 真实命令执行、日志、退出码、Git SHA、SHA-256 和信任等级。
- Work Item 完成前重新验证可信 passed Evidence 的日志完整性。

### 状态和文档

- 15 个 Draft 2020-12 JSON Schema。
- 对象 Schema、Event JSONL 和跨对象引用校验。
- `.ai-flow/` 机器事实与 `docs/board/` 人读看板分离。
- 看板显示版本、目标、下一动作、Work Item 和 Evidence 摘要。
- Git 提交、版本、替代关系和归档规范进入相关 Skill 契约。

### 发布工程

- Go unit test、vet、Shell 语法、Schema JSON 与 E2E CI。
- Linux/macOS/Windows 的 amd64、arm64 六个交叉编译目标。
- tar.gz、zip、独立二进制、checksums GitHub Release 资产。
- tag `v*.*.*` 触发发布，二进制版本由 tag 注入。

## 2. v0.1.x 验收门禁

- 安装、初始化、创建 Work Item、启动 Run、保存 Checkpoint、执行 Evidence、评审就绪、完成、校验和看板生成必须端到端通过。
- 修改 Evidence 日志后，完成 Work Item 必须失败。
- 所有 Schema 可以编译，所有 Skill 可以通过结构验证。
- 发布包内必须包含 Skills、适配器、Schema、安装器、文档、规格和六个平台二进制。
- 一行式安装必须从真实 GitHub Release 下载并通过 checksum 验证。
- 安装器不能覆盖未管理的同名 Skill；卸载默认保留项目状态与人读看板。

## 3. 已完成的 v0.2.0：技术栈自适应执行层

- `profile-project-engineering` 从仓库证据识别语言、框架、架构边界、质量命令、测试系统和视觉验证要求。
- 工程画像使用独立 JSON Schema，并由 `flowctl validate` 校验。
- 开发规则拆为跨语言模块化基线及 Web/Node、Python、Go、Rust、JVM/.NET、移动端 Playbook。
- 测试规则按技术栈拆分，复用项目现有 runner；Web 增加功能 E2E、Playwright 截图回归和人工/AI 视觉审查。
- 社区 Skill 只引用已安装能力，记录来源、版本、理由和信任等级；禁止任务中静默下载。
- Review 增加模块职责、长文件拆分、纯核心/副作用边界、注释质量和 UI 视觉证据门禁。

验收：技术任务必须使用当前工程画像；所有 Core Skills 和 Schema 通过验证；安装器在三 IDE 安装完整的 14 Skill 执行包。

## 4. 已完成的 v0.2.1：既有工作区文档治理

- 初始化对既有或文档密集型项目询问 `keep`、`audit-only` 或 `summarize-and-archive`。
- 发现和移动分成两个审批阶段；只接受用户批准的精确源路径 → 目标路径。
- 文档按当前、历史、重复、未知、受保护、生成和供应商内容分类。
- 历史文档按有证据的版本归档并保留原相对路径；不能确定版本时进入 `unversioned`/`unknown`。
- `workspace-document-inventory.json` 记录哈希、版本依据、审批、执行结果和恢复映射，并由独立 Schema 校验。
- 归档完成后才基于确认的当前版本、能力和风险讨论新需求与开发计划。

验收：未批准文件不会移动；受保护文件默认保留；目标哈希与源哈希一致；旧资料不再污染活跃上下文但仍可恢复。

## 5. 已完成的 v0.2.2：人读看板与运行时入口优化

- 明确 `main.go` 仅为源码入口，普通用户使用 Release 预编译 `flowctl`，不需要 Go。
- 将看板加载、聚合、渲染和测试从 `main.go` 拆成独立 Go 文件。
- `STATUS.md` 增加自然语言摘要、当前大版本、子版本、开发任务和测试状态表。
- `ROADMAP.md` 增加版本化里程碑、任务进度和退出门禁。
- `CURRENT_STATE.md` 增加当前需求、开发架构决策、工程画像、边界和风险。
- `RELEASES.md` 改为读取真实 Release/Evidence，展示已知问题、迁移和回滚。
- Test Spec 增加组件、无障碍、视觉和冒烟测试类型。

验收：人读看板不要求用户理解对象状态码；大版本/小版本、任务、方案、测试和发布结果均可从表格直接查看；缺失信息不会被推断为通过。

## 6. 已完成的 v0.2.3：部分安装恢复

- 为 Cursor Rule 增加独立的受管版本标记。
- 中心安装标记丢失时，允许接管带 `.ai-flow-managed` 的 Skill 副本。
- 通过稳定 AI Flow 签名识别旧版 `.cursor/rules/ai-flow.mdc`。
- 对无法确认归属的同名 Rule 和 Skill 继续拒绝覆盖。
- Shell 与 PowerShell 安装器采用相同恢复规则。

验收：删除 `.ai-flow/install/version` 并保留旧版 Cursor Rule 后可重新安装；用户自建的同名 Rule 仍保持原样且安装失败。

## 7. 已完成的 v0.2.4：多 IDE 任意顺序接入

- Cursor、Codex、Claude Code 可以任意顺序逐次安装到同一工作区。
- 平台集合采用并集合并，不会因增加新平台而移除旧平台。
- 增加平台时同步刷新全部已登记平台，避免共享运行时与 IDE Skill 副本版本漂移。
- 三端共享唯一运行时、机器状态和人读看板。
- 部分删除后可分别识别旧版 Cursor Rule、Codex Skill Pack 和 Claude 入口。
- 增加平台时重新执行同名文件归属检查，不降低首次安装的冲突保护。

验收：Cursor→Codex→Claude、Claude→Codex→Cursor 和一次性 `--all` 产生相同平台集合；均只有一套 `.ai-flow/` 和 `docs/board/`；新增平台遇到用户自建文件时停止且不覆盖。

## 8. 已完成的 v0.2.5：残留入口识别与安装进度

- 平台探测要求 IDE 入口与完整 14 Skill 原生目录同时存在。
- 旧 `AGENTS.md`、`CLAUDE.md`、Claude 入口或 Cursor Rule 残留不再单独激活平台。
- bootstrap 显示下载、校验、解压和安装阶段。
- 安装器显示目标路径、平台集合、运行时来源、Skill 安装和健康检查阶段。

验收：只残留三端入口、执行 `--cursor` 时仅登记 Cursor 且 doctor 通过；网络下载开始后立即显示阶段进度。

## 8.1 已完成的 v0.2.6：初始化边界与工作区清理

- 初始化/既有项目接管只允许对历史文档执行逐路径批准的版本归档。
- 多语言、monorepo、多子项目、嵌套仓库、共享根、构建/部署入口和非文档候选写入 `workspace-structure-inventory.json`。
- 初始化阶段所有非文档候选固定为 `initialization_action: mark-only`，不允许移动、删除、重命名或合并。
- 新增 `clean-project-workspace`；只有初始化后用户明确请求清理，才重新验证当前 Git revision 并生成清理计划。
- `workspace-cleanup/PLAN-*.json` 记录逐路径 fingerprint、目标文件变更前后 fingerprint、组件归属、依赖证据、动作、风险、恢复方式、审批摘要和多组件执行批次；每个批次的前后检查命令必须关联到时间顺序明确且不可复用的实际执行记录。
- 确认废弃源码和其他历史文件分开归档；生成物/缓存只有具备 Git、重新生成或隔离恢复方式时才能清除。
- 新增统一用户语言契约；规划、初始化、清理、长任务进度、错误和完成反馈不再直接暴露 Skill、Milestone、Work Item、Evidence、Playbook、内部 ID、状态值和缩写。
- 看板隐藏需求/测试对象编号和机器目录，将里程碑、门禁、证据、工程画像等改为开发阶段、完成条件、测试结果和技术环境；精确对象关系保留在不渲染的追踪标记中。

验收：15 个 Core Skills 和 17 个 Schema 全部通过；初始化候选无法表达清理动作；共享/未知/受保护路径无法归档；未经逐路径批准的内容无法执行；完成状态必须关联最终 Git revision 和验证记录；三 IDE 安装后共享相同清理状态。

## 7.4 已完成的 v0.4.0：每版一份人话版开发方案

- 看板渲染时同步产出 `docs/board/PLANS.md`（所有未发布版本的方案索引）和 `docs/board/plans/v<版本>.md`（每份方案文档），已发布版本不再生成对应方案文档。
- 每份方案包含固定小节：本计划面向 / 要解决的问题 / 完成后能提供 / 范围内 / 不在范围内 / 验收要点 / 阶段划分（含"完成后能看到"/"完成条件"/"本阶段包含的开发任务"） / 开发任务清单（任务/类型/阶段/状态/负责人） / 技术选型（方向/推荐方案/备选/取舍要点） / 风险与依赖（风险或依赖/影响/缓解） / 相关材料。
- 方案全文以自然语言表达，不以原始 ID、状态机值、内部短名或 commit SHA 作为主语；精确对象关系仍保留在不渲染的 HTML 注释里供追踪。
- 新增 `expectedPlanFiles` / `planVersionLabel` / `normalizeVersionLabel` / `renderPlanIndex` / `humanPlanState` / `renderPlanDoc` / `milestoneWorkBullets` 等渲染辅助；既有的 `humanStatus` / `humanKind` / `humanOwner` / `mdCell` / `compareVersions` / `writeBoardFile` 等 helper 沿用。
- 全面收紧对外表达：新增 `user-communication-contract.md` 末尾的"禁止漏词表"小节，把 `§2 / §N` 短链、`WI / DEC / REQ / MS / ADR` 简称、`WI-7 / fdd1b619` 内部 ID、`form_decisions / form_field_guide` 模块短名、`in_progress / review / blocked` 状态机值、commit SHA 在用户面话术里作为主语全部列入禁止项，并给出"应说"对照短语 + 三道自检关。
- `sync-project-knowledge` SKILL.md 第 7 步和第 9 步同步更新：说明 `render-board` 现在会同时产出索引和每版方案；`board-contract.md` 新增 PLANS.md 和 per-version plan documents 两章，规定小节结构、语言规则、跳过条件和原子的批量写入。
- 升级 `flowctl` 运行时版本号为 0.4.0，`spec/skill-pack.yaml` 同步。

## 8. 已完成的 v0.3.1：用户面语言收紧

- 收紧 `user-communication-contract.md`：新增 "When you have to ask about scope, plan, or backlog changes" 章节，给出内部标签到用户面语言的完整翻译表，以及两张正反对比范例（含一张原样复刻的"漏出 WI 标签"反例）。
- 在 `implement-work-item`、`plan-product-delivery`、`review-change`、`diagnose-and-verify` 四个 Skill 的 user-communication-contract 引用行后追加显式指引：每次涉及范围/计划/需求/测试范围调整的确认问题，先去读那一节再生成问法。
- 升级 `flowctl` 运行时版本号为 0.3.1，`spec/skill-pack.yaml` 同步。

## 9. 已完成的 v0.3.0：真实交互与追踪闭环

- 技术探索改为用户共同确认：后端方案展示优势、弱点、项目适配、迁移/运维/测试影响、推荐原因和恢复方式。
- 前端方向不明确时提供两到三个隔离的 HTML 体验稿，确认布局、视觉、移动端、交互和动画后才进入生产实现。
- 新增技术选择确认记录、推荐依据、体验稿路径和看板预览入口；旧决策记录保持兼容。
- 每轮对话重新读取项目状态；状态插问、需求补充、独立任务、取消替换、无关问答、恢复和 IDE 切换都有明确连续性规则。
- “继续”或插入问题不再被视为技术选型、发布、移动、删除等待确认操作的批准。
- 跨 IDE 或 Agent 接手必须从已保存位置显式交接同一次运行，任务、进度和写入权不会因编辑器变化而重复。
- `flowctl validate` 完整检查 Goal → Requirement → Plan/Decision → Work Item → Test → Run/Checkpoint/Evidence → Release 的存在、归属、双向链接、替代关系和发布测试覆盖。
- 人读看板只由工具生成；完整校验会识别过期或手工改写的看板，防止内部术语和错误进度回到用户界面。
- 验证日志限制在项目验证目录，拒绝路径穿越；交互式 HTML 路径也限制在受管体验稿目录。
- 增加跨 IDE 中断恢复 E2E，以及空白项目、已有多语言项目和中途补充/切换 IDE 的真实模型旅程。

验收：三条真实模型旅程不暴露内部流程术语；用户每次只面对一个可理解的决定；补充信息只使受影响的后续内容失效；三 IDE 共用同一任务、运行和恢复点；看板手工改写会被阻止；完整发布链断链、归属错误、未通过测试或不安全日志路径都会被阻止。

## 10. v0.4.0：团队协作与迁移

- Work Item scope 重叠检测、lease 续期和 worktree 辅助命令。
- Agent/用户身份、Run、commit、PR 的审计关系。
- GitHub/GitLab PR、Checks、CODEOWNERS 和 ruleset 可选适配。
- Schema 与 Skill 包 N-1 → N 数据迁移及失败回滚。
- Goal、Requirement、Decision、Release 和 archive 的确定性 CLI CRUD。
- 看板增量渲染、发布索引和陈旧性检查。

验收：多个 Agent 不会静默覆盖同一 revision 或重叠 scope；升级失败不丢项目数据；受保护路径缺少批准时不能进入发布门禁。

## 11. v0.5.0：安全与生产交付 Profile

- `secure`：依赖锁、SBOM、密钥扫描、漏洞证据、威胁模型和安全审批。
- `delivery`：环境声明、部署证据、渐进发布、回滚和部署审批。
- `observe`：日志、指标、trace、健康验证、事故记录和恢复报告。
- MCP 只做能力适配；缺少供应商 MCP 时仍提供 CLI 或人工证据降级路径。

## 12. 持续测试策略

- Unit：ID、状态转换、revision、锁、checksum、路径和看板汇总。
- Contract：Skill 的输入、输出、停止条件和允许状态变化。
- Schema：合法对象、工程画像、工作区文档清单、工作区结构清单、清理计划、非法状态、悬空引用和 Event JSONL。
- Stack：各语言/框架 Playbook 路由、社区 Skill 来源记录和项目规则优先级。
- Web：Playwright 功能、截图回归、trace 与人工/AI 视觉审查契约。
- E2E：空白/既有项目、成功、失败、篡改、中断恢复、更新和卸载。
- Platform：三个 IDE 的发现、自然语言路由和无 MCP/Hook/子 Agent 降级。
- Release：六个二进制可执行格式、包内容、checksum 和远程 bootstrap。
- Security：路径穿越、命令参数边界、供应链校验、敏感信息和权限边界。

## 13. Definition of Done

一个版本只有同时满足以下条件才算完成：

- 功能代码、测试、Schema、Skill 契约和用户文档同步更新。
- 至少一个真实端到端场景，不只依赖 Agent 文字输出。
- 失败路径、证据完整性和恢复路径经过验证。
- 三平台公共语义一致，平台差异只存在于 adapter。
- 没有新增散落状态或重复事实源。
- 发布物可校验、可更新、可卸载；外部副作用仍受用户和仓库权限控制。
