# 初始化与完整使用流程

## 1. 项目入口

安装后，三个 IDE 的入口不同，但业务流程相同。

| 平台 | 常驻入口 | Skill 入口 |
| --- | --- | --- |
| Cursor | `.cursor/rules/ai-flow.mdc` | `.agents/skills/`，可由 Agent 自动选择或 `/skill-name` |
| Codex | `AGENTS.md` | `.agents/skills/`，可由 Agent 自动选择或 Skill 选择器 |
| Claude Code | `CLAUDE.md` | `/ai-flow` 薄入口，再读取 `.agents/skills/` |

用户应优先用自然语言，不需要手动编排 13 个 Skill。

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
6. 尚未确认技术栈前，不创建业务目录。

## 3. 既有项目接管

用户示例：

```text
把这个已经开发过的项目接入 AI Flow。先告诉我当前版本、已实现能力、测试情况和文档冲突，不要修改代码。
```

流程：

1. `initialize-ai-project` 识别 existing 模式。
2. `adopt-existing-project` 只读扫描 Git、tag、目录、依赖、测试、CI、版本和文档。
3. 输出分为 `observed`、`inferred`、`user-confirmed`。
4. 用户确认当前版本和项目基线。
5. 基线写入 `.ai-flow/baseline/`。
6. 之后才讨论新的 Goal 和计划。

Agent 不得在接管过程中修改业务代码，也不能把推断当成事实。

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
- 错误处理、数据、并发、安全和兼容性风险。
- 测试覆盖和 Evidence 新鲜度。
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
