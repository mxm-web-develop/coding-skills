#!/bin/sh
set -eu

REPOSITORY="mxm-web-develop/coding-skills"
VERSION="${AI_FLOW_VERSION:-latest}"
TARGET_DIR="${AI_FLOW_TARGET:-$(pwd -P)}"
COMMAND="${AI_FLOW_COMMAND:-install}"

fail() {
  printf 'ai-flow bootstrap: %s\n' "$1" >&2
  exit 1
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
    -h|--help)
      printf '%s\n' "Usage: bootstrap.sh [install|update] [--version latest|vX.Y.Z] [--target PATH]"
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

curl -fsSL "$DOWNLOAD_BASE/coding-skills.tar.gz" -o "$BOOTSTRAP_DIR/coding-skills.tar.gz"
curl -fsSL "$DOWNLOAD_BASE/checksums.txt" -o "$BOOTSTRAP_DIR/checksums.txt"

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

tar -xzf "$BOOTSTRAP_DIR/coding-skills.tar.gz" -C "$BOOTSTRAP_DIR"
SOURCE_DIR="$BOOTSTRAP_DIR/coding-skills"
[ -x "$SOURCE_DIR/install/install.sh" ] || fail "release package is incomplete"

"$SOURCE_DIR/install/install.sh" "$COMMAND" --target "$TARGET_DIR" --source "$SOURCE_DIR"
