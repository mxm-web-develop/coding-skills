#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
export AI_FLOW_BUILD_SOURCE=1
TEST_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-install-e2e.XXXXXX")
CONFLICT_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-conflict-e2e.XXXXXX")
NATIVE_CONFLICT_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-native-conflict-e2e.XXXXXX")
RECOVERY_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-recovery-e2e.XXXXXX")
RULE_CONFLICT_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-rule-conflict-e2e.XXXXXX")
CODEX_RECOVERY_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-codex-recovery-e2e.XXXXXX")
CLAUDE_RECOVERY_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-claude-recovery-e2e.XXXXXX")
STALE_ENTRY_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-stale-entry-e2e.XXXXXX")

cleanup() {
  rm -rf "$TEST_ROOT" "$CONFLICT_ROOT" "$NATIVE_CONFLICT_ROOT" "$RECOVERY_ROOT" "$RULE_CONFLICT_ROOT" "$CODEX_RECOVERY_ROOT" "$CLAUDE_RECOVERY_ROOT" "$STALE_ENTRY_ROOT"
}
trap cleanup EXIT HUP INT TERM

printf '%s\n' "# User AGENTS instructions" > "$TEST_ROOT/AGENTS.md"
printf '%s\n' "# User CLAUDE instructions" > "$TEST_ROOT/CLAUDE.md"

"$REPO_ROOT/install/install.sh" install --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
FLOWCTL="$TEST_ROOT/.ai-flow/bin/flowctl"
"$FLOWCTL" project init --root "$TEST_ROOT" --mode existing --name "Lifecycle Project" >/dev/null

for skill_root in .agents/skills .cursor/skills .claude/skills; do
  [ -f "$TEST_ROOT/$skill_root/initialize-ai-project/SKILL.md" ]
  [ -f "$TEST_ROOT/$skill_root/initialize-ai-project/.ai-flow-managed" ]
  [ -f "$TEST_ROOT/$skill_root/profile-project-engineering/SKILL.md" ]
  [ -f "$TEST_ROOT/$skill_root/profile-project-engineering/references/recommended-sources.md" ]
done
[ -f "$TEST_ROOT/.cursor/rules/ai-flow.mdc.ai-flow-managed" ]
DOCTOR_JSON=$("$FLOWCTL" doctor --root "$TEST_ROOT" --json)
for check_name in codex-skills cursor-skills claude-skills; do
  printf '%s\n' "$DOCTOR_JSON" | grep -q "\"name\": \"$check_name\""
done

grep -q "User AGENTS instructions" "$TEST_ROOT/AGENTS.md"
grep -q "User CLAUDE instructions" "$TEST_ROOT/CLAUDE.md"
[ "$(grep -c '<!-- ai-flow:start -->' "$TEST_ROOT/AGENTS.md")" -eq 1 ]

CURSOR_MARKER="$TEST_ROOT/.cursor/skills/initialize-ai-project/.ai-flow-managed"
mv "$CURSOR_MARKER" "$CURSOR_MARKER.saved"
if "$REPO_ROOT/install/install.sh" update --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null 2>&1; then
  printf '%s\n' "update unexpectedly overwrote an unmanaged native Skill" >&2
  exit 1
fi
mv "$CURSOR_MARKER.saved" "$CURSOR_MARKER"

"$REPO_ROOT/install/install.sh" update --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
[ "$(grep -c '<!-- ai-flow:start -->' "$TEST_ROOT/AGENTS.md")" -eq 1 ]
"$FLOWCTL" doctor --root "$TEST_ROOT" >/dev/null

"$REPO_ROOT/install/install.sh" uninstall --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
[ ! -e "$TEST_ROOT/.agents/skills/initialize-ai-project" ]
[ ! -e "$TEST_ROOT/.cursor/skills/initialize-ai-project" ]
[ ! -e "$TEST_ROOT/.claude/skills/initialize-ai-project" ]
[ ! -e "$TEST_ROOT/.ai-flow/bin/flowctl" ]
[ -f "$TEST_ROOT/.ai-flow/manifest.yaml" ]
[ -f "$TEST_ROOT/docs/board/STATUS.md" ]
grep -q "User AGENTS instructions" "$TEST_ROOT/AGENTS.md"
if grep -q '<!-- ai-flow:start -->' "$TEST_ROOT/AGENTS.md"; then
  printf '%s\n' "managed AGENTS block survived uninstall" >&2
  exit 1
fi

mkdir -p "$CONFLICT_ROOT/.agents/skills/initialize-ai-project"
printf '%s\n' "user-owned skill" > "$CONFLICT_ROOT/.agents/skills/initialize-ai-project/SKILL.md"
if "$REPO_ROOT/install/install.sh" install --target "$CONFLICT_ROOT" --source "$REPO_ROOT" >/dev/null 2>&1; then
  printf '%s\n' "installer unexpectedly overwrote an unmanaged Skill" >&2
  exit 1
fi
grep -q "user-owned skill" "$CONFLICT_ROOT/.agents/skills/initialize-ai-project/SKILL.md"

mkdir -p "$NATIVE_CONFLICT_ROOT/.cursor/skills/initialize-ai-project"
printf '%s\n' "user-owned cursor skill" > "$NATIVE_CONFLICT_ROOT/.cursor/skills/initialize-ai-project/SKILL.md"
if "$REPO_ROOT/install/install.sh" install --target "$NATIVE_CONFLICT_ROOT" --source "$REPO_ROOT" >/dev/null 2>&1; then
  printf '%s\n' "installer unexpectedly overwrote an unmanaged Cursor Skill" >&2
  exit 1
fi
grep -q "user-owned cursor skill" "$NATIVE_CONFLICT_ROOT/.cursor/skills/initialize-ai-project/SKILL.md"

# A partially deleted old installation is recoverable when the remaining files
# carry AI Flow ownership markers or the legacy Cursor rule signature.
mkdir -p "$RECOVERY_ROOT/.cursor/rules" "$RECOVERY_ROOT/.cursor/skills/initialize-ai-project"
cp "$REPO_ROOT/adapters/cursor/ai-flow.mdc" "$RECOVERY_ROOT/.cursor/rules/ai-flow.mdc"
cp "$REPO_ROOT/skills/initialize-ai-project/SKILL.md" "$RECOVERY_ROOT/.cursor/skills/initialize-ai-project/SKILL.md"
printf '%s\n' "0.2.1" > "$RECOVERY_ROOT/.cursor/skills/initialize-ai-project/.ai-flow-managed"
"$REPO_ROOT/install/install.sh" install --cursor --target "$RECOVERY_ROOT" --source "$REPO_ROOT" >/dev/null
[ -f "$RECOVERY_ROOT/.ai-flow/install/version" ]
[ -f "$RECOVERY_ROOT/.cursor/rules/ai-flow.mdc.ai-flow-managed" ]
"$RECOVERY_ROOT/.ai-flow/bin/flowctl" doctor --root "$RECOVERY_ROOT" >/dev/null

# A genuinely user-owned rule at the same path must still be protected.
mkdir -p "$RULE_CONFLICT_ROOT/.cursor/rules"
printf '%s\n' "user-owned Cursor rule" > "$RULE_CONFLICT_ROOT/.cursor/rules/ai-flow.mdc"
printf '%s\n' "0.2.3" > "$RULE_CONFLICT_ROOT/.cursor/rules/ai-flow.mdc.ai-flow-managed"
if "$REPO_ROOT/install/install.sh" install --cursor --target "$RULE_CONFLICT_ROOT" --source "$REPO_ROOT" >/dev/null 2>&1; then
  printf '%s\n' "installer unexpectedly overwrote a user-owned Cursor rule" >&2
  exit 1
fi
grep -q "user-owned Cursor rule" "$RULE_CONFLICT_ROOT/.cursor/rules/ai-flow.mdc"

# A v0.1-style Codex pack had no per-Skill markers. Recover it only when the
# complete known pack and stable AI Flow signatures are present.
legacy_codex_skills="initialize-ai-project orchestrate-ai-delivery adopt-existing-project discover-product-goal plan-product-delivery research-and-design-solution specify-tests implement-work-item diagnose-and-verify review-change integrate-git-change manage-release sync-project-knowledge"
mkdir -p "$CODEX_RECOVERY_ROOT/.agents/skills"
for skill_name in $legacy_codex_skills; do
  cp -R "$REPO_ROOT/skills/$skill_name" "$CODEX_RECOVERY_ROOT/.agents/skills/$skill_name"
  rm -f "$CODEX_RECOVERY_ROOT/.agents/skills/$skill_name/.ai-flow-managed"
done
mkdir -p "$CODEX_RECOVERY_ROOT/.agents/skills/profile-project-engineering"
printf '%s\n' "user-owned new Skill" > "$CODEX_RECOVERY_ROOT/.agents/skills/profile-project-engineering/SKILL.md"
if "$REPO_ROOT/install/install.sh" install --codex --target "$CODEX_RECOVERY_ROOT" --source "$REPO_ROOT" >/dev/null 2>&1; then
  printf '%s\n' "legacy migration unexpectedly overwrote a user-owned newer Codex Skill" >&2
  exit 1
fi
grep -q 'user-owned new Skill' "$CODEX_RECOVERY_ROOT/.agents/skills/profile-project-engineering/SKILL.md"
rm -rf "$CODEX_RECOVERY_ROOT/.agents/skills/profile-project-engineering"
"$REPO_ROOT/install/install.sh" install --codex --target "$CODEX_RECOVERY_ROOT" --source "$REPO_ROOT" >/dev/null
[ -f "$CODEX_RECOVERY_ROOT/.agents/skills/profile-project-engineering/.ai-flow-managed" ]
[ "$(sed -n '1p' "$CODEX_RECOVERY_ROOT/.ai-flow/install/platforms")" = "codex" ]

# A legacy Claude entry can be recovered by its dedicated AI Flow signature.
mkdir -p "$CLAUDE_RECOVERY_ROOT/.claude/skills"
cp -R "$REPO_ROOT/adapters/claude/ai-flow" "$CLAUDE_RECOVERY_ROOT/.claude/skills/ai-flow"
rm -f "$CLAUDE_RECOVERY_ROOT/.claude/skills/ai-flow/.ai-flow-managed"
"$REPO_ROOT/install/install.sh" install --claude --target "$CLAUDE_RECOVERY_ROOT" --source "$REPO_ROOT" >/dev/null
[ -f "$CLAUDE_RECOVERY_ROOT/.claude/skills/ai-flow/.ai-flow-managed" ]
[ "$(sed -n '1p' "$CLAUDE_RECOVERY_ROOT/.ai-flow/install/platforms")" = "claude" ]

# Stale routing entries without their native Skill packs are not active
# platforms. Installing Cursor must not make doctor validate empty Codex and
# Claude directories left by a partially deleted old installation.
mkdir -p "$STALE_ENTRY_ROOT/.cursor/rules" "$STALE_ENTRY_ROOT/.claude/skills"
cp "$REPO_ROOT/adapters/cursor/ai-flow.mdc" "$STALE_ENTRY_ROOT/.cursor/rules/ai-flow.mdc"
cp "$REPO_ROOT/adapters/codex/AGENTS.block.md" "$STALE_ENTRY_ROOT/AGENTS.md"
cp "$REPO_ROOT/adapters/claude/CLAUDE.block.md" "$STALE_ENTRY_ROOT/CLAUDE.md"
cp -R "$REPO_ROOT/adapters/claude/ai-flow" "$STALE_ENTRY_ROOT/.claude/skills/ai-flow"
stale_output=$("$REPO_ROOT/install/install.sh" install --cursor --target "$STALE_ENTRY_ROOT" --source "$REPO_ROOT" 2>&1)
printf '%s\n' "$stale_output" | grep -q 'preparing AI Flow'
printf '%s\n' "$stale_output" | grep -q 'running installation health check'
[ "$(tr '\n' ',' < "$STALE_ENTRY_ROOT/.ai-flow/install/platforms")" = "cursor," ]
[ ! -e "$STALE_ENTRY_ROOT/.agents/skills/initialize-ai-project" ]
[ ! -e "$STALE_ENTRY_ROOT/.claude/skills/initialize-ai-project" ]
"$STALE_ENTRY_ROOT/.ai-flow/bin/flowctl" doctor --root "$STALE_ENTRY_ROOT" >/dev/null

printf '%s\n' "AI Flow install lifecycle E2E passed"
