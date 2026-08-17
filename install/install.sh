#!/bin/sh
set -eu

PACK_VERSION="0.1.0"
COMMAND="install"
TARGET_DIR=""
SOURCE_DIR="${AI_FLOW_SOURCE:-}"
PROFILE="core"

CORE_SKILLS="initialize-ai-project orchestrate-ai-delivery adopt-existing-project discover-product-goal plan-product-delivery research-and-design-solution specify-tests implement-work-item diagnose-and-verify review-change integrate-git-change manage-release sync-project-knowledge"

usage() {
  printf '%s\n' "Usage: install.sh [install|update|uninstall] [--target PATH] [--source PATH] [--profile core]"
}

fail() {
  printf 'ai-flow installer: %s\n' "$1" >&2
  exit 1
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    install|update|uninstall)
      COMMAND="$1"
      shift
      ;;
    --target)
      [ "$#" -ge 2 ] || fail "--target requires a path"
      TARGET_DIR="$2"
      shift 2
      ;;
    --source)
      [ "$#" -ge 2 ] || fail "--source requires a path"
      SOURCE_DIR="$2"
      shift 2
      ;;
    --profile)
      [ "$#" -ge 2 ] || fail "--profile requires a value"
      PROFILE="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

[ "$PROFILE" = "core" ] || fail "only the core profile is implemented in this release"

if [ -z "$TARGET_DIR" ]; then
  TARGET_DIR=$(pwd -P)
fi
[ -d "$TARGET_DIR" ] || fail "target is not a directory: $TARGET_DIR"
TARGET_DIR=$(cd "$TARGET_DIR" && pwd -P)
[ "$TARGET_DIR" != "/" ] || fail "refusing to install at filesystem root"

if [ -z "$SOURCE_DIR" ]; then
  SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd -P)
  SOURCE_DIR=$(dirname "$SCRIPT_DIR")
fi
[ -d "$SOURCE_DIR" ] || fail "source is not a directory: $SOURCE_DIR"
SOURCE_DIR=$(cd "$SOURCE_DIR" && pwd -P)

remove_block() {
  managed_file="$1"
  [ -f "$managed_file" ] || return 0
  temp_file=$(mktemp "${TMPDIR:-/tmp}/ai-flow-block.XXXXXX")
  awk '
    /<!-- ai-flow:start -->/ { skip=1; next }
    /<!-- ai-flow:end -->/ { skip=0; next }
    !skip { print }
  ' "$managed_file" > "$temp_file"
  mv "$temp_file" "$managed_file"
}

upsert_block() {
  managed_file="$1"
  block_file="$2"
  mkdir -p "$(dirname "$managed_file")"
  if [ ! -f "$managed_file" ]; then
    : > "$managed_file"
  fi
  remove_block "$managed_file"
  printf '\n' >> "$managed_file"
  awk '{ print }' "$block_file" >> "$managed_file"
}

remove_managed_files() {
  for skill_name in $CORE_SKILLS; do
    case "$skill_name" in
      *[!a-z0-9-]*|'') fail "invalid managed skill name: $skill_name" ;;
    esac
    rm -rf "$TARGET_DIR/.agents/skills/$skill_name"
  done
  rm -rf "$TARGET_DIR/.claude/skills/ai-flow"
  rm -f "$TARGET_DIR/.cursor/rules/ai-flow.mdc"
  rm -f "$TARGET_DIR/.ai-flow/bin/flowctl" "$TARGET_DIR/.ai-flow/bin/flowctl.exe"
  rm -f "$TARGET_DIR/.ai-flow/install/version" "$TARGET_DIR/.ai-flow/install/profile"
  rm -rf "$TARGET_DIR/.ai-flow/runtime/schemas"
  remove_block "$TARGET_DIR/AGENTS.md"
  remove_block "$TARGET_DIR/CLAUDE.md"
}

INSTALL_MARKER="$TARGET_DIR/.ai-flow/install/version"
if [ "$COMMAND" = "update" ] || [ "$COMMAND" = "uninstall" ]; then
  [ -f "$INSTALL_MARKER" ] || fail "no managed AI Flow installation found at target"
fi

if [ "$COMMAND" = "install" ] && [ ! -f "$INSTALL_MARKER" ]; then
  for skill_name in $CORE_SKILLS; do
    [ ! -e "$TARGET_DIR/.agents/skills/$skill_name" ] || fail "existing unmanaged Skill would be overwritten: $skill_name"
  done
  [ ! -e "$TARGET_DIR/.claude/skills/ai-flow" ] || fail "existing unmanaged Claude ai-flow entry would be overwritten"
  [ ! -e "$TARGET_DIR/.cursor/rules/ai-flow.mdc" ] || fail "existing unmanaged Cursor ai-flow rule would be overwritten"
fi

if [ "$COMMAND" = "uninstall" ]; then
  remove_managed_files
  printf 'Removed AI Flow managed runtime, Skills, and platform entries from %s\n' "$TARGET_DIR"
  printf '%s\n' "Project state under .ai-flow and human reports under docs/board were preserved."
  exit 0
fi

[ -d "$SOURCE_DIR/skills" ] || fail "source has no skills directory; use --source or AI_FLOW_SOURCE"
[ -f "$SOURCE_DIR/adapters/codex/AGENTS.block.md" ] || fail "source adapters are incomplete"

BUILD_DIR=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-install.XXXXXX")
trap 'rm -rf "$BUILD_DIR"' EXIT HUP INT TERM

RUNTIME_SOURCE=""
OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH_NAME=$(uname -m)
case "$ARCH_NAME" in
  x86_64|amd64) ARCH_NAME="amd64" ;;
  arm64|aarch64) ARCH_NAME="arm64" ;;
  *) fail "unsupported architecture: $ARCH_NAME" ;;
esac
PACKAGED_RUNTIME="$SOURCE_DIR/dist/flowctl-$OS_NAME-$ARCH_NAME"
if [ -x "$PACKAGED_RUNTIME" ]; then
  RUNTIME_SOURCE="$PACKAGED_RUNTIME"
elif command -v go >/dev/null 2>&1 && [ -f "$SOURCE_DIR/go.mod" ]; then
  (cd "$SOURCE_DIR" && go build -o "$BUILD_DIR/flowctl" ./cmd/flowctl)
  RUNTIME_SOURCE="$BUILD_DIR/flowctl"
else
  fail "no compatible flowctl binary found and Go is unavailable"
fi

mkdir -p "$TARGET_DIR/.agents/skills" "$TARGET_DIR/.claude/skills" "$TARGET_DIR/.cursor/rules" "$TARGET_DIR/.ai-flow/bin" "$TARGET_DIR/.ai-flow/install" "$TARGET_DIR/.ai-flow/runtime"

for skill_name in $CORE_SKILLS; do
  [ -f "$SOURCE_DIR/skills/$skill_name/SKILL.md" ] || fail "missing source Skill: $skill_name"
  rm -rf "$TARGET_DIR/.agents/skills/$skill_name"
  cp -R "$SOURCE_DIR/skills/$skill_name" "$TARGET_DIR/.agents/skills/$skill_name"
done

rm -rf "$TARGET_DIR/.claude/skills/ai-flow"
cp -R "$SOURCE_DIR/adapters/claude/ai-flow" "$TARGET_DIR/.claude/skills/ai-flow"
cp "$SOURCE_DIR/adapters/cursor/ai-flow.mdc" "$TARGET_DIR/.cursor/rules/ai-flow.mdc"
cp "$RUNTIME_SOURCE" "$TARGET_DIR/.ai-flow/bin/flowctl"
chmod 0755 "$TARGET_DIR/.ai-flow/bin/flowctl"
rm -rf "$TARGET_DIR/.ai-flow/runtime/schemas"
cp -R "$SOURCE_DIR/schemas" "$TARGET_DIR/.ai-flow/runtime/schemas"

upsert_block "$TARGET_DIR/AGENTS.md" "$SOURCE_DIR/adapters/codex/AGENTS.block.md"
upsert_block "$TARGET_DIR/CLAUDE.md" "$SOURCE_DIR/adapters/claude/CLAUDE.block.md"

printf '%s\n' "$PACK_VERSION" > "$TARGET_DIR/.ai-flow/install/version"
printf '%s\n' "$PROFILE" > "$TARGET_DIR/.ai-flow/install/profile"
printf 'schema_version: 1\nprofile: %s\nplatforms:\n  cursor: detected\n  codex: detected\n  claude_code: adapter\n' "$PROFILE" > "$TARGET_DIR/.ai-flow/capabilities.yaml"

"$TARGET_DIR/.ai-flow/bin/flowctl" doctor --root "$TARGET_DIR"
printf 'AI Flow %s %s completed at %s\n' "$PACK_VERSION" "$COMMAND" "$TARGET_DIR"
printf '%s\n' "Next: ask your IDE to use initialize-ai-project, or invoke the platform-specific Skill directly."
