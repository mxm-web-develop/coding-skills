# Conversation continuity

Treat every user message as a possible continuation, interruption, supplement, replacement, question, approval response, or unrelated side request. Re-read repository state before deciding which one it is; do not depend on chat memory alone.

## Classify the new message

- **Status or explanation question:** answer from current project records without ending, duplicating, or advancing active work. Then state whether the previous task remains active and its next understandable step.
- **Clarification within accepted scope:** record the clarification and continue from the current safe point when it does not invalidate completed work.
- **Material addition or changed acceptance:** save current progress, explain which completed or planned parts are affected, update the requirement/plan/decision/test records, and re-run only the downstream work made stale.
- **New independent request:** preserve the current task and ask which should run first only when both would write overlapping areas. Otherwise create a separate traceable task and keep ownership boundaries clear.
- **Replacement or cancellation:** confirm the intended replacement when ambiguity could discard useful work. Preserve the history, cancel or supersede affected active records, and explain what can be reused.
- **Unrelated side question:** answer it without treating it as scope, approval, cancellation, or completion. Briefly remind the user where the development task is paused.
- **Resume request or IDE switch:** read the shared current state, active task, latest saved progress, Git status, lease, and code revision. Resume the same run through an explicit ownership handoff; do not open a duplicate run because the editor or Agent identity changed. Resume only if records and code agree; otherwise explain the drift and the safe recovery action.

## Pending decisions

A reply counts as approval only when it clearly answers the exact pending choice. A new idea, status question, silence, “continue,” or approval of a different topic does not approve a technology choice, file move, push, merge, release, deployment, deletion, or scope expansion.

When a message both answers the pending choice and adds new information, record both. If the new information changes the consequences of that choice, show the updated impact and confirm again rather than using the stale approval.

## Preserve the causal chain

Before pausing or changing direction:

1. Save what is complete, what changed, the next safe action, open questions, and the current code revision.
2. Update the originating need when the user's intent changed; increment its revision or create a superseding record instead of silently rewriting history.
3. Mark dependent plans, technical choices, tests, or implementation results stale when their assumptions no longer hold.
4. Keep unaffected completed work and verification results; do not restart the whole workflow by default.
5. Regenerate the user board after machine state changes.

## User-facing response

Use three short parts when an interruption affects active work:

1. Answer or acknowledge the newest message.
2. Explain its impact on the ongoing task in product language.
3. State the next action and ask at most one decision when required.

Do not say that routing changed, a checkpoint was written, an object revision increased, or a state machine moved. Say, for example:

```text
这个补充会改变退款完成的判断方式。已经完成的页面布局可以保留，但接口设计和对应测试需要重新确认。我先停在当前安全位置，更新这两部分后再继续开发；其他已通过的检查不需要重做。
```

For a side question:

```text
当前版本仍是 v1.2.0。你刚才的开发任务没有取消，已经完成的部分也已保留；回答完这个问题后，下一步仍是确认批量导入失败时的处理方式。
```
