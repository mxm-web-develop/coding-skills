#!/bin/sh
set -eu
REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)
TEST_ROOT=$(mktemp -d "${TMPDIR:-/tmp}/ai-flow-upgrade-e2e.XXXXXX")
trap 'rm -rf "$TEST_ROOT"' EXIT HUP INT TERM
export AI_FLOW_BUILD_SOURCE=1
"$REPO_ROOT/install/install.sh" install --codex --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
FLOWCTL="$TEST_ROOT/.ai-flow/bin/flowctl"
"$FLOWCTL" project init --root "$TEST_ROOT" --mode existing --name "升级连续性" >/dev/null
mkdir -p "$TEST_ROOT/src"
printf 'uncommitted business code\n' > "$TEST_ROOT/src/export.txt"
WORK_ID=$("$FLOWCTL" work create --root "$TEST_ROOT" --title "继续导出功能" --acceptance "保留字段顺序" --scope src)
RUN_ID=$("$FLOWCTL" work start --root "$TEST_ROOT" --id "$WORK_ID" --owner original)
"$FLOWCTL" checkpoint save --root "$TEST_ROOT" --run "$RUN_ID" --phase implementing --summary "CSV 已开始实现" --next "补齐转义测试" --question "下载交互尚未确认" >/dev/null
# A supported old state without the new optional lifecycle fields.
sed 's/pack_version: 1.1.0/pack_version: 0.4.5/' "$TEST_ROOT/.ai-flow/manifest.yaml" > "$TEST_ROOT/manifest.tmp"
mv "$TEST_ROOT/manifest.tmp" "$TEST_ROOT/.ai-flow/manifest.yaml"
jq 'del(.workflow_version)' "$TEST_ROOT/.ai-flow/work-items/$WORK_ID.json" > "$TEST_ROOT/work.tmp"
mv "$TEST_ROOT/work.tmp" "$TEST_ROOT/.ai-flow/work-items/$WORK_ID.json"
printf '原需求：保留字段顺序；下载交互待确认。\n' > "$TEST_ROOT/.ai-flow/reports/old-notes.md"
CODE_BEFORE=$(cksum "$TEST_ROOT/src/export.txt")
"$REPO_ROOT/install/install.sh" update --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
if "$FLOWCTL" checkpoint resume --root "$TEST_ROOT" --run "$RUN_ID" --owner original >/dev/null 2>&1; then
  echo 'upgrade review was bypassed by resume' >&2; exit 1
fi
[ "$CODE_BEFORE" = "$(cksum "$TEST_ROOT/src/export.txt")" ]
jq -e --arg run "$RUN_ID" '.run_id == $run and .status == "in_progress" and .workflow_version == 1' "$TEST_ROOT/.ai-flow/work-items/$WORK_ID.json" >/dev/null
"$FLOWCTL" checkpoint latest --root "$TEST_ROOT" --run "$RUN_ID" | grep -q '下载交互尚未确认'
# The fixture already carries the original requirement in acceptance_criteria;
# explicitly inspect the saved original before declaring it reconciled.
UPGRADE_ID=$(jq -r .id "$TEST_ROOT/.ai-flow/state/upgrade.json")
grep -q '保留字段顺序' "$TEST_ROOT/.ai-flow/archive/upgrades/$UPGRADE_ID/originals/.ai-flow/reports/old-notes.md"
"$FLOWCTL" project upgrade --root "$TEST_ROOT" --mode finish --resolved .ai-flow/reports/old-notes.md --summary "保留字段顺序已在任务要求中；下载交互仍待确认；继续转义测试" >/dev/null
"$FLOWCTL" checkpoint resume --root "$TEST_ROOT" --run "$RUN_ID" --owner original >/dev/null
"$FLOWCTL" validate --root "$TEST_ROOT" --machine-only >/dev/null
"$REPO_ROOT/install/install.sh" update --target "$TEST_ROOT" --source "$REPO_ROOT" >/dev/null
[ "$UPGRADE_ID" = "$(jq -r .id "$TEST_ROOT/.ai-flow/state/upgrade.json")" ]
[ "$CODE_BEFORE" = "$(cksum "$TEST_ROOT/src/export.txt")" ]
echo 'Upgrade installation and original task continuity passed'
