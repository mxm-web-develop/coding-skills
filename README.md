# Coding Skills · AI Flow

AI Flow 是一套运行在 Cursor、Codex 和 Claude Code 内部的 AI 开发流程。它把产品目标、需求、技术决策、测试、实现、验证、Git、版本和项目文档连接成一条可恢复、可审计的工作流。

它不需要独立服务或数据库：机器事实保存在项目根目录 `.ai-flow/`，人读看板保存在 `docs/board/`，所有内容可以跟随 Git 由个人或团队共同维护。

## 核心能力

- 14 个同源 Agent Skills，并安装到三个 IDE 各自可发现的项目目录。
- 基于仓库证据生成工程画像，按真实语言、框架、现有命令和已安装社区 Skill 选择开发与测试 Playbook。
- 模块化代码门禁：按职责拆分多文件、核心逻辑优先纯函数、副作用显式隔离、注释解释意图和约束。
- Web UI 同时要求功能 E2E、Playwright 截图回归和人工/AI 视觉设计审查。
- 空白项目和既有项目两种初始化方式。
- 既有或文档密集型工作区可选择保持原样、只读盘点，或在逐路径批准后按版本总结归档散落历史文档。
- 自然语言自动路由：状态、需求、功能、Bug、测试、评审、Git、版本和文档请求都会进入相应流程。
- Work Item、Harness Run、Checkpoint、Evidence 状态机。
- 测试证据绑定真实命令、退出码、Git SHA、日志和 SHA-256。
- 四份自然语言人读看板，按大版本/小版本展示任务、方案决策、测试状态、发布和下一步。
- JSON Schema 与跨对象链接校验。
- Cursor、Codex、Claude Code 三平台入口。
- macOS、Linux、Windows 的 Release 二进制和一行式安装。
- 安装、更新和卸载不会覆盖未被 AI Flow 管理的同名 Skill。

## 一分钟开始

### macOS / Linux / WSL

进入目标项目根目录后选择当前 IDE：

```bash
# Cursor
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh | sh -s -- --cursor

# Codex
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh | sh -s -- --codex

# Claude Code
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh | sh -s -- --claude
```

可以组合 `--cursor --codex`。也可以先安装任意一个 IDE，之后用相同命令增加另一个；安装顺序不影响结果，已有平台会同步刷新到同一版本，`.ai-flow/` 状态和 `docs/board/` 都会保留。不传平台参数时默认安装全部，兼容旧命令。

### Windows PowerShell

进入目标项目根目录后运行：

```powershell
$env:AI_FLOW_PLATFORMS="cursor" # cursor、codex、claude 或逗号组合
irm https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.ps1 | iex
Remove-Item Env:AI_FLOW_PLATFORMS
```

安装器会：

1. 下载最新 GitHub Release。
2. 校验发布包 SHA-256。
3. 安装适合当前操作系统和架构的 `flowctl`。
4. 将 14 个 Skill 安装到所选 IDE 的原生目录。
5. 只创建所选 IDE 的常驻路由入口。
6. 保留已有 `AGENTS.md`、`CLAUDE.md` 和其他 IDE 规则。
7. 运行健康检查。

用户机器不需要安装 Go。`main.go` 只是源码入口；远程安装默认下载当前系统和架构对应的预编译 `flowctl` 单文件二进制。只有 AI Flow 仓库开发者显式设置 `AI_FLOW_BUILD_SOURCE=1` 从源码验证时才需要 Go。

## 初始化项目

安装和项目初始化是两个步骤。安装只放入工具，不猜测项目状态。

安装后，在 Cursor、Codex 或 Claude Code 中直接说：

```text
初始化这个项目并接入 AI Flow。
```

Agent 会先判断项目属于哪一种模式：

- `greenfield`：空白或尚未开始开发的项目，先讨论第一个产品目标。
- `existing`：已经有代码、Git 历史、依赖、测试或版本记录的项目，先只读扫描并由用户确认基线。

如果既有项目散落着大量历史方案、旧报告或重复文档，Agent 会继续询问：保持原样、只做只读盘点，还是生成精确清单后总结并归档。选择归档不会立即移动文件；必须再次确认每一条源路径和目标路径。批准的历史文档按版本进入 `.ai-flow/archive/legacy-documents/`，当前 README、许可证、当前 ADR、运行手册和工具引用文档默认保留原位。

也可以手动初始化：

```bash
# 空白项目
.ai-flow/bin/flowctl project init \
  --root . \
  --mode greenfield \
  --name "My Project"

# 既有项目
.ai-flow/bin/flowctl project init \
  --root . \
  --mode existing \
  --name "Existing Project"
```

Windows 使用：

```powershell
.\.ai-flow\bin\flowctl.exe project init --root . --mode existing --name "Existing Project"
```

初始化后会生成：

```text
.ai-flow/                 # 机器可读事实、运行、证据和归档
docs/board/               # 给人看的简洁看板
.agents/skills/           # 选择 --codex 时生成
.cursor/skills/           # 选择 --cursor 时生成
.claude/skills/           # 选择 --claude 时生成
.cursor/rules/            # Cursor 常驻路由，仅 --cursor
AGENTS.md                  # Codex 常驻路由，仅 --codex
CLAUDE.md                  # Claude Code 常驻路由，仅 --claude
```

目录名必须是 `.agents/skills`（`agents` 为复数），不是 `.agent`。安装或更新完成后，需要 Reload IDE Window 并开始一个新的 Agent chat，让 IDE 重新发现 Skill 元数据。

## 日常使用

用户不需要记 Skill 名称或 CLI。正常表达意图即可。

```text
项目现在是什么版本？已经完成了哪些功能？
```

```text
下一个版本要增加卖家画像和行为评分，先跟我讨论需求。
```

```text
结算页面偶尔会重复提交订单，复现并修复这个 Bug。
```

```text
继续上次中断的长任务，先告诉我 checkpoint 和剩余风险。
```

```text
所有测试都通过了吗？把真实证据和 Git revision 告诉我。
```

```text
准备提交和发布 v1.1.0，但不要自动 push 或 tag。
```

Orchestrator 会读取仓库状态并选择短路径或完整流程。状态查询不会启动开发任务；任何代码修改都会绑定 Work Item、测试和证据。

## 标准交付流程

```text
目标对齐
  → 需求与计划
  → 工程画像与技术栈 Playbook
  → 技术调研和决策
  → 测试先行
  → Work Item 实现
  → 诊断与真实验证
  → 独立评审
  → Git 集成
  → 版本与发布
  → 机器状态、看板和归档同步
```

小修改可以合并讨论和设计步骤，但不能跳过测试、证据、评审、Git 追踪和知识同步。

## 长任务 Harness

每个长任务使用：

- Work Item：需要交付的最小可验证任务。
- Run：一次具体执行，包含 owner、预算、阶段和 Git SHA。
- Lease：限制同一 Work Item 只有一个写入者。
- Checkpoint：中断前保存已完成步骤、变更文件、下一动作和 Git SHA。
- Evidence：执行真实命令并保存日志、退出码和哈希。

即使切换 IDE、会话中断或上下文被压缩，也可以从 `.ai-flow/runs/` 恢复，而不是依赖聊天记忆。

## 查看项目状态

```bash
.ai-flow/bin/flowctl status --root .
.ai-flow/bin/flowctl status --root . --json
.ai-flow/bin/flowctl doctor --root .
.ai-flow/bin/flowctl validate --root .
```

人读看板位于：

- `docs/board/STATUS.md`
- `docs/board/ROADMAP.md`
- `docs/board/CURRENT_STATE.md`
- `docs/board/RELEASES.md`

看板由机器状态生成，不应直接修改事实。

## 更新

macOS / Linux / WSL：

```bash
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh \
  | AI_FLOW_COMMAND=update sh
```

Windows PowerShell：

```powershell
$env:AI_FLOW_COMMAND="update"
irm https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.ps1 | iex
Remove-Item Env:AI_FLOW_COMMAND
```

固定版本安装：

```bash
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh \
  | AI_FLOW_VERSION=v0.2.5 sh
```

## 卸载

需要使用本地安装脚本或下载发布包后执行：

```bash
./install/install.sh uninstall --target /path/to/project --source .
```

```powershell
.\install\install.ps1 -Command uninstall -Target C:\path\to\project -Source .
```

卸载只移除 AI Flow 管理的二进制、Skills 和平台入口。项目产生的 `.ai-flow/` 状态、证据、历史以及 `docs/board/` 默认保留。

## 本地开发安装

```bash
git clone https://github.com/mxm-web-develop/coding-skills.git
cd coding-skills
./install/install.sh install --target /path/to/project --source .
```

如果 Release 二进制不存在，安装器会在检测到 Go 时从源码构建 `flowctl`。

## 文档

- [安装、更新和卸载](docs/installation.md)
- [初始化与完整使用流程](docs/workflow.md)
- [人读看板效果示例](docs/examples/human-board-demo.md)
- [flowctl CLI 参考](docs/cli-reference.md)
- [系统架构](docs/architecture/system.md)
- [开发计划](docs/development-plan.md)

## 安全边界

- 默认不自动 push、merge、tag、发布或部署。
- 外部 Evidence 手动录入时保持 `unverified`；只有由 `flowctl evidence run` 实际执行的本地命令才是 `verified-local`。
- Work Item 完成必须至少关联一个可信的 passed Evidence。
- 更新和卸载要求存在安装标记，避免删除用户自建的同名 Skill。
- Release 安装包必须通过 SHA-256 校验。

## 当前版本

`v0.2.5` 提供 14 个 Core Skills 和 15 个 JSON Schema，支持 Cursor、Codex、Claude Code 按任意顺序逐次加入同一工作区并共享唯一 `.ai-flow/` 与 `docs/board/`。安装器只把“入口和完整原生 Skill Pack 同时存在”的 IDE 视为已安装，并在下载、校验、安装和健康检查阶段持续反馈进度。
