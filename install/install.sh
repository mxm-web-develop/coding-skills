#!/bin/sh
set -eu

PACK_VERSION="0.3.0"
COMMAND="install"
TARGET_DIR=""
SOURCE_DIR="${AI_FLOW_SOURCE:-}"
PROFILE="core"
SELECT_CURSOR=0
SELECT_CODEX=0
SELECT_CLAUDE=0
PLATFORM_SELECTION_SEEN=0

CORE_SKILLS="initialize-ai-project orchestrate-ai-delivery adopt-existing-project clean-project-workspace discover-product-goal plan-product-delivery profile-project-engineering research-and-design-solution specify-tests implement-work-item diagnose-and-verify review-change integrate-git-change manage-release sync-project-knowledge"
LEGACY_CODEX_SKILLS="initialize-ai-project orchestrate-ai-delivery adopt-existing-project discover-product-goal plan-product-delivery research-and-design-solution specify-tests implement-work-item diagnose-and-verify review-change integrate-git-change manage-release sync-project-knowledge"

usage() {
  printf '%s\n' "Usage: install.sh [install|update|uninstall] [--cursor] [--codex] [--claude] [--all] [--target PATH] [--source PATH] [--profile core]"
}

fail() {
  printf 'ai-flow installer: %s\n' "$1" >&2
  exit 1
}

info() {
  printf 'ai-flow installer: %s\n' "$1" >&2
}

select_platform() {
  case "$1" in
    cursor) SELECT_CURSOR=1 ;;
    codex) SELECT_CODEX=1 ;;
    claude|claude-code) SELECT_CLAUDE=1 ;;
    all) SELECT_CURSOR=1; SELECT_CODEX=1; SELECT_CLAUDE=1 ;;
    *) fail "unsupported platform: $1" ;;
  esac
  PLATFORM_SELECTION_SEEN=1
}

parse_platform_selection() {
  selection_value="$1"
  old_ifs=$IFS
  IFS=', '
  for selected_platform in $selection_value; do
    [ -n "$selected_platform" ] && select_platform "$selected_platform"
  done
  IFS=$old_ifs
}

[ -z "${AI_FLOW_PLATFORMS:-}" ] || parse_platform_selection "$AI_FLOW_PLATFORMS"

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
    --cursor)
      select_platform cursor
      shift
      ;;
    --codex)
      select_platform codex
      shift
      ;;
    --claude|--claude-code)
      select_platform claude
      shift
      ;;
    --all)
      select_platform all
      shift
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
    rm -rf "$TARGET_DIR/.cursor/skills/$skill_name"
    rm -rf "$TARGET_DIR/.claude/skills/$skill_name"
  done
  rm -rf "$TARGET_DIR/.claude/skills/ai-flow"
  rm -f "$TARGET_DIR/.cursor/rules/ai-flow.mdc" "$TARGET_DIR/.cursor/rules/ai-flow.mdc.ai-flow-managed"
  rm -f "$TARGET_DIR/.ai-flow/bin/flowctl" "$TARGET_DIR/.ai-flow/bin/flowctl.exe"
  rm -f "$TARGET_DIR/.ai-flow/install/version" "$TARGET_DIR/.ai-flow/install/profile"
  rm -f "$TARGET_DIR/.ai-flow/install/platforms"
  rm -rf "$TARGET_DIR/.ai-flow/runtime/schemas"
  remove_block "$TARGET_DIR/AGENTS.md"
  remove_block "$TARGET_DIR/CLAUDE.md"
}

is_managed_skill() {
  candidate_skill="$1"
  [ -d "$candidate_skill" ] && [ -f "$candidate_skill/SKILL.md" ] && [ -f "$candidate_skill/.ai-flow-managed" ]
}

is_managed_cursor_rule() {
  candidate_rule="$1"
  [ -f "$candidate_rule" ] || return 1
  grep -q '^description: Route repository development through AI Flow$' "$candidate_rule" \
    && grep -q 'This repository uses AI Flow' "$candidate_rule" \
    && grep -q '\.ai-flow/manifest.yaml' "$candidate_rule" \
    && grep -q 'orchestrate-ai-delivery' "$candidate_rule"
}

is_managed_claude_entry() {
  candidate_entry="$1"
  is_managed_skill "$candidate_entry" && return 0
  [ -f "$candidate_entry/SKILL.md" ] || return 1
  grep -q '^name: ai-flow$' "$candidate_entry/SKILL.md" \
    && grep -q '^# AI Flow Claude Entry$' "$candidate_entry/SKILL.md" \
    && grep -q 'orchestrate-ai-delivery/SKILL.md' "$candidate_entry/SKILL.md" \
    && grep -q '\.ai-flow/manifest.yaml' "$candidate_entry/SKILL.md"
}

is_legacy_codex_pack() {
  legacy_root="$1"
  for legacy_skill in $LEGACY_CODEX_SKILLS; do
    legacy_file="$legacy_root/$legacy_skill/SKILL.md"
    [ -f "$legacy_file" ] || return 1
    grep -q "^name: $legacy_skill$" "$legacy_file" || return 1
  done
  grep -q '\.ai-flow/manifest.yaml' "$legacy_root/orchestrate-ai-delivery/SKILL.md" \
    && grep -q 'flowctl project init' "$legacy_root/initialize-ai-project/SKILL.md" \
    && grep -q '^# Sync Project Knowledge$' "$legacy_root/sync-project-knowledge/SKILL.md"
}

has_complete_skill_pack() {
  skill_pack_root="$1"
  for required_skill in $CORE_SKILLS; do
    [ -f "$skill_pack_root/$required_skill/SKILL.md" ] || return 1
  done
}

INSTALL_MARKER="$TARGET_DIR/.ai-flow/install/version"
if [ "$COMMAND" = "update" ] || [ "$COMMAND" = "uninstall" ]; then
  [ -f "$INSTALL_MARKER" ] || fail "no managed AI Flow installation found at target"
fi

if [ "$COMMAND" = "uninstall" ] && [ "$PLATFORM_SELECTION_SEEN" -eq 1 ]; then
  fail "platform flags are not supported with uninstall; uninstall removes the complete managed installation"
fi

if [ "$PLATFORM_SELECTION_SEEN" -eq 0 ]; then
  if [ "$COMMAND" = "update" ] && [ -f "$TARGET_DIR/.ai-flow/install/platforms" ]; then
    parse_platform_selection "$(tr '\n' ',' < "$TARGET_DIR/.ai-flow/install/platforms")"
  else
    select_platform all
  fi
fi

# Platform flags add capabilities to an existing installation. Always refresh
# every previously active platform so their generated Skill copies cannot drift
# from the shared runtime version.
if [ "$COMMAND" != "uninstall" ] && [ -f "$TARGET_DIR/.ai-flow/install/platforms" ]; then
  parse_platform_selection "$(tr '\n' ',' < "$TARGET_DIR/.ai-flow/install/platforms")"
fi

selected_platform_display=""
[ "$SELECT_CURSOR" -eq 0 ] || selected_platform_display="cursor"
[ "$SELECT_CODEX" -eq 0 ] || selected_platform_display="${selected_platform_display:+$selected_platform_display,}codex"
[ "$SELECT_CLAUDE" -eq 0 ] || selected_platform_display="${selected_platform_display:+$selected_platform_display,}claude"
info "preparing AI Flow $PACK_VERSION for $TARGET_DIR (platforms: $selected_platform_display)"

if [ "$COMMAND" = "install" ] && [ ! -f "$INSTALL_MARKER" ]; then
  LEGACY_CODEX_PACK=0
  if [ "$SELECT_CODEX" -eq 1 ] && is_legacy_codex_pack "$TARGET_DIR/.agents/skills"; then
    LEGACY_CODEX_PACK=1
  fi
  for skill_name in $CORE_SKILLS; do
    codex_skill="$TARGET_DIR/.agents/skills/$skill_name"
    cursor_skill="$TARGET_DIR/.cursor/skills/$skill_name"
    claude_skill="$TARGET_DIR/.claude/skills/$skill_name"
    LEGACY_CODEX_SKILL=0
    case " $LEGACY_CODEX_SKILLS " in
      *" $skill_name "*) [ "$LEGACY_CODEX_PACK" -eq 0 ] || LEGACY_CODEX_SKILL=1 ;;
    esac
    [ "$SELECT_CODEX" -eq 0 ] || [ ! -e "$codex_skill" ] || is_managed_skill "$codex_skill" || [ "$LEGACY_CODEX_SKILL" -eq 1 ] || fail "existing unmanaged Skill would be overwritten: .agents/skills/$skill_name"
    [ "$SELECT_CURSOR" -eq 0 ] || [ ! -e "$cursor_skill" ] || is_managed_skill "$cursor_skill" || fail "existing unmanaged Skill would be overwritten: .cursor/skills/$skill_name"
    [ "$SELECT_CLAUDE" -eq 0 ] || [ ! -e "$claude_skill" ] || is_managed_skill "$claude_skill" || fail "existing unmanaged Skill would be overwritten: .claude/skills/$skill_name"
  done
  claude_entry="$TARGET_DIR/.claude/skills/ai-flow"
  cursor_rule="$TARGET_DIR/.cursor/rules/ai-flow.mdc"
  [ "$SELECT_CLAUDE" -eq 0 ] || [ ! -e "$claude_entry" ] || is_managed_claude_entry "$claude_entry" || fail "existing unmanaged Claude ai-flow entry would be overwritten"
  [ "$SELECT_CURSOR" -eq 0 ] || [ ! -e "$cursor_rule" ] || is_managed_cursor_rule "$cursor_rule" || fail "existing unmanaged Cursor ai-flow rule would be overwritten"
fi

if [ -f "$INSTALL_MARKER" ]; then
  installed_pack_version=$(sed -n '1p' "$INSTALL_MARKER")
  for skill_name in $CORE_SKILLS; do
    if [ "$SELECT_CURSOR" -eq 1 ]; then
      native_skill="$TARGET_DIR/.cursor/skills/$skill_name"
      [ ! -e "$native_skill" ] || [ -f "$native_skill/.ai-flow-managed" ] || fail "existing unmanaged native Skill would be overwritten: .cursor/skills/$skill_name"
    fi
    if [ "$SELECT_CLAUDE" -eq 1 ]; then
      native_skill="$TARGET_DIR/.claude/skills/$skill_name"
      [ ! -e "$native_skill" ] || [ -f "$native_skill/.ai-flow-managed" ] || fail "existing unmanaged native Skill would be overwritten: .claude/skills/$skill_name"
    fi
    if [ "$SELECT_CODEX" -eq 1 ]; then
      native_skill="$TARGET_DIR/.agents/skills/$skill_name"
      if [ -e "$native_skill" ] && [ ! -f "$native_skill/.ai-flow-managed" ] && [ "$installed_pack_version" != "0.1.0" ]; then
        fail "existing unmanaged native Skill would be overwritten: .agents/skills/$skill_name"
      fi
    fi
  done
  if [ "$SELECT_CURSOR" -eq 1 ]; then
    cursor_rule="$TARGET_DIR/.cursor/rules/ai-flow.mdc"
    [ ! -e "$cursor_rule" ] || is_managed_cursor_rule "$cursor_rule" || fail "existing unmanaged Cursor ai-flow rule would be overwritten"
  fi
  if [ "$SELECT_CLAUDE" -eq 1 ]; then
    claude_entry="$TARGET_DIR/.claude/skills/ai-flow"
    [ ! -e "$claude_entry" ] || is_managed_claude_entry "$claude_entry" || fail "existing unmanaged Claude ai-flow entry would be overwritten"
  fi
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
if [ "${AI_FLOW_BUILD_SOURCE:-0}" = "1" ]; then
  info "building flowctl from source"
  command -v go >/dev/null 2>&1 || fail "AI_FLOW_BUILD_SOURCE=1 requires Go"
  [ -f "$SOURCE_DIR/go.mod" ] || fail "AI_FLOW_BUILD_SOURCE=1 requires repository source"
  (cd "$SOURCE_DIR" && go build -o "$BUILD_DIR/flowctl" ./cmd/flowctl)
  RUNTIME_SOURCE="$BUILD_DIR/flowctl"
elif [ -x "$PACKAGED_RUNTIME" ]; then
  info "using packaged flowctl for $OS_NAME/$ARCH_NAME"
  RUNTIME_SOURCE="$PACKAGED_RUNTIME"
elif command -v go >/dev/null 2>&1 && [ -f "$SOURCE_DIR/go.mod" ]; then
  (cd "$SOURCE_DIR" && go build -o "$BUILD_DIR/flowctl" ./cmd/flowctl)
  RUNTIME_SOURCE="$BUILD_DIR/flowctl"
else
  fail "no compatible flowctl binary found and Go is unavailable"
fi

mkdir -p "$TARGET_DIR/.ai-flow/bin" "$TARGET_DIR/.ai-flow/install" "$TARGET_DIR/.ai-flow/runtime"
[ "$SELECT_CODEX" -eq 0 ] || mkdir -p "$TARGET_DIR/.agents/skills"
[ "$SELECT_CURSOR" -eq 0 ] || mkdir -p "$TARGET_DIR/.cursor/skills" "$TARGET_DIR/.cursor/rules"
[ "$SELECT_CLAUDE" -eq 0 ] || mkdir -p "$TARGET_DIR/.claude/skills"

info "installing shared runtime and IDE Skill copies"

for skill_name in $CORE_SKILLS; do
  [ -f "$SOURCE_DIR/skills/$skill_name/SKILL.md" ] || fail "missing source Skill: $skill_name"
  selected_skill_roots=""
  [ "$SELECT_CODEX" -eq 0 ] || selected_skill_roots="$selected_skill_roots .agents/skills"
  [ "$SELECT_CURSOR" -eq 0 ] || selected_skill_roots="$selected_skill_roots .cursor/skills"
  [ "$SELECT_CLAUDE" -eq 0 ] || selected_skill_roots="$selected_skill_roots .claude/skills"
  for skill_root in $selected_skill_roots; do
    skill_target="$TARGET_DIR/$skill_root/$skill_name"
    rm -rf "$skill_target"
    cp -R "$SOURCE_DIR/skills/$skill_name" "$skill_target"
    printf '%s\n' "$PACK_VERSION" > "$skill_target/.ai-flow-managed"
  done
done

if [ "$SELECT_CLAUDE" -eq 1 ]; then
  rm -rf "$TARGET_DIR/.claude/skills/ai-flow"
  cp -R "$SOURCE_DIR/adapters/claude/ai-flow" "$TARGET_DIR/.claude/skills/ai-flow"
  printf '%s\n' "$PACK_VERSION" > "$TARGET_DIR/.claude/skills/ai-flow/.ai-flow-managed"
  upsert_block "$TARGET_DIR/CLAUDE.md" "$SOURCE_DIR/adapters/claude/CLAUDE.block.md"
fi
if [ "$SELECT_CURSOR" -eq 1 ]; then
  cp "$SOURCE_DIR/adapters/cursor/ai-flow.mdc" "$TARGET_DIR/.cursor/rules/ai-flow.mdc"
  printf '%s\n' "$PACK_VERSION" > "$TARGET_DIR/.cursor/rules/ai-flow.mdc.ai-flow-managed"
fi
if [ "$SELECT_CODEX" -eq 1 ]; then
  upsert_block "$TARGET_DIR/AGENTS.md" "$SOURCE_DIR/adapters/codex/AGENTS.block.md"
fi
cp "$RUNTIME_SOURCE" "$TARGET_DIR/.ai-flow/bin/flowctl"
chmod 0755 "$TARGET_DIR/.ai-flow/bin/flowctl"
rm -rf "$TARGET_DIR/.ai-flow/runtime/schemas"
cp -R "$SOURCE_DIR/schemas" "$TARGET_DIR/.ai-flow/runtime/schemas"

printf '%s\n' "$PACK_VERSION" > "$TARGET_DIR/.ai-flow/install/version"
printf '%s\n' "$PROFILE" > "$TARGET_DIR/.ai-flow/install/profile"

ACTIVE_CURSOR=$SELECT_CURSOR
ACTIVE_CODEX=$SELECT_CODEX
ACTIVE_CLAUDE=$SELECT_CLAUDE
if [ -f "$TARGET_DIR/.ai-flow/install/platforms" ]; then
  while IFS= read -r installed_platform; do
    case "$installed_platform" in
      cursor) ACTIVE_CURSOR=1 ;;
      codex) ACTIVE_CODEX=1 ;;
      claude) ACTIVE_CLAUDE=1 ;;
    esac
  done < "$TARGET_DIR/.ai-flow/install/platforms"
else
  if [ -f "$TARGET_DIR/.cursor/rules/ai-flow.mdc" ] && has_complete_skill_pack "$TARGET_DIR/.cursor/skills"; then ACTIVE_CURSOR=1; fi
  if [ -f "$TARGET_DIR/AGENTS.md" ] && grep -q '<!-- ai-flow:start -->' "$TARGET_DIR/AGENTS.md" && has_complete_skill_pack "$TARGET_DIR/.agents/skills"; then ACTIVE_CODEX=1; fi
  if [ -f "$TARGET_DIR/.claude/skills/ai-flow/SKILL.md" ] && has_complete_skill_pack "$TARGET_DIR/.claude/skills"; then ACTIVE_CLAUDE=1; fi
fi
platform_file="$TARGET_DIR/.ai-flow/install/platforms"
: > "$platform_file"
[ "$ACTIVE_CURSOR" -eq 0 ] || printf '%s\n' cursor >> "$platform_file"
[ "$ACTIVE_CODEX" -eq 0 ] || printf '%s\n' codex >> "$platform_file"
[ "$ACTIVE_CLAUDE" -eq 0 ] || printf '%s\n' claude >> "$platform_file"
printf 'schema_version: 1\nprofile: %s\nplatforms:\n  cursor: %s\n  codex: %s\n  claude_code: %s\n' "$PROFILE" "$ACTIVE_CURSOR" "$ACTIVE_CODEX" "$ACTIVE_CLAUDE" > "$TARGET_DIR/.ai-flow/capabilities.yaml"

info "running installation health check"
"$TARGET_DIR/.ai-flow/bin/flowctl" doctor --root "$TARGET_DIR"
active_platform_display=$(tr '\n' ',' < "$platform_file" | sed 's/,$//')
printf 'Active IDE platforms: %s\n' "$active_platform_display"
printf 'AI Flow %s %s completed at %s\n' "$PACK_VERSION" "$COMMAND" "$TARGET_DIR"
printf '%s\n' "Next: reload the IDE window, start a new Agent chat, then ask to initialize the project or invoke initialize-ai-project directly."
