# 安装、更新和卸载

## 1. 支持环境

| 系统 | 架构 | 安装入口 |
| --- | --- | --- |
| macOS | amd64、arm64 | `bootstrap.sh` |
| Linux / WSL | amd64、arm64 | `bootstrap.sh` |
| Windows PowerShell | amd64、arm64 | `bootstrap.ps1` |

Release 包内包含静态 Go 二进制，目标项目无需安装 Go、Node.js 或 Python。

仓库中的 `cmd/flowctl/main.go` 是开发源码，不是用户安装入口。安装器会按操作系统和 CPU 架构下载/复制预编译的 `.ai-flow/bin/flowctl[.exe]`。只有维护者设置 `AI_FLOW_BUILD_SOURCE=1` 进行源码级安装测试时才要求本机存在 Go。

## 2. 一行式安装

先进入目标项目根目录。

### macOS / Linux / WSL

```bash
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh | sh -s -- --cursor
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh | sh -s -- --codex
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh | sh -s -- --claude
```

多个 IDE 共用一个工作区时可以组合：

```bash
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh \
  | sh -s -- --cursor --codex
```

不传 `--cursor`、`--codex`、`--claude` 时默认安装全部，保持向后兼容。`--all` 也可以显式选择全部平台。

### Windows PowerShell

```powershell
$env:AI_FLOW_PLATFORMS="cursor"
irm https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.ps1 | iex
Remove-Item Env:AI_FLOW_PLATFORMS
```

PowerShell 可使用 `cursor`、`codex`、`claude` 或逗号组合，例如 `cursor,codex`。下载脚本后直接执行时，也支持 `-Cursor`、`-Codex`、`-Claude` 和 `-All` switch。

远程 bootstrap 过程：

1. 从 GitHub Releases 下载 `coding-skills.tar.gz` 或 `coding-skills.zip`。
2. 下载 `checksums.txt`。
3. 校验压缩包 SHA-256。
4. 解压到临时目录。
5. 调用包内平台安装器。
6. 安装完成后删除临时目录。

## 3. 指定目标目录和版本

### Shell 环境变量

使用环境变量：

```bash
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh \
  | AI_FLOW_TARGET=/path/to/project AI_FLOW_VERSION=v0.2.3 sh -s -- --cursor
```

也可以传参：

```bash
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh \
  | sh -s -- --version v0.2.3 --target /path/to/project --cursor
```

### PowerShell 环境变量

```powershell
$env:AI_FLOW_TARGET="C:\work\my-project"
$env:AI_FLOW_VERSION="v0.2.3"
$env:AI_FLOW_PLATFORMS="cursor"
irm https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.ps1 | iex
Remove-Item Env:AI_FLOW_TARGET
Remove-Item Env:AI_FLOW_VERSION
Remove-Item Env:AI_FLOW_PLATFORMS
```

## 4. 从 Git clone 安装

```bash
git clone https://github.com/mxm-web-develop/coding-skills.git
cd coding-skills
./install/install.sh install --cursor --target /path/to/project --source .
```

PowerShell：

```powershell
git clone https://github.com/mxm-web-develop/coding-skills.git
cd coding-skills
.\install\install.ps1 -Command install -Cursor -Target C:\work\my-project -Source .
```

源码安装优先使用 `dist/` 中与当前系统匹配的二进制。没有匹配二进制时，如果本机安装了 Go，则现场构建。

## 5. 安装内容

以下路径按选择的平台生成；不会为未选择的 IDE 创建 Skill 或入口文件。

```text
.agents/skills/<14 core skills>/       # Codex
.cursor/skills/<14 core skills>/       # Cursor
.claude/skills/<14 core skills>/       # Claude Code
.claude/skills/ai-flow/SKILL.md         # Claude /ai-flow 入口
.cursor/rules/ai-flow.mdc
.ai-flow/bin/flowctl[.exe]
.ai-flow/runtime/schemas/*.schema.json
.ai-flow/install/version
.ai-flow/install/profile
.ai-flow/capabilities.yaml
AGENTS.md       # 添加带标记的 AI Flow 区块
CLAUDE.md       # 添加带标记的 AI Flow 区块
```

安装器只维护 `<!-- ai-flow:start -->` 与 `<!-- ai-flow:end -->` 之间的内容，不覆盖文件中其他说明。各平台 Skill 副本都由仓库中的同一份 `skills/` 源生成，并带有管理标记；更新时若发现同名但不受管理的 Skill，会停止而不是覆盖。

### IDE Skill 发现矩阵

| IDE | 官方项目级发现目录 | AI Flow 安装目录 |
| --- | --- | --- |
| Cursor | `.cursor/skills/`、`.agents/skills/` | `--cursor` 安装 `.cursor/skills/`；同时选择 Codex 时也存在 `.agents/skills/` |
| Codex | `.agents/skills/` | `.agents/skills/` |
| Claude Code | `.claude/skills/` | `.claude/skills/` |

注意目录名是 `.agents/skills`，`agents` 为复数；`.agent` 或 `.agent/skills` 不属于这些 IDE 的约定目录。官方依据：[Cursor Agent Skills](https://cursor.com/docs/skills)、[Codex Skills](https://learn.chatgpt.com/docs/build-skills)、[Claude Code Skills](https://code.claude.com/docs/en/skills)。

## 6. 安装与初始化的区别

安装完成时可能看到：

```text
project-state WARNING run initialize-ai-project when not initialized
```

这是正常状态。安装器不应该判断一个代码仓库当前开发到了哪里。请随后在 IDE 中提出“初始化这个项目”，或手动执行 `flowctl project init`。

## 7. 更新

### 远程更新

```bash
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh \
  | AI_FLOW_COMMAND=update sh
```

不传平台时更新当前已安装的平台；要增量增加 Codex：

```bash
curl -fsSL https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.sh \
  | AI_FLOW_COMMAND=update sh -s -- --codex
```

```powershell
$env:AI_FLOW_COMMAND="update"
$env:AI_FLOW_PLATFORMS="cursor"
irm https://raw.githubusercontent.com/mxm-web-develop/coding-skills/main/install/bootstrap.ps1 | iex
Remove-Item Env:AI_FLOW_COMMAND
Remove-Item Env:AI_FLOW_PLATFORMS
```

### 本地更新

```bash
git pull --ff-only
./install/install.sh update --target /path/to/project --source .
```

带平台参数的更新只更新所选平台，并保留此前安装的其他平台；例如在 Cursor 安装上运行 `update --codex` 会增加 Codex 支持。更新不会修改项目对象和人读看板。

## 8. 卸载

```bash
./install/install.sh uninstall --target /path/to/project --source .
```

```powershell
.\install\install.ps1 -Command uninstall -Target C:\work\my-project -Source .
```

卸载始终移除整套 AI Flow 管理文件，因此不接受平台参数；项目状态和看板仍按下述规则保留。

卸载保留：

- `.ai-flow/manifest.yaml`
- Goals、Requirements、Work Items、Runs、Checkpoints、Evidence、Releases 和 archive
- `docs/board/`
- 用户原有 `AGENTS.md`、`CLAUDE.md` 内容

如需删除项目数据，必须由用户另外明确处理。

## 9. 健康检查

```bash
.ai-flow/bin/flowctl doctor --root .
```

检查内容：

- 当前平台二进制。
- 所选平台目录中的 14 个 Core Skills。
- 所选 Cursor、Codex 或 Claude Code 入口。
- JSON Schema 安装数量。
- 项目是否已经初始化。

## 10. 常见问题

### 安装器拒绝覆盖同名 Skill

目标项目存在未被 AI Flow 管理的同名 Skill。安装器会停止，避免破坏用户文件。先检查冲突内容，再决定重命名或人工合并。

### `update` 或 `uninstall` 提示没有安装标记

`update` 和 `uninstall` 要求 `.ai-flow/install/version`。如果这个标记被手动删除，请重新执行 `install`，不要手工伪造标记。安装器会通过 Skill 的 `.ai-flow-managed` 标记或旧版 Cursor Rule 的多项 AI Flow 内容签名恢复受管安装；无法确认归属的同名文件仍会停止并要求人工检查。Cursor Rule 的独立版本标记只用于审计，不能单独授权覆盖内容不匹配的文件。

### 重新安装提示 `existing unmanaged Cursor ai-flow rule`

从 `v0.2.3` 起，安装器可以识别旧版本生成的 `.cursor/rules/ai-flow.mdc`，即使中心安装标记已被删除，也可以直接重新执行安装。若仍出现该错误，说明该文件不符合 AI Flow 生成规则的签名，安装器会保护它不被覆盖；请先查看内容并决定是否重命名或人工合并。

### 找不到 Release

确认仓库至少发布了一个 GitHub Release，并检查 `AI_FLOW_VERSION` 是否为实际 tag。

### checksum 失败

不要绕过。重新下载；若持续失败，检查代理、缓存和 GitHub Release 资产是否匹配。

### IDE 没有发现 Skill

1. 运行 `flowctl doctor`。
2. 确认已选择平台对应的 `codex-skills`、`cursor-skills` 或 `claude-skills` 是 `OK`；未安装的平台不会出现在检查列表中。
3. 确认当前 IDE 打开的是执行安装命令时的目标项目根目录，而不是父目录、子目录或另一个 worktree。
4. 安装或更新后必须 Reload IDE Window，并新建一个 Agent chat；Cursor 在启动时发现 Skills，旧会话不会可靠刷新。
5. Cursor 可输入 `/` 并选择 `initialize-ai-project`；Codex 可从 Skill 选择器选择它；Claude Code 可输入 `/initialize-ai-project` 或 `/ai-flow`。
6. 如果显式调用能工作而自然语言不触发，检查 `.cursor/rules/ai-flow.mdc`、`AGENTS.md` 或 `CLAUDE.md` 是否被其他更高优先级规则覆盖。
