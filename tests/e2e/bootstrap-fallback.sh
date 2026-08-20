#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
TMPDIR=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-bootstrap-fallback-e2e.XXXXXX")
cleanup() { rm -rf "$TMPDIR"; }
trap cleanup EXIT HUP INT TERM

FAKE_BIN="$TMPDIR/bin"
mkdir -p "$FAKE_BIN"

# fake curl: always fails with a connection-reset-like error
cat > "$FAKE_BIN/curl" <<'CURL_EOF'
#!/bin/sh
echo "fake curl: connection reset" >&2
exit 7
CURL_EOF
chmod +x "$FAKE_BIN/curl"

# fake git: present so command -v git succeeds, but clone/ls-remote fail
cat > "$FAKE_BIN/git" <<'GIT_EOF'
#!/bin/sh
case "$1" in
  ls-remote|clone|fetch|push|pull)
    echo "fake git: network blocked" >&2
    exit 128
    ;;
esac
exec /usr/bin/env -i PATH="/usr/bin:/bin" git "$@"
GIT_EOF
chmod +x "$FAKE_BIN/git"

PROJECT_DIR="$TMPDIR/project"
mkdir -p "$PROJECT_DIR"

# Run bootstrap.sh with faked PATH. Expect non-zero exit + recovery instructions.
set +e
PATH="$FAKE_BIN:/usr/bin:/bin" sh "$REPO_ROOT/install/bootstrap.sh" update \
  --version "v9.9.9" \
  --target "$PROJECT_DIR" \
  > "$TMPDIR/stdout.log" 2> "$TMPDIR/stderr.log"
RC=$?
set -e

echo "exit=$RC"
echo "--- stdout (first 8 lines) ---"
head -8 "$TMPDIR/stdout.log"
echo "--- stderr (full) ---"
cat "$TMPDIR/stderr.log"

if [ "$RC" -eq 0 ]; then
  printf 'FAIL: bootstrap.sh should have exited non-zero when both paths fail\n'
  exit 1
fi

grep -q "方式 A" "$TMPDIR/stderr.log" || { printf 'FAIL: missing 浏览器手抄 方式 A\n'; exit 1; }
grep -q "方式 B" "$TMPDIR/stderr.log" || { printf 'FAIL: missing SSH 拉源码 方式 B\n'; exit 1; }
grep -q "方式 C" "$TMPDIR/stderr.log" || { printf 'FAIL: missing 诊断 方式 C\n'; exit 1; }
grep -q "git clone --depth 1 --branch v9.9.9" "$TMPDIR/stderr.log" || { printf 'FAIL: missing correct version in git clone command\n'; exit 1; }
grep -q "releases/tag/v9.9.9" "$TMPDIR/stderr.log" || { printf 'FAIL: missing release URL with correct version\n'; exit 1; }

printf '\nOK: bootstrap.sh prints recovery instructions when both HTTPS and SSH fail\n'
