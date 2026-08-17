#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
TEST_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-install-e2e.XXXXXX")
CONFLICT_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-conflict-e2e.XXXXXX")

cleanup() {
  rm -rf "$TEST_ROOT" "$CONFLICT_ROOT"
}
trap cleanup EXIT HUP INT TERM

printf '%s\n' "# User AGENTS instructions" > "$TEST_ROOT/AGENTS.md"
printf '%s\n' "# User CLAUDE instructions" > "$TEST_ROOT/CLAUDE.md"

"$REPO_ROOT/install/install.sh" install --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
FLOWCTL="$TEST_ROOT/.ai-flow/bin/flowctl"
"$FLOWCTL" project init --root "$TEST_ROOT" --mode existing --name "Lifecycle Project" >/dev/null

grep -q "User AGENTS instructions" "$TEST_ROOT/AGENTS.md"
grep -q "User CLAUDE instructions" "$TEST_ROOT/CLAUDE.md"
[ "$(grep -c '<!-- ai-flow:start -->' "$TEST_ROOT/AGENTS.md")" -eq 1 ]

"$REPO_ROOT/install/install.sh" update --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
[ "$(grep -c '<!-- ai-flow:start -->' "$TEST_ROOT/AGENTS.md")" -eq 1 ]
"$FLOWCTL" doctor --root "$TEST_ROOT" >/dev/null

"$REPO_ROOT/install/install.sh" uninstall --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
[ ! -e "$TEST_ROOT/.agents/skills/initialize-ai-project" ]
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

printf '%s\n' "AI Flow install lifecycle E2E passed"
