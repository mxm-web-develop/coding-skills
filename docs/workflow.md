# 初始化与完整使用流程

## 1. 项目入口

安装后，三个 IDE 的入口不同，但业务流程相同。

| 平台 | 常驻入口 | Skill 入口 |
| --- | --- | --- |
| Cursor | `.cursor/rules/ai-flow.mdc` | `.cursor/skills/`，可由 Agent 自动选择或 `/skill-name`；`.agents/skills/` 作为兼容发现目录 |
| Codex | `AGENTS.md` | `.agents/skills/`，可由 Agent 自动选择或 Skill 选择器 |
| Claude Code | `CLAUDE.md` | `.claude/skills/`，可自动选择或 `/skill-name`，并保留 `/ai-flow` 总入口 |

用户应优先用自然语言，不需要手动编排 15 个 Skill。

### 1.1 对话和反馈原则

AI Flow 内部会使用精确的任务、状态和验证记录，但不会要求用户学习这些名字。面向用户的提问、选择、长任务进度、失败原因和完成汇报默认只说明：现在要达成什么、已经完成什么、测试结果如何、有什么风险、下一步做什么，以及是否需要用户决定。

例如，内部可以记录三个阶段和七个可独立验证的任务，但不会说：

```text
按 skill 推荐拆 3 milestones / 7 vertical slices，采用 3MS+7WI plan。
```

用户看到的是：

```text
建议分成三个阶段，共七项开发任务。每项完成后都能单独检查效果，出现问题也不会影响整批工作。下面先给你看整体安排，确认后再开始开发。
```

Skill 名称、对象编号、状态值、缩写、revision、hash 和 `.ai-flow` 内部路径仅在用户主动要求审计、排错或技术细节时作为补充展示。

安装或更新发生在 IDE 已经打开之后时，先 Reload Window 并创建新 Agent chat。Skill 元数据通常在会话启动时发现，不能用安装前已经存在的聊天判断安装是否成功。

## 2. 空白项目初始化

用户示例：

```text
初始化这个空白项目。V1 目标是一个同时服务买家和卖家的电商应用，并根据行为构建人物画像。
```

流程：

1. `initialize-ai-project` 确认没有既有代码和版本历史。
2. 执行 greenfield 初始化，建立 `.ai-flow` 和看板。
3. `discover-product-goal` 讨论用户、价值、范围、非目标、约束和验收。
4. 用户确认 Goal 和 Requirements。
5. `plan-product-delivery` 拆成里程碑和可验证 Work Items。
6. 技术栈确认后，`profile-project-engineering` 建立工程画像和执行 Playbook；确认前不创建业务目录。

## 3. 既有项目接管

用户示例：

```text
把这个已经开发过的项目接入 AI Flow。先告诉我当前版本、已实现能力、测试情况和文档冲突，不要修改代码。
```

流程：

1. `initialize-ai-project` 识别 existing 模式。
2. `adopt-existing-project` 只读扫描 Git、tag、目录、依赖、测试、CI、版本和文档。
3. 如果发现散落历史文档，询问用户选择 `keep`、`audit-only` 或 `summarize-and-archive`。
4. 输出分为 `observed`、`inferred`、`user-confirmed`。
5. 用户确认当前版本和项目基线；归档模式还需要另一次逐路径批准。
6. 基线和文档清单写入 `.ai-flow/baseline/`，批准的历史资料按版本进入 `.ai-flow/archive/legacy-documents/`。
7. `profile-project-engineering` 根据清单、锁文件、配置、CI、代码布局和现有测试生成工程画像。
8. `discover-product-goal` 和 `plan-product-delivery` 才开始讨论基于当前版本的新需求与开发计划。

Agent 不得在接管过程中修改业务代码，也不能把推断当成事实。

初始化还会生成 `.ai-flow/baseline/workspace-structure-inventory.json`，记录语言、子项目、组件依赖、共享目录、嵌套仓库、构建/部署入口、生成目录和疑似废弃的非文档内容。这些条目统一为 `initialization_action: mark-only`：只提供后续调查线索，不代表已经废弃，也不授权移动或删除。

### 3.1 工作区文档盘点与归档

初始化对既有或文档密集型项目提供三种选择：

- `keep`：保持现有文档位置，不建立移动计划。
- `audit-only`：只生成当前权威、历史、重复、冲突、未知和受保护文档清单。
- `summarize-and-archive`：先完成相同的只读盘点，再让用户批准精确的源路径 → 目标路径映射。

初次说“帮我清理”不等于授权移动。只有分类为历史或已确认重复的文件才能进入归档；当前 README、LICENSE/NOTICE、CONTRIBUTING、SECURITY、CODEOWNERS、当前 CHANGELOG、活动 ADR/索引、运行手册和被构建/CI/工具引用的文档默认受保护。

归档结构保留原相对路径：

```text
.ai-flow/archive/legacy-documents/<version-bucket>/<original-relative-path>
```

版本优先依据文档元数据、Git tag、已确认发布记录、版本目录/文件名和用户确认。只有提交日期时使用 `unversioned` 或 `unknown`，不能猜版本。`.ai-flow/baseline/workspace-document-inventory.json` 保存原路径、目标路径、SHA-256、版本依据、批准人和执行结果，可用于恢复和审计。

完成后，看板只展示当前权威、各历史版本摘要、冲突和未知数量；不会把旧文档全文重新放回活跃上下文。

### 3.2 初始化后的项目工作区清理

只有项目初始化完成，并且用户明确提出类似下面的请求时，才触发 `clean-project-workspace`：

```text
清理项目工作区，把已经确认废弃的代码和目录归档，并清除可以重新生成的构建产物。
```

清理流程会重新扫描当前 Git revision，而不是直接执行初始化时的候选标记：

1. 识别 VCS 边界、嵌套仓库、monorepo/workspace、语言、子项目和共享依赖。
2. 检查 import、构建、测试、CI、部署、运行时配置、代码生成、插件注册和用户确认。
3. 写入 `.ai-flow/workspace-cleanup/PLAN-<id>.json`，但不立即修改文件。
4. 按子项目展示精确路径、动作、风险、恢复方式和验证命令，等待用户批准计划 revision 和映射摘要。
5. 按依赖批次执行；每批失败立即停止并恢复，全部通过后刷新基线、工程画像和看板。

确认废弃的源码进入 `.ai-flow/archive/legacy-code/<version>/`，其他历史文件进入 `.ai-flow/archive/legacy-files/<version>/`。生成物和缓存不作为历史源码归档，只有在存在 Git、重新生成或隔离恢复方式时才能清除。共享目录、未知语言、嵌套仓库、密钥、迁移、基础设施和有未提交重叠修改的路径默认保留。

### 3.3 工程画像与社区 Skill

`.ai-flow/baseline/engineering-profile.json` 记录真实语言、框架、架构边界、构建/质量命令、测试系统、视觉验证要求和选中的 Playbook。清单、锁文件、框架配置、CI 或测试工具变化后，技术任务开始前必须刷新画像。

社区 Skill 是可选增强：只引用当前 IDE 已安装且与画像匹配的 Skill，并记录名称、来源、版本、理由和信任级别。任务执行中不会静默下载第三方 Skill；项目自身规则、现有框架和已接受决策始终优先。

## 4. 状态查询短路径

用户示例：

```text
项目现在做到哪个版本？下一步是什么？测试是否全部通过？
```

流程：

1. Orchestrator 调用 `flowctl status --json`。
2. 读取 active Work Items、最新 Evidence 和 Release。
3. 直接回答，不创建开发 Run。

状态回答至少区分：

- 当前记录版本。
- active Goal。
- 已完成、进行中、阻塞的 Work Items。
- verified、failed、unverified 测试结论。
- 下一动作和所需批准。

## 5. 新功能流程

```text
目标/需求
  → 计划和 Work Items
  → 工程画像与技术栈 Playbook
  → 调研与设计决策
  → 测试规格
  → 开始 Work Item 和取得 lease
  → 实现
  → Checkpoint
  → Evidence
  → Review
  → Git integration
  → Release
  → Knowledge sync
```

用户可以说：

```text
为卖家增加商品批量导入。从需求讨论开始，测试先行，完成后准备提交但不要 push。
```

## 6. Bug 修复流程

用户可以说：

```text
支付回调有时会重复入账。先写能证明问题的测试，再定位根因和修复。
```

流程：

1. 创建最小 Bug Work Item。
2. `specify-tests` 写或找到能正确复现症状的失败测试。
3. `diagnose-and-verify` 最小化复现、提出假设并定位根因。
4. 如果设计本身错误，返回 `research-and-design-solution`，而不是只改代码。
5. 修复后运行复现、受影响测试和回归测试。
6. Evidence 必须记录命令、exit code、Git SHA 和日志哈希。
7. Review 通过后进入 Git 和修复版本流程。

### 6.1 按技术栈选择测试

- TypeScript/Node、Python、Go、Rust、JVM/.NET 和移动端分别使用匹配的测试 Playbook。
- 优先复用项目现有 runner、CI 命令和覆盖规则，不为了偏好并存第二套框架。
- 核心领域规则优先使用快速、确定性的单元测试；外部边界使用集成/契约测试；关键用户路径才进入 E2E。
- Web UI 变更需要三层证据：功能 E2E、稳定状态的 Playwright 截图回归、人工/AI 打开截图进行视觉设计审查。
- 截图基线只能在设计变更被确认后更新；失败应保留 actual/expected/diff 和 trace，而不是直接接受新基线。

## 7. Work Item 状态机

```text
draft → ready → in_progress → ready_for_review → done
                    ↘ blocked → ready
draft/ready/in_progress/blocked/ready_for_review → cancelled
```

规则：

- `start` 创建 Run 和带 TTL 的写 lease。
- 相同 Work Item 的有效 lease 存在时，第二个 Agent 不能启动写入。
- 状态修改支持 `--expect-revision`，revision 不一致会停止更新。
- `done` 要求至少一个属于该 Work Item 的可信 passed Evidence。

## 8. 长任务与 Checkpoint

长任务开始后应在以下时机保存 Checkpoint：

- 阶段切换。
- 上下文即将压缩。
- 等待外部输入或审批。
- 用户暂停。
- 准备重试。
- 当前会话即将结束。

Checkpoint 包含：

- Run、Work Item 和序号。
- 当前 phase。
- 已完成内容和变更文件。
- 下一条可直接执行的动作。
- Open questions。
- 当前 Git SHA。

恢复时默认要求 Git SHA 与 checkpoint 一致。只有明确理解漂移后才能使用 `--allow-git-drift`。

## 9. Evidence 信任模型

| 创建方式 | trust | 能否单独支持 Work Item 完成 |
| --- | --- | --- |
| `flowctl evidence run` 实际执行命令 | `verified-local` | 可以，前提是结果 passed |
| 手动 `evidence record` 外部链接/Agent 描述 | `unverified` | 不可以 |
| 未来经过可信 CI 导入 | `verified-ci` | 可以 |

`evidence verify` 会重新计算日志 SHA-256；可选要求 Evidence Git SHA 必须等于当前 HEAD。

## 10. Review 与 Git

Review 必须固定 base/head revision，检查：

- Requirement 和 acceptance criteria。
- 设计一致性。
- 模块职责、长代码是否合理拆成目录和多文件、核心逻辑与副作用是否分离、注释是否解释真实约束。
- 错误处理、数据、并发、安全和兼容性风险。
- 技术栈对应的测试覆盖、UI 视觉证据和 Evidence 新鲜度。
- 是否引入未批准的依赖/API/迁移。

Git 默认约定：

```text
feat(profile): add seller persona scoring

Work-Item: WI-20260817-ab12cd34
Requirement: REQ-20260817-1234abcd
Goal: GOAL-20260817-9876abcd
Evidence: EV-20260817-a1b2c3d4
```

默认不自动 push、merge、tag 或发布。

## 11. 版本和小修复

默认 SemVer：

- 大版本目标：`v1.0.0`
- 兼容新能力：`v1.1.0`
- Bug/小修改：`v1.1.1`、`v1.1.2`

发布记录必须引用 Work Items、commits、reviews 和 Evidence，并包含迁移、已知问题和回滚方式。

## 12. 文档同步与旧方案归档

方案发生变化时：

1. 新 Decision 声明 `supersedes`。
2. 旧 Decision 写入 `superseded_by`。
3. 旧对象从 active index 移除，转入 `.ai-flow/archive/`。
4. 看板重新从当前状态生成。
5. 旧方案只保留历史链接，不继续作为当前实现依据。

四份看板面向用户，而不是面向 Schema：

- `STATUS.md` 先说明“当前处于哪个大版本、正在开发哪个小版本、完成/进行中/阻塞多少任务、测试是否通过、下一步是什么”，再展示版本、任务和测试表。
- `ROADMAP.md` 按目标版本展示里程碑、预期结果、任务完成度和完成门禁。
- `CURRENT_STATE.md` 展示当前有效需求、选用了什么开发方案、技术栈、视觉验证要求、边界和风险。
- `RELEASES.md` 展示实际发布版本、包含任务、测试证据、已知问题、迁移和回滚。

状态页只保留当前大版本及仍在活动的任务；旧版本详情进入发布历史，避免看板越用越长。

## 13. 团队与多 Agent

- 一个 Work Item 一个写 owner。
- 推荐一个 Work Item 一个 branch/worktree。
- Scope 重叠的 Work Items 串行处理或重新切片。
- Lease 过期后可以接管，但必须先检查 checkpoint、working tree 和 Git SHA。
- 所有人和 Agent 共用同一套 `.ai-flow` 对象和 Git 追踪 ID。

## 14. 每次完成后的用户汇报

简洁汇报：

- 当前版本和 Goal。
- 完成的 Work Item。
- 实际验证命令和 Evidence ID。
- Review/Git/Release 状态。
- 未解决风险或阻塞。
- 下一步和是否需要用户批准。
