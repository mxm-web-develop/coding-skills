package main

import (
	"strings"
	"testing"
)

func TestLintMessage_EmptyTextIsClean(t *testing.T) {
	if got := lintMessage(""); len(got) != 0 {
		t.Fatalf("expected empty violations, got %d: %+v", len(got), got)
	}
}

func TestLintMessage_BareSectionRefsAreFlagged(t *testing.T) {
	cases := []string{
		"看一下 §38 那段",
		"按第 7 节说的做",
		"Phase 3 已经完成",
		"Module 2 没找到",
		"Step 1 是初始化",
	}
	for _, c := range cases {
		v := lintMessage(c)
		if len(v) == 0 {
			t.Errorf("expected violation for %q, got none", c)
		}
	}
}

func TestLintMessage_WorkItemIDsAreFlagged(t *testing.T) {
	cases := []string{
		"WI-20260819-a11c1109 这个任务",
		"推进 MS-2 阶段",
		"DEC-7 决定",
	}
	for _, c := range cases {
		v := lintMessage(c)
		if len(v) == 0 {
			t.Errorf("expected violation for %q, got none", c)
		}
	}
}

func TestLintMessage_CommitSHAsAreFlagged(t *testing.T) {
	cases := []string{
		"b7ca6850 这个 commit",
		"hash is fdd1b619e7a8",
	}
	for _, c := range cases {
		v := lintMessage(c)
		foundSHA := false
		for _, x := range v {
			if strings.Contains(x.Reason, "commit SHA") {
				foundSHA = true
				break
			}
		}
		if !foundSHA {
			t.Errorf("expected commit-SHA violation for %q, got %+v", c, v)
		}
	}
}

func TestLintMessage_StatusValuesAreFlagged(t *testing.T) {
	cases := []string{
		"Task status: in_progress",
		"状态: review",
		"状态: done",
		"状态: blocked",
	}
	for _, c := range cases {
		v := lintMessage(c)
		if len(v) == 0 {
			t.Errorf("expected violation for %q, got none", c)
		}
	}
}

func TestLintMessage_PlainEnglishNotFlagged(t *testing.T) {
	// Make sure common English phrasings are NOT flagged.
	cases := []string{
		"Please review the PR when you have time.",
		"All done for today.",
		"The work is blocked by a meeting.",
		"She approved the design yesterday.",
		"Approval is the next step.",
		"Looking at the agent's reasoning, the plan is solid.",
	}
	for _, c := range cases {
		v := lintMessage(c)
		if len(v) != 0 {
			t.Errorf("expected NO violation for plain English %q, got %+v", c, v)
		}
	}
}

func TestLintMessage_InternalToolNamesAreFlagged(t *testing.T) {
	cases := []string{
		"下一步用 form_decisions 收集信息",
		"先跑 form_field_guide",
		"调用 api_execute_confirm 确认",
	}
	for _, c := range cases {
		v := lintMessage(c)
		if len(v) == 0 {
			t.Errorf("expected violation for %q, got none", c)
		}
	}
}

func TestLintMessage_AbsolutePathsAreFlagged(t *testing.T) {
	cases := []string{
		"修改 /Users/me/project/file.go 里的内容",
		"看 /home/ubuntu/build.log",
	}
	for _, c := range cases {
		v := lintMessage(c)
		if len(v) == 0 {
			t.Errorf("expected violation for %q, got none", c)
		}
	}
}

func TestLintMessage_CleanTextIsClean(t *testing.T) {
	cases := []string{
		"我已经改完了，现在更新版本。",
		"你可以现在就跑一下试试。",
		"看一下「安装」那一段在讲什么。",
	}
	for _, c := range cases {
		v := lintMessage(c)
		if len(v) != 0 {
			t.Errorf("expected NO violation for clean text %q, got %+v", c, v)
		}
	}
}

func TestRunLintMessage_StdinCleanReturnsNil(t *testing.T) {
	// Cannot easily test stdin without spawning a subprocess; covered by
	// the unit tests above. This test only ensures runLintMessage does not
	// panic on empty args help flag.
	if err := runLintMessage([]string{"--help"}); err != nil {
		t.Fatalf("--help should return nil, got %v", err)
	}
}
