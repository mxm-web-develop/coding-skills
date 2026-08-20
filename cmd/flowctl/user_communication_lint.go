package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
)

// forbiddenPattern pairs a compiled regex with a human-readable rewrite hint.
type forbiddenPattern struct {
	pattern *regexp.Regexp
	reason  string
}

// forbiddenSectionRefPattern matches bare decorative section references
// (§N / 第 N 节 / Phase N / Module N / Step N). Shared between render-board
// (lint generated docs/board files) and the lint-message CLI (lint chat
// output). The user-communication-contract (禁止漏词表, v0.4.2) bans these
// from being the primary expression in user-facing text.
var forbiddenSectionRefPattern = regexp.MustCompile(`§\s*\d+|第\s*\d+\s*节|\bPhase\s*\d+|\bModule\s*\d+|\bStep\s*\d+`)

var forbiddenMessagePatterns = []forbiddenPattern{
	{
		pattern: forbiddenSectionRefPattern,
		reason:  "禁止悬空引用：用自然语言概括 + 链接，或直接复述内容；不要展示裸编号",
	},
	{
		pattern: regexp.MustCompile(`\bWI-[0-9a-zA-Z][0-9a-zA-Z-]*`),
		reason:  "禁止在用户面前贴开发任务 ID，改说这次要改的是什么",
	},
	{
		pattern: regexp.MustCompile(`\bMS-[0-9a-zA-Z][0-9a-zA-Z-]*`),
		reason:  "禁止在用户面前贴阶段 ID，改说现在进行到哪一步",
	},
	{
		pattern: regexp.MustCompile(`\b(?:DEC|ADR|REQ)-[0-9a-zA-Z][0-9a-zA-Z-]*`),
		reason:  "禁止在用户面前贴技术决策 / 技术选择 / 用户需求 ID，改说对应的内容",
	},
	{
		pattern: regexp.MustCompile(`\b(?:in_progress|not_started)\b|:\s*(?:done|blocked|review|approved|cancelled|in_progress|not_started)\b`),
		reason:  "禁止原样展示状态值，改说开发中 / 等待外部信息 / 复核中 / 已完成；普通英文里 review / done 不算违规",
	},
	{
		pattern: regexp.MustCompile(`\bform_decisions\b|\bform_field_guide\b|\bapi_execute_confirm\b`),
		reason:  "禁止展示内部模块短名，改说完整中文短语（表单-决策收集 / 表单字段说明 / 接口执行确认步骤）",
	},
	{
		pattern: regexp.MustCompile(`\b[0-9a-f]{7,40}\b`),
		reason:  "禁止在用户面前贴完整 commit SHA，改说一句话概括做了什么",
	},
	{
		pattern: regexp.MustCompile(`(?:^|[\s,;:])(?:/Users/|/home/|C:\\)[^\s,;:]{4,}`),
		reason:  "禁止在标题 / 主语里贴绝对机器路径，除非用户明确要求排查",
	},
}

// messageViolation is one hit found by lintMessage. Returned as JSON when
// the CLI is invoked.
type messageViolation struct {
	Match  string `json:"match"`
	Reason string `json:"reason"`
}

// lintMessage scans arbitrary user-facing text against the forbidden
// message list. Returns one entry per hit (empty slice when clean).
func lintMessage(text string) []messageViolation {
	var out []messageViolation
	for _, p := range forbiddenMessagePatterns {
		matches := p.pattern.FindAllString(text, -1)
		for _, m := range matches {
			out = append(out, messageViolation{Match: m, Reason: p.reason})
		}
	}
	return out
}

func runLintMessage(args []string) error {
	file := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			fmt.Println(`flowctl lint-message [--file PATH]

Reads text from stdin (default) or --file PATH and scans it against the
user-facing forbidden wording list. Prints one violation per line and
exits 1 when violations are found, 0 when clean, 2 on usage errors.`)
			return nil
		case "--file":
			if i+1 >= len(args) {
				return fmt.Errorf("--file requires a path")
			}
			file = args[i+1]
			i++
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	var data []byte
	var err error
	if file != "" && file != "-" {
		data, err = os.ReadFile(file)
	} else {
		data, err = io.ReadAll(os.Stdin)
	}
	if err != nil {
		return err
	}

	text := stripHTMLCommentsForLint(string(data))
	violations := lintMessage(text)
	if len(violations) == 0 {
		return nil
	}

	fmt.Printf("found %d violation(s):\n", len(violations))
	for _, v := range violations {
		fmt.Printf("  %-32s  →  %s\n", v.Match, v.Reason)
	}
	return fmt.Errorf("forbidden wording detected")
}
