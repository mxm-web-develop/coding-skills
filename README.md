# Coding Skills · AI Flow

AI Flow 是一套运行在 Cursor、Codex 和 Claude Code 内部的 AI 开发流程。它把产品目标、需求、技术决策、测试、实现、验证、Git、版本和项目文档连接成一条可恢复、可审计的工作流。

它不需要独立服务或数据库：机器事实保存在项目根目录 `.ai-flow/`，人读看板保存在 `docs/board/`，所有内容可以跟随 Git 由个人或团队共同维护。

## 核心能力

- 15 个同源 Agent Skills，并安装到三个 IDE 各自可发现的项目目录。
- 基于仓库证据生成工程画像，按真实语言、框架、现有命令和已安装社区 Skill 选择开发与测试 Playbook。
- 模块化代码门禁：按职责拆分多文件、核心逻辑优先纯函数、副作用显式隔离、注释解释意图和约束。
- Web UI 同时要求功能 E2E、Playwright 截图回归和人工/AI 视觉设计审查。
- 需求确认后以自然语言共同完成技术选型：后端展示候选方案、优缺点、项目影响和推荐依据；前端可先生成两到三个隔离的 HTML 体验方向，确认布局、风格、动画和交互后再写生产代码。
- 空白项目和既有项目两种初始化方式。
- 既有或文档密集型工作区可选择保持原样、只读盘点，或在逐路径批准后按版本总结归档散落历史文档。
- 初始化只标记疑似废弃代码、目录和生成内容；用户后续明确要求清理时，独立 Skill 才按多语言/多子项目依赖图提出可恢复计划。
- 对话、选择、进度、错误和看板使用自然的产品/项目语言；Skill、对象 ID、状态值、缩写、哈希和机器目录默认只保留在内部记录中。
- 自然语言自动路由：状态、需求、功能、Bug、测试、评审、Git、版本和文档请求都会进入相应流程。
- Work Item、Harness Run、Checkpoint、Evidence 状态机。
- 对话中断可恢复：状态问题、需求补充、独立任务、取消替换、无关问答和 IDE 切换都从共享项目记录继续，不把“继续”或插话误认为批准。
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
4. 将 15 个 Skill 安装到所选 IDE 的原生目录。
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
继续上次中断的开发，先告诉我已经完成了什么和剩余风险。
```

```text
所有测试都通过了吗？请说明实际测试结果和对应的代码版本。
```

```text
准备提交和发布 v1.1.0，但不要自动 push 或 tag。
```

项目助手会读取仓库状态并选择合适流程。状态查询不会启动开发任务；任何代码修改都会关联到具体开发任务和实际测试结果。

准备在真实项目中完整试用时，可按 [v0.3.0 人工全流程验收](docs/manual-acceptance.md) 依次检查空白项目、已有项目、中途补充、跨 IDE 恢复和发布确认。

## 标准交付流程

```text
目标对齐
  → 需求与计划
  → 工程画像与技术栈 Playbook
  → 技术/体验方案比较与用户确认
  → 测试先行
  → Work Item 实现
  → 诊断与真实验证
  → 独立评审
  → Git 集成
  → 版本与发布
  → 机器状态、看板和归档同步
```

小修改可以合并讨论和设计步骤，但不能跳过测试、证据、评审、Git 追踪和知识同步。

## 整体工作机制

AI Flow 想做的事很直接：**让 AI Agent 像一个负责的项目经理一样跟你协作**，而不是一次性吐出大段代码再让你自己收拾。三个 IDE 共用同一份"项目大脑"：

- **判断与协作** 由 15 个 Agent Skill 负责——什么时候问什么问题、何时该停、何时需要你点头。
- **确定性动作** 由本地 CLI `flowctl` 负责——写文件、校验 Schema、生成看板、打 tag 之类不靠 AI 蒙的事。
- **项目记忆** 落在 `.ai-flow/` 机器状态和 `docs/board/` 人读看板，跟着 Git 走，谁来都能恢复。
- **跨 IDE 同步** 通过共享目录和 Skill 元数据实现，Cursor、Codex、Claude Code 看到的是同一份事实，不会出现"在 A 里看到的进度到 B 里就变了"。

整套流程被拆成 **8 个职责清晰的模块**，每个模块由 1–2 个 Skill 负责，按顺序串联。状态查询、Bug 修复、新功能开发等不同入口会跳过部分模块；模块之间是"上一关过不了，下一关不开"的关系。

## 模块划分与每一步做了什么

把上面的标准流程展开，15 个 Skill 归到 8 个模块：

### 模块 1：项目入口与编排

- **负责 Skill**：`initialize-ai-project`、`orchestrate-ai-delivery`
- **做什么**：识别项目是空白还是既有；听懂你说的话是状态查询、新需求、Bug 修复还是别的意图；把任务派发到合适的后续模块。
- **典型动作**：自动路由自然语言请求，决定要不要启动开发任务，避免"你随口说一句它就动手改代码"。

### 模块 2：既有项目接管与基线（仅既有项目）

- **负责 Skill**：`adopt-existing-project`
- **做什么**：只读扫描 Git 历史、依赖、测试、文档和已有版本，把"已经有什么"如实写下来；明确区分代码直接证明的事实、合理推断、需要你确认的内容。
- **典型动作**：生成 `.ai-flow/baseline/`，记录当前版本、已实现功能、文档冲突、待确认事项；**不会自动改业务代码**，也不会把推断当成事实。

### 模块 3：目标对齐与规划

- **负责 Skill**：`discover-product-goal`、`plan-product-delivery`
- **做什么**：多轮讨论 V1/V2 的目标、用户、价值、范围、非目标和验收条件；目标确认后把它拆成里程碑和可独立验证的 Work Item。
- **典型动作**：先问"做给谁、解决什么问题、不做什么、成功长什么样"，再拆成 5–10 个能分别验收的小任务，每个任务都有明确的"做完是什么样"。

### 模块 4：工程画像与技术方案

- **负责 Skill**：`profile-project-engineering`、`research-and-design-solution`
- **做什么**：扫描仓库得到真实语言、框架、测试命令，挑选合适的执行 Playbook；遇到关键技术选择时列出候选方案的优缺点、项目影响和推荐；前端方向不明确时生成 2–3 个隔离的 HTML 体验稿让你预览。
- **典型动作**：写代码之前先告诉你"为什么推荐 A，放弃 B 是因为什么"，并在你确认前不进入实现阶段。

### 模块 5：测试先行

- **负责 Skill**：`specify-tests`
- **做什么**：基于已批准的需求和技术方案，先写测试规格（单元、集成、E2E、视觉），建立需求 → 测试 → Work Item 的可追踪关系。
- **典型动作**：每个 Work Item 都至少有一条失败用例先等着，**没有"测试"就没有"完成"**。

### 模块 6：实现与验证

- **负责 Skill**：`implement-work-item`、`diagnose-and-verify`
- **做什么**：按工程画像和测试规格写最小可工作代码；运行真实命令收集证据（命令、退出码、日志、Git SHA、SHA-256），把通过的测试结果落到 `.ai-flow/evidence/`。
- **典型动作**：每完成一个 Work Item 都用真实命令验证，不靠"我觉得跑过了"。Bug 修复时也走同样路径——先写复现用例，再修，再验证。

### 模块 7：评审与 Git 集成

- **负责 Skill**：`review-change`、`integrate-git-change`
- **做什么**：独立视角审查需求覆盖、模块拆分、风险和测试充分性；做原子提交、写清晰的提交信息、关联 Work Item 和 Evidence；**默认不自动 push、不自动 merge、不自动 tag**。
- **典型动作**：发现明显漏洞会暂停让你确认；提交时把 Work Item 和测试结果串成一条可追踪的链。

### 模块 8：发布与知识同步

- **负责 Skill**：`manage-release`、`sync-project-knowledge`
- **做什么**：核对发布门禁（Schema 合法、追踪链完整、测试通过、看板新鲜）；计算下一个版本号、生成变更摘要；刷新四份人读看板，标记被替代的旧方案并归档。
- **典型动作**：发布前会逐项核对，没通过的项目直接卡住，让你看清还差什么；不会"差不多就发"。

### 横向模块：工作区治理（按需触发）

- **负责 Skill**：`clean-project-workspace`
- **什么时候触发**：只有项目初始化完成后，你**明确**说"清理工作区/归档废弃代码"时才会跑。
- **做什么**：重新扫描当前 Git revision，按多语言组件依赖图把确认废弃的代码、目录、生成物归档到 `.ai-flow/archive/`，保留可恢复映射。
- **典型动作**：默认不动；只有你点头才会动文件，过程中任何失败都会立即停下并恢复。

## 自动适配你本地已经装好的 Skill

AI Flow 的 15 个 Core Skill 只负责**判断流程**——什么时候问什么、什么时候该停、什么时候需要你确认。具体怎么写代码，会优先用你本地 IDE 里**已经装好的、跟当前工程画像匹配的 Skill**（比如 UI 调优用 `impeccable`、React 性能用 Vercel 的 React Skill、Python 测试用 pytest Skill）。

### 怎么工作

1. **发现**：你把一个 Skill 装到 `.cursor/skills/`、`.agents/skills/` 或 `.claude/skills/` 任意一个目录，AI Flow 在下一次画像时就能看到它。
2. **画像匹配**：AI Flow 扫描仓库，识别出真实语言、框架、测试命令、是否涉及 UI/视觉等，然后从你已装的 Skill 里挑出最匹配当前任务的几个。
3. **登记**：被选中的 Skill 会写进 `.ai-flow/baseline/engineering-profile.json`，记录名称、来源、版本、信任等级和"为什么选它"。
4. **按需调用**：下游 Skill（实现、评审、诊断等）会读这份画像，按当前任务需要调用对应的社区 Skill。装上之后不用你手动触发，告诉 Agent "重新拉一下工程画像"或直接开始下一个任务就行。

### 信任分级

AI Flow 不会在执行过程中偷偷下载、安装或升级任何你本地没装的第三方 Skill。被选中的 Skill 按来源打分，登记到画像里：

1. 语言、框架或工具的**官方**维护
2. 官方生态或知名工程组织
3. 维护良好的**成熟社区**项目
4. 个人发布者——**需明确审核过**才纳入

### 优先级规则

**项目自身的规则、已确认的技术决策、用户授权永远优先**——社区 Skill 只是补强，不会越权。例如你已经决定项目用 Vue 3，AI Flow 不会因为某个 React Skill 评分高就改用 React；项目里已有的 ESLint 配置也不会被新 Skill 的偏好覆盖。

### 常用 Skill 源（AI Flow 已审核过）

| 来源 | 适合场景 | 信任 |
|---|---|---|
| Vercel Agent Skills | React/Next.js 性能、组合模式、React Native | 官方 |
| Anthropic Skills | 前端设计、Web 应用测试 | 官方 |
| GitHub Awesome Copilot | Jest、pytest、Playwright、Vue/Pinia 等 | 官方生态 |
| Superpowers | TDD、系统化调试、验证前完成 | 成熟社区 |

这只是发现用的"短名单"，不是自动依赖。每个被选中的 Skill 都要在使用前看完它的完整说明，并登记实际选中的来源和版本。

## 典型使用案例

下面五个案例展示 AI Flow 在真实项目里怎么跑。每个案例都按"**你说** → **AI Flow 的实际走法** → **结束时**"来写，覆盖最常见的几种情况。

### 案例 1：从零启动一个空白项目

> 你有一个空文件夹，想做一个买家、卖家都能用的电商网站。

- **你说**：`初始化这个空白项目并接入 AI Flow。`
- **AI Flow 的实际走法**：先和你多轮聊清楚"做给谁、解决什么问题、不做什么、成功长什么样"——不会一上来就问"用 Vue 还是 React"。技术选型时列出候选方案和取舍，等你确认。目标定了之后拆成 5–8 个能分别验收的小任务，每完成一项就跑真实测试、把证据给你看，从不直接说"做完了"。
- **结束时**：看板上能看到 V1 的目标、所有 Work Item 的状态、测试通过率和下一步建议。

### 案例 2：接管一个已经有代码的老项目

> 仓库跑了一年，有代码、有 Git、有测试、有各种历史方案文档；想从下个版本开始用 AI Flow 管起来。

- **你说**：`把这个项目接入 AI Flow，先告诉我现在是什么版本、已经实现了什么、文档里有没有冲突，不要改代码。`
- **AI Flow 的实际走法**：只读扫描仓库，输出一份"项目基线"——当前版本、已有功能、测试情况、文档冲突、待确认事项。**整个过程完全不动业务代码**。事实分三类标注：代码直接证明的、合理推断的、需要你确认的，不会混在一起。如果有散落历史文档，问你"保持原样 / 只读盘点 / 批准后归档"，选哪个都行，默认不动文件。你对基线满意后，才开始聊新版本的需求和计划。
- **结束时**：项目状态写进 `.ai-flow/baseline/`，后续所有改动都以这份基线为对照，不会偷偷改业务代码。

### 案例 3：修一个线上偶发的 Bug

> 结算页偶尔会重复提交订单，要复现并修复。

- **你说**：`结算页重复提交订单的 Bug，复现并修复。`
- **AI Flow 的实际走法**：先写一个能稳定复现这个 Bug 的失败用例（**测试先行**，没有可复现的失败用例就不会动手改代码）。定位根因后改代码，跑全部相关测试，把命令、退出码、日志、Git SHA 当作证据落库。独立评审通过后做原子提交，关联到这个 Bug 的 Work Item 上。**不会出现"AI 说修好了但其实没测过"**——证据不通过，任务就不算完成。
- **结束时**：Bug 修复可追溯到具体的 commit、测试用例和证据，不只是"AI 说修好了"。

### 案例 4：开发到一半被打断，跨 IDE 恢复

> 用 Codex 写了三个 Work Item，要切到 Cursor 继续；或者午饭前停在某个任务上，下午回来继续。

- **你说**：`继续上次中断的开发。先告诉我做到了哪、下一步是什么。`
- **AI Flow 的实际走法**：先告诉你当前大版本、哪些任务已完成、哪些正在做、有没有阻塞项、上次保存的 Git SHA 是哪个。然后核对当前代码是否和暂停时一致——不一致会先解释差异再决定怎么处理，**不会蒙头接着干**。注意："继续"两个字**不会**自动批准之前所有待你确认的事项，技术方案、文件删除、目录移动这些仍需要你单独点头。切到另一个 IDE 后看到的是同一份项目状态和同一份看板，没有第二份进度文档。
- **结束时**：无缝接续，不会因为换 IDE 或重启就丢上下文。

### 案例 5：完成一个版本并发布

> 所有任务都跑通了，准备发版本。

- **你说**：`所有测试都通过了吗？准备发布 v1.1.0，但不要自动 push 或 tag。`
- **AI Flow 的实际走法**：先用真实命令汇总测试结果——哪个代码版本、跑了哪条命令、退出码是什么、日志哈希是多少，**不是它自己说"通过了"就算**。再逐项核对发布门禁：Schema 是否合法、追踪链是否完整、每个 Work Item 是否都有可信的测试证据、看板是否新鲜、有没有阻塞项。有一项不过就停下来告诉你还差什么；都过了会列出"接下来要做的事"（创建 tag、写 changelog 等），每一步等你点头。**默认不会自动 push / merge / tag / 发布 / 部署**——这些动作都明确分别需要授权。
- **结束时**：版本记录落进 `.ai-flow/releases/`，看板上出现新版本行，旧的 Work Item 自动归档到对应小版本下。

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
  | AI_FLOW_VERSION=v0.3.0 sh
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

`v0.4.1` 让"手动记录的验证证据"可以脱离开发执行独立存在：`evidence record` 新增 `--mode` 参数；当 `--source=agent-claim` 或 `--mode=external` 时不再要求传 `--run`，反之 `--source=external` 在默认 `mode=run` 下仍必须传 `--run`（向后兼容）。`--source=local` 仅在 `evidence run` 里合法，`evidence record` 不再接受。Schema 上 `run_id` 变为可空，独立 evidence 在校验时跳过"开发执行 ↔ 验证证据"反向链路；只有 `source=local` 没有 run 才报错。完整规则见 `docs/cli-reference.md` 中 `evidence record` 章节。

`v0.4.0` 给每个未发布的版本都配了一份"人话版"的开发方案，方便 PM、团队成员和相关方在没有 AI Flow 背景的情况下也能看明白这一版要做什么、怎么拆、用了什么技术、有什么风险。`render-board` 现在除了原有的四份看板外，还会同步生成 `docs/board/PLANS.md`（所有未发布版本的方案索引）和 `docs/board/plans/v<版本>.md`（每份方案文档）；方案用面向/要解决的问题/完成后能提供/范围内/不在范围内/验收要点/阶段划分/开发任务清单/技术选型/风险与依赖这类自然语言小节组织，不再以原始 ID 或状态机值作为主语。同步在 `user-communication-contract.md` 增加"禁止漏词表"小节，明确禁止在面向用户的话术里出现 `§2`/`WI`/`DEC`/`form_decisions`/`in_progress`/commit SHA 这类内部词作为主语。

`v0.3.1` 是 `v0.3.0` 的小幅修订：当实现、计划、评审或诊断过程中发现任务范围/需求范围/测试范围与原计划不一致时，禁止在面向用户的多选确认里暴露 "WI / WI scope / scope JSON / ADR / Evidence" 之类的内部标签或内部 ID。具体翻译规则和正反范例见 `skills/orchestrate-ai-delivery/references/user-communication-contract.md` 末尾的 "When you have to ask about scope, plan, or backlog changes" 章节。

`v0.3.0` 提供 15 个 Core Skills 和 17 个 JSON Schema。新版增加技术方案对比与前端 HTML 体验稿确认，固化对话中断、补充、恢复和跨 IDE 交接，并对需求、方案、任务、测试、证据与发布关系做完整校验。人读看板由工具统一生成并检查新鲜度，Cursor、Codex、Claude Code 继续共享唯一 `.ai-flow/` 与 `docs/board/`。
