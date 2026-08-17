#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
TEST_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-continuity-e2e.XXXXXX")
export AI_FLOW_BUILD_SOURCE=1

cleanup() {
  rm -rf "$TEST_ROOT"
}
trap cleanup EXIT HUP INT TERM

"$REPO_ROOT/install/install.sh" install --claude --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
FLOWCTL="$TEST_ROOT/.ai-flow/bin/flowctl"
"$FLOWCTL" project init --root "$TEST_ROOT" --mode greenfield --name "Continuity Project" >/dev/null

WORK_ID=$("$FLOWCTL" work create --root "$TEST_ROOT" --kind feature --title "Add refund status" --acceptance "Support can see the final result" --scope "src/refunds/**")
RUN_ID=$("$FLOWCTL" work start --root "$TEST_ROOT" --id "$WORK_ID" --owner "shared-agent")
CHECKPOINT_ID=$("$FLOWCTL" checkpoint save --root "$TEST_ROOT" --run "$RUN_ID" --phase researching --summary "Confirmed the refund states" --next "Confirm timeout behavior" --completed "Mapped the visible states" --question "How long may a refund remain pending?")

WORK_PATH="$TEST_ROOT/.ai-flow/work-items/$WORK_ID.json"
RUN_PATH="$TEST_ROOT/.ai-flow/runs/$RUN_ID/run.json"
CHECKPOINT_PATH="$TEST_ROOT/.ai-flow/runs/$RUN_ID/checkpoints/$CHECKPOINT_ID.json"
WORK_BEFORE=$(cksum "$WORK_PATH")
RUN_BEFORE=$(cksum "$RUN_PATH")
CHECKPOINT_BEFORE=$(cksum "$CHECKPOINT_PATH")

# Adding another IDE must preserve the one shared task and its resumable state.
"$REPO_ROOT/install/install.sh" install --cursor --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
[ "$(find "$TEST_ROOT/.ai-flow/work-items" -maxdepth 1 -name 'WI-*.json' | wc -l | tr -d ' ')" -eq 1 ]
[ "$WORK_BEFORE" = "$(cksum "$WORK_PATH")" ]
[ "$RUN_BEFORE" = "$(cksum "$RUN_PATH")" ]
[ "$CHECKPOINT_BEFORE" = "$(cksum "$CHECKPOINT_PATH")" ]
grep -q '^claude$' "$TEST_ROOT/.ai-flow/install/platforms"
grep -q '^cursor$' "$TEST_ROOT/.ai-flow/install/platforms"
[ -f "$TEST_ROOT/.claude/skills/orchestrate-ai-delivery/SKILL.md" ]
[ -f "$TEST_ROOT/.cursor/skills/orchestrate-ai-delivery/SKILL.md" ]

"$FLOWCTL" checkpoint latest --root "$TEST_ROOT" --run "$RUN_ID" | grep -q 'Confirm timeout behavior'
if "$FLOWCTL" checkpoint resume --root "$TEST_ROOT" --run "$RUN_ID" --owner "cursor-agent" >/dev/null 2>&1; then
  printf '%s\n' "ownership changed without an explicit handoff" >&2
  exit 1
fi
"$FLOWCTL" checkpoint resume --root "$TEST_ROOT" --run "$RUN_ID" --owner "cursor-agent" --handoff-from "shared-agent" >/dev/null
[ "$(find "$TEST_ROOT/.ai-flow/runs" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')" -eq 1 ]
jq -e '.owner == "cursor-agent" and .status == "running"' "$RUN_PATH" >/dev/null
jq -e '.owner == "cursor-agent" and .run_id == $run' --arg run "$RUN_ID" "$WORK_PATH" >/dev/null
jq -e '.owner == "cursor-agent" and .run_id == $run' --arg run "$RUN_ID" "$TEST_ROOT/.ai-flow/locks/$WORK_ID.json" >/dev/null
EVIDENCE_ID=$("$FLOWCTL" evidence run --root "$TEST_ROOT" --work "$WORK_ID" --run "$RUN_ID" --test "refund-continuity-smoke" --quiet -- sh -c "printf refund-ready")
"$FLOWCTL" work review-ready --root "$TEST_ROOT" --id "$WORK_ID" >/dev/null
"$FLOWCTL" work complete --root "$TEST_ROOT" --id "$WORK_ID" --evidence "$EVIDENCE_ID" >/dev/null
"$FLOWCTL" validate --root "$TEST_ROOT" --machine-only >/dev/null
"$FLOWCTL" render-board --root "$TEST_ROOT" >/dev/null
"$FLOWCTL" validate --root "$TEST_ROOT" >/dev/null

STATUS_JSON=$("$FLOWCTL" status --root "$TEST_ROOT" --json)
printf '%s\n' "$STATUS_JSON" | grep -q '"work_done": 1'
printf '%s\n' "$STATUS_JSON" | grep -q '"evidence_passed": 1'
grep -q 'Add refund status' "$TEST_ROOT/docs/board/STATUS.md"

printf 'AI Flow conversation continuity E2E passed: %s %s %s %s\n' "$WORK_ID" "$RUN_ID" "$CHECKPOINT_ID" "$EVIDENCE_ID"
