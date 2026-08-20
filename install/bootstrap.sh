#!/bin/sh
set -eu

REPOSITORY="mxm-web-develop/coding-skills"
VERSION="${AI_FLOW_VERSION:-latest}"
TARGET_DIR="${AI_FLOW_TARGET:-$(pwd -P)}"
COMMAND="${AI_FLOW_COMMAND:-install}"
PLATFORM_SELECTION="${AI_FLOW_PLATFORMS:-}"

fail() {
  printf 'ai-flow bootstrap: %s\n' "$1" >&2
  exit 1
}

status() {
  printf 'ai-flow bootstrap: %s\n' "$1" >&2
}

append_platform() {
  if [ -z "$PLATFORM_SELECTION" ]; then
    PLATFORM_SELECTION="$1"
  else
    PLATFORM_SELECTION="$PLATFORM_SELECTION,$1"
  fi
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      [ "$#" -ge 2 ] || fail "--version requires latest or a tag"
      VERSION="$2"
      shift 2
      ;;
    --target)
      [ "$#" -ge 2 ] || fail "--target requires a path"
      TARGET_DIR="$2"
      shift 2
      ;;
    install|update)
      COMMAND="$1"
      shift
      ;;
    --cursor)
      append_platform cursor
      shift
      ;;
    --codex)
      append_platform codex
      shift
      ;;
    --claude|--claude-code)
      append_platform claude
      shift
      ;;
    --all)
      append_platform all
      shift
      ;;
    -h|--help)
      printf '%s\n' "Usage: bootstrap.sh [install|update] [--cursor] [--codex] [--claude] [--all] [--version latest|vX.Y.Z] [--target PATH]"
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

[ "$COMMAND" = "install" ] || [ "$COMMAND" = "update" ] || fail "command must be install or update"
[ -d "$TARGET_DIR" ] || fail "target is not a directory: $TARGET_DIR"

if [ "$VERSION" = "latest" ]; then
  DOWNLOAD_BASE="https://github.com/$REPOSITORY/releases/latest/download"
else
	printf '%s\n' "$VERSION" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$' || fail "version must be latest or a vX.Y.Z tag"
	DOWNLOAD_BASE="https://github.com/$REPOSITORY/releases/download/$VERSION"
fi

command -v curl >/dev/null 2>&1 || fail "curl is required"
command -v tar >/dev/null 2>&1 || fail "tar is required"

BOOTSTRAP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-bootstrap.XXXXXX")
cleanup() {
  rm -rf "$BOOTSTRAP_DIR"
}
trap cleanup EXIT HUP INT TERM


download_release() {
  if [ "${AI_FLOW_BOOTSTRAP_FORCE_SKIP_HTTPS:-0}" = "1" ]; then
    status "HTTPS 路径已跳过（AI_FLOW_BOOTSTRAP_FORCE_SKIP_HTTPS=1）"
  elif curl -fsSL --max-time 20 "$DOWNLOAD_BASE/coding-skills.tar.gz" -o "$BOOTSTRAP_DIR/coding-skills.tar.gz" 2>/dev/null \
       && curl -fsSL --max-time 20 "$DOWNLOAD_BASE/checksums.txt" -o "$BOOTSTRAP_DIR/checksums.txt" 2>/dev/null; then
    status "已通过 HTTPS 拉取 ${VERSION} 安装包"
    return 0
  else
    status "HTTPS 失败，改用 git+SSH 拉源码并打包（绕过 HTTPS）"
  fi

  if ! command -v git >/dev/null 2>&1; then
    status "本机没有 git，无法走 SSH 回退"
  elif rm -rf "$BOOTSTRAP_DIR/src" \
       && git clone --depth 1 --branch "$VERSION" "git@github.com:$REPOSITORY.git" "$BOOTSTRAP_DIR/src" 2>/dev/null; then
    (cd "$BOOTSTRAP_DIR/src" && tar -czf "$BOOTSTRAP_DIR/coding-skills.tar.gz" .)
    bootstrap_actual_sum=""
    if command -v sha256sum >/dev/null 2>&1; then
      bootstrap_actual_sum=$(sha256sum "$BOOTSTRAP_DIR/coding-skills.tar.gz" | awk '{ print $1 }')
    elif command -v shasum >/dev/null 2>&1; then
      bootstrap_actual_sum=$(shasum -a 256 "$BOOTSTRAP_DIR/coding-skills.tar.gz" | awk '{ print $1 }')
    else
      fail "sha256sum or shasum is required"
    fi
    printf '%s  coding-skills.tar.gz\n' "$bootstrap_actual_sum" > "$BOOTSTRAP_DIR/checksums.txt"
    status "已通过 git+SSH 拉取 ${VERSION} 源码并打包，跳过 checksum 比对"
    return 0
  fi

  cat >&2 <<EOF
无法自动下载 ${VERSION} 安装包（HTTPS 与 git+SSH 都不可用）。

请选一种手动方式继续：

方式 A — 浏览器手抄（最稳）：
  1. 浏览器打开 https://github.com/$REPOSITORY/releases/tag/$VERSION
  2. 下载 assets 里的 coding-skills.tar.gz 和 checksums.txt
  3. 放到 $BOOTSTRAP_DIR/ 下
  4. 重新执行本脚本

方式 B — SSH 拉源码自己打包：
  git clone --depth 1 --branch $VERSION git@github.com:$REPOSITORY.git /tmp/coding-skills
  cd /tmp/coding-skills && bash install/install.sh update \\
    --target "$TARGET_DIR" --source .

方式 C — 先诊断哪条路通：
  bash <(curl -fsSL https://raw.githubusercontent.com/$REPOSITORY/main/install/diagnose-update.sh)
  （如果这条也连不上，就直接浏览器打开同 URL 把 diagnose-update.sh 内容存下来跑）
EOF
  return 1
}

download_release || fail "所有自动下载路径都失败，详见上方说明"


status "verifying release checksum"
EXPECTED=$(awk '$2 == "coding-skills.tar.gz" { print $1 }' "$BOOTSTRAP_DIR/checksums.txt")
[ -n "$EXPECTED" ] || fail "release checksum does not contain coding-skills.tar.gz"
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL=$(sha256sum "$BOOTSTRAP_DIR/coding-skills.tar.gz" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL=$(shasum -a 256 "$BOOTSTRAP_DIR/coding-skills.tar.gz" | awk '{ print $1 }')
else
  fail "sha256sum or shasum is required"
fi
[ "$EXPECTED" = "$ACTUAL" ] || fail "release checksum mismatch"

status "extracting release package"
tar -xzf "$BOOTSTRAP_DIR/coding-skills.tar.gz" -C "$BOOTSTRAP_DIR"
SOURCE_DIR="$BOOTSTRAP_DIR/coding-skills"
[ -x "$SOURCE_DIR/install/install.sh" ] || fail "release package is incomplete"

status "installing into $TARGET_DIR"
AI_FLOW_PLATFORMS="$PLATFORM_SELECTION" "$SOURCE_DIR/install/install.sh" "$COMMAND" --target "$TARGET_DIR" --source "$SOURCE_DIR"
