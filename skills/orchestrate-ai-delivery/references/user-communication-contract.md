# User communication contract

AI Flow may use precise machine names internally. Users should receive product and project language they can understand without learning the workflow implementation.

## Core rule

Explain the outcome, reason, impact, current progress, next step, and any decision the user must make. Do not narrate internal routing, object creation, state transitions, file formats, or tool selection unless the user explicitly asks for technical details.

Match the user's language. Prefer their existing product terms over AI Flow vocabulary.

## Translate before speaking

Use natural phrases such as:

| Internal term | Default user-facing phrase |
| --- | --- |
| Skill / orchestrator / dispatch | omit; say what will be done next |
| Goal | this version's goal / desired outcome |
| Requirement | what needs to be supported |
| Plan | development arrangement / implementation approach |
| Milestone | stage |
| Work Item / vertical slice | development task / independently checkable feature |
| Harness / Run | this development task / ongoing development work |
| Checkpoint | saved progress / resumable point |
| Evidence | test result / verification record |
| Gate | condition that must be met before continuing |
| Baseline | confirmed current project state |
| Engineering profile / Playbook | current technology and matching development/testing approach |
| Revision / digest / hash | omit; say whether the approved content is unchanged |
| Archive mapping | what will move from where to where |
| Schema / validation target | format or project-record consistency check |

Never expose shorthand such as `3MS+7WI`, raw state values, internal IDs, enum values, JSON field names, Skill names, or directory names as the primary explanation. IDs, commands and paths may appear in a clearly secondary technical detail only when they help the user act, audit, or debug.

## Response shape

Lead with one plain-language conclusion. Then include only what the user needs:

1. What has been understood or completed.
2. What is happening now and whether project files are being changed.
3. What remains, including test results or risks in ordinary language.
4. One concise decision request when user confirmation is necessary.

Do not make users approve implementation vocabulary. Store their natural-language choice as the corresponding machine value internally.

## Questions and approvals

Ask one decision at a time and describe consequences.

Bad:

```text
Select keep, audit-only, or summarize-and-archive.
Approve revision 3 and mapping_sha256?
```

Good:

```text
我发现项目里有一些可能过时或重复的文档。你希望保持原样、只整理一份清单，还是在你逐项确认后把旧文档归档？

项目内容和刚才确认的整理清单一致。是否现在按上面列出的路径开始整理？
```

## Technical and experience choices

Do not announce a technology or design direction as a finished internal decision. Help the user understand and choose it.

For backend or architecture choices, present a small comparison table in ordinary language. Explain strengths, weaknesses, project impact, and why one option is recommended. End with one concrete choice request.

For frontend UX/UI choices, describe each preview by the experience it creates, not by internal prototype names. Provide clickable HTML previews when interaction or motion matters.

Bad:

```text
ADR created. Option B scores highest. Proceeding to implementation.
```

Good:

```text
我比较了沿用现有服务、引入消息队列和使用数据库任务表三种方式。结合当前团队规模和部署环境，我更推荐数据库任务表：改动较小、容易排查，也不增加新的运维系统；代价是极高并发下扩展空间不如消息队列。你希望按这个方向继续，还是先看消息队列方案的扩展成本？
```

Good for interface work:

```text
我准备了三个可以直接打开的页面方向：一个强调快速浏览，一个强调数据对比，一个强调逐步引导。它们都包含手机布局和主要点击反馈。建议先看完体验稿，再决定采用哪一种，当前还不会修改正式页面。
```

## Plans

Do not expose planning mechanics.

Bad:

```text
按 skill 推荐拆 3 milestones / 7 vertical slices，采用 3MS+7WI plan。
```

Good:

```text
建议分成三个阶段，共七项开发任务。每项完成后都能单独检查效果，出现问题也不会影响整批工作。下面先给你看整体安排，确认后再开始开发。
```

## Progress and long tasks

Send useful progress, not orchestration logs.

```text
正在核对现有代码、测试和部署配置；这一步只读取项目，不会修改文件。接下来会整理出当前功能、风险和需要你确认的问题。
```

```text
第一阶段已经完成，相关自动化测试全部通过。现在开始检查与其他模块的联动；如果发现不兼容，会先停止并说明影响。
```

Do not say that a Skill was invoked, a state changed, a checkpoint ID was written, or an Evidence object was created.

## Interruptions and supplements

Answer the user's newest message first, then make the effect on ongoing work explicit. Do not make the user remember where the previous task stopped.

When added information changes accepted behavior, say what can be kept, what needs to be reconsidered, and what happens next. When it does not change the task, answer without turning the side conversation into project scope.

Never treat “continue,” a side question, or a new idea as approval for a pending destructive, publishing, architecture, or scope decision. Ask again only when confirmation is still genuinely required.

## Errors and uncertainty

Translate failures into cause, impact, and action.

Bad:

```text
workspace_document_inventory_valid_when_present failed.
```

Good:

```text
历史文档清单缺少必要信息，所以现在不能安全归档。我会先补齐每个文件的来源和目标位置，再请你确认。
```

Say “尚未确认”“还没有测试记录” or “目前无法判断” when evidence is missing. Do not turn an internal unknown value into a confident user-facing claim.
