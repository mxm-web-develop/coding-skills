#!/bin/sh
# AI Flow smart entry. Fetches bootstrap.sh with raw.githubusercontent.com first,
# then a mirror, so non-mirror users pay zero extra latency on the fast path.
# Override mirror via AI_FLOW_DOWNLOAD_MIRROR or disable mirror via
# AI_FLOW_BOOTSTRAP_FORCE_SKIP_MIRROR=1.
set -u

REPOSITORY="${AI_FLOW_REPOSITORY:-mxm-web-develop/coding-skills}"
MIRROR="${AI_FLOW_DOWNLOAD_MIRROR:-https://ghproxy.com}"
SKIP_MIRROR="${AI_FLOW_BOOTSTRAP_FORCE_SKIP_MIRROR:-0}"

RAW_URL="https://raw.githubusercontent.com/$REPOSITORY/main/install/bootstrap.sh"
MIRROR_URL="$MIRROR/$RAW_URL"

TMP=$(mktemp -u "${TMPDIR:-/tmp}/ai-flow-get.XXXXXX")
trap 'rm -f "$TMP"' EXIT HUP INT TERM

fetch_to_tmp() {
  url="$1"
  timeout="$2"
  curl -fsSL --max-time "$timeout" -o "$TMP" "$url" 2>/dev/null
}

if fetch_to_tmp "$RAW_URL" 12; then
  :
elif [ "$SKIP_MIRROR" = "1" ]; then
  cat >&2 <<EOF
ai-flow: 直连 raw.githubusercontent.com 失败，且 AI_FLOW_BOOTSTRAP_FORCE_SKIP_MIRROR=1 已禁用镜像。

请在浏览器打开下面的 URL，把脚本内容存成本地文件再执行：
  $RAW_URL
EOF
  exit 1
elif fetch_to_tmp "$MIRROR_URL" 25; then
  printf 'ai-flow: 直连 raw.githubusercontent.com 不通，自动改走镜像 %s\n' "$MIRROR" >&2
else
  cat >&2 <<EOF
ai-flow: 直连和镜像都无法下载 bootstrap.sh。

请在浏览器打开下面任一链接，把脚本内容存成本地文件再执行（推荐先跑 install/diagnose-update.sh 看本机哪些路径通）：
  $RAW_URL
  $MIRROR_URL
EOF
  exit 1
fi

sh "$TMP" "$@"
