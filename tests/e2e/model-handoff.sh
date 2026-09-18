#!/bin/sh
# Deterministic role-contract regression, not a real provider/model benchmark.
set -eu
REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
TEST_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-model-handoff.XXXXXX")
trap 'rm -rf "$TEST_ROOT"' EXIT HUP INT TERM
AI_FLOW_BUILD_SOURCE=1 "$REPO_ROOT/install/install.sh" install --codex --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
FLOW="$TEST_ROOT/.ai-flow/bin/flowctl"
"$FLOW" project init --root "$TEST_ROOT" --mode existing --name 'cross-model fixture' >/dev/null
printf 'module example.test/handoff\n\ngo 1.22\n' > "$TEST_ROOT/go.mod"
printf 'package handoff\nfunc Twice(n int) int { return n * 2 }\n' > "$TEST_ROOT/value.go"
printf 'package handoff\nimport "testing"\nfunc TestTwice(t *testing.T) { for _, n := range []int{0, -2, 4} { if Twice(n) != n+n {t.Fatal(n)} } }\n' > "$TEST_ROOT/value_test.go"
WORK=$("$FLOW" work create --root "$TEST_ROOT" --title 'Double signed values' --scope value.go --scope value_test.go --acceptance 'Zero, negative and positive integers are doubled')
"$FLOW" memory scan --root "$TEST_ROOT" --work "$WORK"
"$FLOW" memory confirm-profile --root "$TEST_ROOT" --work "$WORK" --by planner --source go.mod
mkdir -p "$TEST_ROOT/.ai-flow/reports/$WORK"
cat > "$TEST_ROOT/.ai-flow/reports/$WORK/spec.json" <<'JSON'
{"outcome":"Double signed integers","interfaces":["Twice(int) int"],"invariants":["No new dependency"],"steps":["Verify signed and zero cases"],"edge_cases":["zero and negative inputs"],"entry_points":["value.go"],"sources":[],"facts":[],"tests":[{"id":"double-values","command":["go","test","-count=1","./..."],"expected":"Signed and zero cases pass"}],"stop_conditions":["Arithmetic behavior is ambiguous"],"capabilities":["Go tests and implementation"]}
JSON
"$FLOW" handoff prepare --root "$TEST_ROOT" --work "$WORK" --file "$TEST_ROOT/.ai-flow/reports/$WORK/spec.json" --by planner --executor implementer --reviewer reviewer --model fixture-only
RUN=$("$FLOW" work start --root "$TEST_ROOT" --id "$WORK" --owner implementer)
"$FLOW" context --root "$TEST_ROOT" --work "$WORK" --for executor > "$TEST_ROOT/.ai-flow/reports/$WORK/context.json"
if "$FLOW" context --root "$TEST_ROOT" --work "$WORK" --for executor --max-bytes 10 >/dev/null 2>&1; then
  echo 'partial context was allowed' >&2; exit 1
fi
"$FLOW" handoff escalate --root "$TEST_ROOT" --work "$WORK" --by implementer --reason 'Need confirmation of signed input behavior' --evidence value_test.go
if "$FLOW" handoff resolve --root "$TEST_ROOT" --work "$WORK" --by implementer --reason 'guess' >/dev/null 2>&1; then
  echo 'executor resolved a planning conflict' >&2; exit 1
fi
"$FLOW" handoff resolve --root "$TEST_ROOT" --work "$WORK" --by planner --reason 'Existing acceptance explicitly includes negative values; follow tests'
"$FLOW" evidence run --root "$TEST_ROOT" --work "$WORK" --run "$RUN" --test double-values --quiet -- go test -count=1 ./...
if "$FLOW" work review --root "$TEST_ROOT" --id "$WORK" --reviewer implementer --decision approved --summary 'self' >/dev/null 2>&1; then
  echo 'self review passed' >&2; exit 1
fi
"$FLOW" work review --root "$TEST_ROOT" --id "$WORK" --reviewer reviewer --decision approved --summary 'Fixture review of signed cases'
"$FLOW" work acceptance --root "$TEST_ROOT" --id "$WORK" --instructions 'Fixture acceptance only' --step 'Check signed cases' --expected 'All doubled'
"$FLOW" work accept --root "$TEST_ROOT" --id "$WORK" --by fixture-user --result passed --feedback 'Simulated acceptance for regression only'
"$FLOW" work complete --root "$TEST_ROOT" --id "$WORK" --summary 'Fixture closeout preserves execution contract'
"$FLOW" work show --root "$TEST_ROOT" --id "$WORK" | grep -q '"executor_model": "fixture-only"'
"$FLOW" validate --root "$TEST_ROOT" >/dev/null
printf 'cross-model contract lifecycle passed (no actual model benchmark)\n'
