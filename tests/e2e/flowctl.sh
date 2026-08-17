#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
TEST_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-e2e.XXXXXX")

cleanup() {
  rm -rf "$TEST_ROOT"
}
trap cleanup EXIT HUP INT TERM

"$REPO_ROOT/install/install.sh" install --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
FLOWCTL="$TEST_ROOT/.ai-flow/bin/flowctl"

"$FLOWCTL" project init --root "$TEST_ROOT" --mode greenfield --name "E2E Project" >/dev/null
cp "$REPO_ROOT/tests/fixtures/engineering-profile.json" "$TEST_ROOT/.ai-flow/baseline/engineering-profile.json"
WORK_ID=$("$FLOWCTL" work create --root "$TEST_ROOT" --kind bug --title "Prove the delivery loop" --acceptance "proof command passes" --scope "src/**")
RUN_ID=$("$FLOWCTL" work start --root "$TEST_ROOT" --id "$WORK_ID" --owner "e2e-agent")
CHECKPOINT_ID=$("$FLOWCTL" checkpoint save --root "$TEST_ROOT" --run "$RUN_ID" --phase implementing --summary "Prepared proof" --next "Run proof" --completed "created work item")
EVIDENCE_ID=$("$FLOWCTL" evidence run --root "$TEST_ROOT" --work "$WORK_ID" --run "$RUN_ID" --test "TEST-E2E-PROOF" --quiet -- sh -c "printf proof")

"$FLOWCTL" evidence verify --root "$TEST_ROOT" --id "$EVIDENCE_ID" >/dev/null
"$FLOWCTL" work review-ready --root "$TEST_ROOT" --id "$WORK_ID" >/dev/null

EVIDENCE_LOG="$TEST_ROOT/.ai-flow/evidence/logs/$EVIDENCE_ID.log"
cp "$EVIDENCE_LOG" "$EVIDENCE_LOG.original"
printf 'tampered\n' >> "$EVIDENCE_LOG"
if "$FLOWCTL" work complete --root "$TEST_ROOT" --id "$WORK_ID" --evidence "$EVIDENCE_ID" >/dev/null 2>&1; then
  printf '%s\n' "tampered evidence unexpectedly completed work" >&2
  exit 1
fi
mv "$EVIDENCE_LOG.original" "$EVIDENCE_LOG"

"$FLOWCTL" work complete --root "$TEST_ROOT" --id "$WORK_ID" --evidence "$EVIDENCE_ID" >/dev/null
"$FLOWCTL" validate --root "$TEST_ROOT" >/dev/null
"$FLOWCTL" render-board --root "$TEST_ROOT" >/dev/null

[ -f "$TEST_ROOT/.ai-flow/work-items/$WORK_ID.json" ]
[ -f "$TEST_ROOT/.ai-flow/runs/$RUN_ID/checkpoints/$CHECKPOINT_ID.json" ]
[ -f "$TEST_ROOT/.ai-flow/evidence/$EVIDENCE_ID.json" ]
grep -q '| 0 | 0 | 0 | 0 | 0 | 1 | 0 |' "$TEST_ROOT/docs/board/STATUS.md"
grep -q '| 1 | 0 | 0 |' "$TEST_ROOT/docs/board/STATUS.md"

printf 'AI Flow E2E passed: %s %s %s %s\n' "$WORK_ID" "$RUN_ID" "$CHECKPOINT_ID" "$EVIDENCE_ID"
