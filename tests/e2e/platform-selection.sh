#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
export AI_FLOW_BUILD_SOURCE=1
CURSOR_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-cursor-e2e.XXXXXX")
CLAUDE_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-claude-e2e.XXXXXX")

cleanup() {
  rm -rf "$CURSOR_ROOT" "$CLAUDE_ROOT"
}
trap cleanup EXIT HUP INT TERM

"$REPO_ROOT/install/install.sh" install --cursor --target "$CURSOR_ROOT" --source "$REPO_ROOT" >/dev/null
CURSOR_FLOWCTL="$CURSOR_ROOT/.ai-flow/bin/flowctl"
[ -f "$CURSOR_ROOT/.cursor/skills/initialize-ai-project/SKILL.md" ]
[ -f "$CURSOR_ROOT/.cursor/skills/profile-project-engineering/SKILL.md" ]
[ -f "$CURSOR_ROOT/.cursor/rules/ai-flow.mdc" ]
[ ! -e "$CURSOR_ROOT/.agents/skills/initialize-ai-project" ]
[ ! -e "$CURSOR_ROOT/.claude/skills/initialize-ai-project" ]
[ ! -e "$CURSOR_ROOT/AGENTS.md" ]
[ ! -e "$CURSOR_ROOT/CLAUDE.md" ]
[ "$(sed -n '1p' "$CURSOR_ROOT/.ai-flow/install/platforms")" = "cursor" ]
CURSOR_DOCTOR=$("$CURSOR_FLOWCTL" doctor --root "$CURSOR_ROOT" --json)
printf '%s\n' "$CURSOR_DOCTOR" | grep -q '"name": "cursor-skills"'
if printf '%s\n' "$CURSOR_DOCTOR" | grep -q '"name": "codex-skills"'; then
  printf '%s\n' "Cursor-only doctor unexpectedly checked Codex" >&2
  exit 1
fi

"$REPO_ROOT/install/install.sh" update --target "$CURSOR_ROOT" --source "$REPO_ROOT" >/dev/null
[ ! -e "$CURSOR_ROOT/.agents/skills/initialize-ai-project" ]
[ ! -e "$CURSOR_ROOT/.claude/skills/initialize-ai-project" ]

"$REPO_ROOT/install/install.sh" update --codex --target "$CURSOR_ROOT" --source "$REPO_ROOT" >/dev/null
[ -f "$CURSOR_ROOT/.cursor/skills/initialize-ai-project/SKILL.md" ]
[ -f "$CURSOR_ROOT/.agents/skills/initialize-ai-project/SKILL.md" ]
[ -f "$CURSOR_ROOT/AGENTS.md" ]
[ ! -e "$CURSOR_ROOT/.claude/skills/initialize-ai-project" ]
grep -q '^cursor$' "$CURSOR_ROOT/.ai-flow/install/platforms"
grep -q '^codex$' "$CURSOR_ROOT/.ai-flow/install/platforms"

AI_FLOW_PLATFORMS=claude "$REPO_ROOT/install/install.sh" install --target "$CLAUDE_ROOT" --source "$REPO_ROOT" >/dev/null
[ -f "$CLAUDE_ROOT/.claude/skills/initialize-ai-project/SKILL.md" ]
[ -f "$CLAUDE_ROOT/.claude/skills/profile-project-engineering/SKILL.md" ]
[ -f "$CLAUDE_ROOT/.claude/skills/ai-flow/SKILL.md" ]
[ -f "$CLAUDE_ROOT/CLAUDE.md" ]
[ ! -e "$CLAUDE_ROOT/.cursor/skills/initialize-ai-project" ]
[ ! -e "$CLAUDE_ROOT/.agents/skills/initialize-ai-project" ]

printf '%s\n' "AI Flow platform selection E2E passed"
