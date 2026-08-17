#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
export AI_FLOW_BUILD_SOURCE=1
CURSOR_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-cursor-e2e.XXXXXX")
CLAUDE_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-claude-e2e.XXXXXX")
REVERSE_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-reverse-e2e.XXXXXX")
ALL_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-all-e2e.XXXXXX")
CURSOR_CONFLICT_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-cursor-add-conflict-e2e.XXXXXX")
CLAUDE_CONFLICT_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-claude-add-conflict-e2e.XXXXXX")

cleanup() {
  rm -rf "$CURSOR_ROOT" "$CLAUDE_ROOT" "$REVERSE_ROOT" "$ALL_ROOT" "$CURSOR_CONFLICT_ROOT" "$CLAUDE_CONFLICT_ROOT"
}
trap cleanup EXIT HUP INT TERM

assert_all_platforms() {
  project_root="$1"
  for platform in cursor codex claude; do
    [ "$(grep -c "^$platform$" "$project_root/.ai-flow/install/platforms")" -eq 1 ]
  done
  [ "$(wc -l < "$project_root/.ai-flow/install/platforms" | tr -d ' ')" -eq 3 ]
  [ -f "$project_root/.cursor/skills/initialize-ai-project/SKILL.md" ]
  [ -f "$project_root/.agents/skills/initialize-ai-project/SKILL.md" ]
  [ -f "$project_root/.claude/skills/initialize-ai-project/SKILL.md" ]
  [ -f "$project_root/.cursor/rules/ai-flow.mdc" ]
  [ -f "$project_root/AGENTS.md" ]
  [ -f "$project_root/CLAUDE.md" ]
}

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

printf '%s\n' "0.2.3" > "$CURSOR_ROOT/.cursor/skills/initialize-ai-project/.ai-flow-managed"
"$REPO_ROOT/install/install.sh" install --codex --target "$CURSOR_ROOT" --source "$REPO_ROOT" >/dev/null
[ "$(sed -n '1p' "$CURSOR_ROOT/.cursor/skills/initialize-ai-project/.ai-flow-managed")" = "0.2.5" ]
[ -f "$CURSOR_ROOT/.cursor/skills/initialize-ai-project/SKILL.md" ]
[ -f "$CURSOR_ROOT/.agents/skills/initialize-ai-project/SKILL.md" ]
[ -f "$CURSOR_ROOT/AGENTS.md" ]
[ ! -e "$CURSOR_ROOT/.claude/skills/initialize-ai-project" ]
grep -q '^cursor$' "$CURSOR_ROOT/.ai-flow/install/platforms"
grep -q '^codex$' "$CURSOR_ROOT/.ai-flow/install/platforms"
"$REPO_ROOT/install/install.sh" install --claude --target "$CURSOR_ROOT" --source "$REPO_ROOT" >/dev/null
assert_all_platforms "$CURSOR_ROOT"

AI_FLOW_PLATFORMS=claude "$REPO_ROOT/install/install.sh" install --target "$CLAUDE_ROOT" --source "$REPO_ROOT" >/dev/null
[ -f "$CLAUDE_ROOT/.claude/skills/initialize-ai-project/SKILL.md" ]
[ -f "$CLAUDE_ROOT/.claude/skills/profile-project-engineering/SKILL.md" ]
[ -f "$CLAUDE_ROOT/.claude/skills/ai-flow/SKILL.md" ]
[ -f "$CLAUDE_ROOT/CLAUDE.md" ]
[ ! -e "$CLAUDE_ROOT/.cursor/skills/initialize-ai-project" ]
[ ! -e "$CLAUDE_ROOT/.agents/skills/initialize-ai-project" ]

# Reverse order must converge on the same single runtime and platform set.
"$REPO_ROOT/install/install.sh" install --claude --target "$REVERSE_ROOT" --source "$REPO_ROOT" >/dev/null
"$REPO_ROOT/install/install.sh" install --codex --target "$REVERSE_ROOT" --source "$REPO_ROOT" >/dev/null
"$REPO_ROOT/install/install.sh" install --cursor --target "$REVERSE_ROOT" --source "$REPO_ROOT" >/dev/null
assert_all_platforms "$REVERSE_ROOT"
[ "$(find "$REVERSE_ROOT" -path '*/.ai-flow/bin/flowctl' -type f | wc -l | tr -d ' ')" -eq 1 ]
"$REVERSE_ROOT/.ai-flow/bin/flowctl" project init --root "$REVERSE_ROOT" --mode greenfield --name "Reverse Platform Project" >/dev/null
[ -f "$REVERSE_ROOT/docs/board/STATUS.md" ]
[ "$(find "$REVERSE_ROOT" -path '*/docs/board/STATUS.md' -type f | wc -l | tr -d ' ')" -eq 1 ]

# One-shot installation is equivalent to incremental installation.
"$REPO_ROOT/install/install.sh" install --all --target "$ALL_ROOT" --source "$REPO_ROOT" >/dev/null
assert_all_platforms "$ALL_ROOT"

# Adding a platform must not weaken unmanaged-file protection.
"$REPO_ROOT/install/install.sh" install --codex --target "$CURSOR_CONFLICT_ROOT" --source "$REPO_ROOT" >/dev/null
mkdir -p "$CURSOR_CONFLICT_ROOT/.cursor/rules"
printf '%s\n' "user-owned Cursor rule" > "$CURSOR_CONFLICT_ROOT/.cursor/rules/ai-flow.mdc"
if "$REPO_ROOT/install/install.sh" install --cursor --target "$CURSOR_CONFLICT_ROOT" --source "$REPO_ROOT" >/dev/null 2>&1; then
  printf '%s\n' "adding Cursor unexpectedly overwrote a user-owned rule" >&2
  exit 1
fi
grep -q 'user-owned Cursor rule' "$CURSOR_CONFLICT_ROOT/.cursor/rules/ai-flow.mdc"

"$REPO_ROOT/install/install.sh" install --cursor --target "$CLAUDE_CONFLICT_ROOT" --source "$REPO_ROOT" >/dev/null
mkdir -p "$CLAUDE_CONFLICT_ROOT/.claude/skills/ai-flow"
printf '%s\n' "user-owned Claude entry" > "$CLAUDE_CONFLICT_ROOT/.claude/skills/ai-flow/SKILL.md"
if "$REPO_ROOT/install/install.sh" install --claude --target "$CLAUDE_CONFLICT_ROOT" --source "$REPO_ROOT" >/dev/null 2>&1; then
  printf '%s\n' "adding Claude unexpectedly overwrote a user-owned entry" >&2
  exit 1
fi
grep -q 'user-owned Claude entry' "$CLAUDE_CONFLICT_ROOT/.claude/skills/ai-flow/SKILL.md"

printf '%s\n' "AI Flow platform selection E2E passed"
