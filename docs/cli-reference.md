# flowctl CLI 参考

默认从当前目录向上寻找 `.ai-flow/`。所有命令都可用 `--root PATH` 指定目标项目。

## 全局命令

### version

```bash
flowctl version
```

显示运行时版本。

### doctor

```bash
flowctl doctor --root .
flowctl doctor --root . --json
```

分别检查 Codex `.agents/skills/`、Cursor `.cursor/skills/`、Claude Code `.claude/skills/` 中的 15 个 Skills，以及三平台入口、JSON Schema 和初始化状态。

### status

```bash
flowctl status --root .
flowctl status --root . --json
```

读取 `.ai-flow/project.yaml` 和 `.ai-flow/state/current.yaml`。

### validate

```bash
flowctl validate --root .
flowctl validate --root . --json
```

执行：

1. Draft 2020-12 JSON Schema 验证。
2. `.ai-flow/baseline/engineering-profile.json` 工程画像验证（文件存在时）。
3. `.ai-flow/baseline/workspace-document-inventory.json` 文档盘点、审批和恢复映射验证（文件存在时）。
4. `.ai-flow/baseline/workspace-structure-inventory.json` 多语言组件图和初始化候选标记验证（文件存在时）。
5. `.ai-flow/workspace-cleanup/*.json` 清理动作、组件/批次引用、逐路径审批摘要、当前内容指纹、恢复方式和最终验证记录（文件存在时）。
6. Goal/Requirement/Work Item/Run/Checkpoint/Evidence 链接验证。
7. Evidence 日志路径引用检查。
8. Event JSONL 逐行验证。

### cleanup digest

```bash
flowctl cleanup digest --root . --plan .ai-flow/workspace-cleanup/PLAN-<id>.json
```

计算待用户确认内容的稳定 SHA-256。摘要覆盖计划版本、Git revision、结构清单内容摘要、组件边界文件指纹、开发任务、范围、每条源路径与内容指纹、动作、目标、恢复方式、执行批次和风险摘要；执行状态、结果和完成时间不参与计算。用户确认后把输出写入计划的批准记录，任何受批准内容发生变化都会使校验失败。

### render-board

```bash
flowctl render-board --root .
```

从机器状态重新生成 `docs/board/`。

生成内容：

- `STATUS.md`：自然语言项目摘要、当前大版本、子版本进度、开发任务、测试项、证据、阻塞和下一步。
- `ROADMAP.md`：当前目标、目标版本、里程碑、任务完成度和退出门禁。
- `CURRENT_STATE.md`：当前需求、技术方案决策、语言、框架、代码组织方式、适用的开发与测试规范、边界和风险。
- `RELEASES.md`：真实 Release 对象、版本变更、测试证据、已知问题、迁移和回滚。

缺少事实时显示“未记录”“待确认”或“尚无证据”，不会猜测通过状态。

## 项目初始化

```bash
flowctl project init --root . --mode greenfield --name "Project Name"
flowctl project init --root . --mode existing --name "Project Name"
```

参数：

- `--mode`：`greenfield` 或 `existing`。
- `--name`：项目名称。

命令是幂等的：已有核心状态文件不会被覆盖。

## Work Item

### create

```bash
flowctl work create \
  --root . \
  --kind bug \
  --title "Prevent duplicate payment callback" \
  --priority high \
  --goal GOAL-20260817-9876abcd \
  --requirement REQ-20260817-1234abcd \
  --acceptance "The same provider event is applied once" \
  --scope "services/payments/**"
```

可重复参数：

- `--requirement`
- `--acceptance`
- `--scope`

`--kind`：`feature`、`bug`、`refactor`、`test`、`docs`、`research`、`chore`。

`--status`：`draft` 或 `ready`。

命令输出新 Work Item ID。

### list

```bash
flowctl work list --root .
flowctl work list --root . --status in_progress --json
```

### show

```bash
flowctl work show --root . --id WI-20260817-ab12cd34
```

### ready

```bash
flowctl work ready --root . --id WI-20260817-ab12cd34 --expect-revision 2
```

把 `draft` 或 `blocked` Work Item 转为 `ready`。

### start

```bash
flowctl work start \
  --root . \
  --id WI-20260817-ab12cd34 \
  --owner "agent:codex-session-1" \
  --ttl 60m \
  --max-elapsed-minutes 120 \
  --max-retries 3 \
  --max-changed-files 40
```

创建 Harness Run 和 lease，输出 Run ID。

### block

```bash
flowctl work block \
  --root . \
  --id WI-20260817-ab12cd34 \
  --reason "Waiting for API contract confirmation"
```

### review-ready

```bash
flowctl work review-ready --root . --id WI-20260817-ab12cd34
```

### complete

```bash
flowctl work complete \
  --root . \
  --id WI-20260817-ab12cd34 \
  --evidence EV-20260817-a1b2c3d4
```

Evidence 必须属于该 Work Item，结果为 `passed`，并且 trust 不能是 `unverified`。

### cancel

```bash
flowctl work cancel \
  --root . \
  --id WI-20260817-ab12cd34 \
  --reason "Requirement was withdrawn"
```

## Checkpoint

### save

```bash
flowctl checkpoint save \
  --root . \
  --run RUN-20260817-1122aabb \
  --phase implementing \
  --summary "Added idempotency key storage" \
  --next "Run payment integration tests" \
  --completed "Added migration" \
  --completed "Added repository method" \
  --changed-file "services/payments/store.go" \
  --question "Confirm retention period"
```

命令输出 Checkpoint ID。

### list

```bash
flowctl checkpoint list --root . --run RUN-20260817-1122aabb
flowctl checkpoint list --root . --run RUN-20260817-1122aabb --json
```

### show / latest

```bash
flowctl checkpoint show \
  --root . \
  --run RUN-20260817-1122aabb \
  --id CP-20260817-3344ccdd

flowctl checkpoint latest --root . --run RUN-20260817-1122aabb
```

### resume

```bash
flowctl checkpoint resume \
  --root . \
  --run RUN-20260817-1122aabb \
  --owner "agent:codex-session-1" \
  --ttl 60m
```

默认要求当前 Git SHA 与最新 Checkpoint 一致。明确接受漂移时添加 `--allow-git-drift`。

## Evidence

### run

```bash
flowctl evidence run \
  --root . \
  --work WI-20260817-ab12cd34 \
  --run RUN-20260817-1122aabb \
  --test TEST-PAYMENT-IDEMPOTENCY \
  -- go test ./services/payments/...
```

`--` 后的内容作为命令和参数直接执行，不经过 Shell 字符串拼接。命令输出同时显示并写入日志；添加 `--quiet` 只写日志。

成功或失败都会创建 Evidence。命令失败时 CLI 返回非零退出码，并在错误信息中给出 Evidence ID。

### record

```bash
flowctl evidence record \
  --root . \
  --work WI-20260817-ab12cd34 \
  --run RUN-20260817-1122aabb \
  --test EXTERNAL-UAT \
  --source external \
  --uri "https://example.test/report/123" \
  --description "Manual UAT report"
```

手动记录始终为 `unverified`，不能单独支持 Work Item 完成。

### list / show

```bash
flowctl evidence list --root . --work WI-20260817-ab12cd34
flowctl evidence list --root . --result passed --json
flowctl evidence show --root . --id EV-20260817-a1b2c3d4
```

### verify

```bash
flowctl evidence verify --root . --id EV-20260817-a1b2c3d4
flowctl evidence verify --root . --id EV-20260817-a1b2c3d4 --require-current-git
```

重新计算日志 SHA-256，并报告 Evidence 记录的 Git SHA 是否等于当前 Git HEAD。

## 乐观并发

Work Item 状态转换、Checkpoint 保存和恢复支持 `--expect-revision N`。如果对象当前 revision 不等于 N，命令失败，调用者必须重新读取并合并，不允许最后写入者静默覆盖。

## 退出码

- `0`：命令成功；Evidence 测试命令通过。
- `1`：校验失败、状态转换不允许、revision 冲突、lease 冲突、Evidence 命令失败或文件错误。
- `2`：主命令或参数用法错误。
