# AI Flow 开发计划

状态：v0.1.1 跨 IDE 发现修复基线

更新日期：2026-08-17

计划原则：先确保三套 IDE 共用一个可安装、可恢复、可验证的小闭环，再增加团队并发、远程平台和生产交付复杂度。

## 1. 已完成的 v0.1.x

### 安装与平台入口

- Shell 与 PowerShell 本地安装、更新、卸载。
- GitHub Release bootstrap、一行式远程安装与 SHA-256 校验。
- Cursor、Codex、Claude Code 路由入口；业务规则使用一个源码并安装到各自的原生发现目录。
- 13 个 Core Skill，均通过 Skill 结构校验。

### 项目与流程

- greenfield 与 existing 两种项目初始化模式。
- Orchestrator 自然语言路由、状态查询短路径和完整交付路径。
- Work Item 状态机、revision 乐观并发和单写 lease。
- Harness Run、预算、Checkpoint 保存与 Git SHA 恢复检查。
- Evidence 真实命令执行、日志、退出码、Git SHA、SHA-256 和信任等级。
- Work Item 完成前重新验证可信 passed Evidence 的日志完整性。

### 状态和文档

- 13 个 Draft 2020-12 JSON Schema。
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

## 3. v0.2.0：团队协作与迁移

- Work Item scope 重叠检测、lease 续期/接管和 worktree 辅助命令。
- Agent/用户身份、Run、commit、PR 的审计关系。
- GitHub/GitLab PR、Checks、CODEOWNERS 和 ruleset 可选适配。
- Schema 与 Skill 包 N-1 → N 数据迁移及失败回滚。
- Goal、Requirement、Decision、Release 和 archive 的确定性 CLI CRUD。
- 看板增量渲染、发布索引和陈旧性检查。

验收：多个 Agent 不会静默覆盖同一 revision 或重叠 scope；升级失败不丢项目数据；受保护路径缺少批准时不能进入发布门禁。

## 4. v0.3.0：安全与生产交付 Profile

- `secure`：依赖锁、SBOM、密钥扫描、漏洞证据、威胁模型和安全审批。
- `delivery`：环境声明、部署证据、渐进发布、回滚和部署审批。
- `observe`：日志、指标、trace、健康验证、事故记录和恢复报告。
- MCP 只做能力适配；缺少供应商 MCP 时仍提供 CLI 或人工证据降级路径。

## 5. 持续测试策略

- Unit：ID、状态转换、revision、锁、checksum、路径和看板汇总。
- Contract：Skill 的输入、输出、停止条件和允许状态变化。
- Schema：合法对象、非法状态、悬空引用和 Event JSONL。
- E2E：空白/既有项目、成功、失败、篡改、中断恢复、更新和卸载。
- Platform：三个 IDE 的发现、自然语言路由和无 MCP/Hook/子 Agent 降级。
- Release：六个二进制可执行格式、包内容、checksum 和远程 bootstrap。
- Security：路径穿越、命令参数边界、供应链校验、敏感信息和权限边界。

## 6. Definition of Done

一个版本只有同时满足以下条件才算完成：

- 功能代码、测试、Schema、Skill 契约和用户文档同步更新。
- 至少一个真实端到端场景，不只依赖 Agent 文字输出。
- 失败路径、证据完整性和恢复路径经过验证。
- 三平台公共语义一致，平台差异只存在于 adapter。
- 没有新增散落状态或重复事实源。
- 发布物可校验、可更新、可卸载；外部副作用仍受用户和仓库权限控制。
