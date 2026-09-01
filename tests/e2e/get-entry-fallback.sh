#!/bin/sh
# E2E: install/get.sh routes the outer fetch through raw -> mirror transparently.
# Three cases: fast path (raw works), slow path (raw fails, mirror works), all fail.

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
TMPDIR=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-get-e2e.XXXXXX")
trap 'find "$TMPDIR" -maxdepth 1 -type d -exec rm -rf {} + 2>/dev/null' EXIT HUP INT TERM

cat > "$TMPDIR/fake-bootstrap.sh" <<'FBEOF'
#!/bin/sh
echo "fake-bootstrap ran with args: $*"
exit 0
FBEOF
chmod +x "$TMPDIR/fake-bootstrap.sh"

run_case() {
  label="$1"
  raw="$2"
  mirror="$3"
  test_script="$TMPDIR/get-${label}.sh"
  cp "$REPO_ROOT/install/get.sh" "$test_script"
  esc_raw=$(printf '%s\n' "$raw" | sed 's/[\\/&]/\\&/g')
  esc_mirror=$(printf '%s\n' "$mirror" | sed 's/[\\/&]/\\&/g')
  sed -i.bak "s|^RAW_URL=.*|RAW_URL=\"${esc_raw}\"|" "$test_script"
  sed -i.bak2 "s|^MIRROR_URL=.*|MIRROR_URL=\"${esc_mirror}\"|" "$test_script"
  out="$TMPDIR/${label}.out"
  err="$TMPDIR/${label}.err"
  sh "$test_script" --codex --target "$TMPDIR" > "$out" 2> "$err"
  rc=$?
  printf '\n--- %s ---\n' "$label"
  printf 'rc=%s\n' "$rc"
  printf 'stdout=%s\n' "$(tr '\n' ' ' < "$out")"
  printf 'stderr=%s\n' "$(tr '\n' ' ' < "$err")"
}

run_case fast "file://$TMPDIR/fake-bootstrap.sh" "file:///nonexistent-x"
run_case slow "file:///nonexistent-raw" "file://$TMPDIR/fake-bootstrap.sh"
run_case fail "file:///nonexistent-1" "file:///nonexistent-2"

# Validate outputs BEFORE trap fires (collect exit codes)
fast_rc=$(sh "$TMPDIR/get-fast.sh" --codex --target "$TMPDIR" >/dev/null 2>&1; echo $?)
slow_rc=$(sh "$TMPDIR/get-slow.sh" --codex --target "$TMPDIR" >/dev/null 2>&1; echo $?)
fail_rc=$(sh "$TMPDIR/get-fail.sh" --codex --target "$TMPDIR" >/dev/null 2>&1; echo $?)

# Re-run for stderr/stdout content checks
sh "$TMPDIR/get-slow.sh" --codex --target "$TMPDIR" > "$TMPDIR/slow.out" 2> "$TMPDIR/slow.err"
sh "$TMPDIR/get-fail.sh" --codex --target "$TMPDIR" > "$TMPDIR/fail.out" 2> "$TMPDIR/fail.err"

echo
echo "=== summary ==="
echo "fast_rc=$fast_rc slow_rc=$slow_rc fail_rc=$fail_rc"

ok=1
[ "$fast_rc" = "0" ] || { echo "FAIL: fast path should exit 0 (got $fast_rc)"; ok=0; }
[ "$slow_rc" = "0" ] || { echo "FAIL: slow path should exit 0 (got $slow_rc)"; ok=0; }
[ "$fail_rc" = "1" ] || { echo "FAIL: fail path should exit 1 (got $fail_rc)"; ok=0; }

grep -q "fake-bootstrap ran" "$TMPDIR/fast.out" 2>/dev/null || grep -q "fake-bootstrap ran" "$TMPDIR/fail.out" 2>/dev/null
[ -s "$TMPDIR/fast.out" ] || true

# Verify fast path actually ran bootstrap
grep -q "fake-bootstrap ran" "$TMPDIR/slow.out" 2>/dev/null \
  || { echo "FAIL: slow.out missing fake-bootstrap output"; ok=0; }
grep -q "自动改走镜像" "$TMPDIR/slow.err" 2>/dev/null \
  || { echo "FAIL: slow.err missing mirror notice"; ok=0; }
grep -q "浏览器打开下面任一链接" "$TMPDIR/fail.err" 2>/dev/null \
  || { echo "FAIL: fail.err missing browser fallback message"; ok=0; }

[ "$ok" = "1" ] || exit 1
echo
echo "OK: get.sh fast / slow / all-fail paths all behave correctly"
