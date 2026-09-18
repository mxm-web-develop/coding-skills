package main

import (
	"flag"
	"fmt"
	"strings"
)

type intentRoute struct {
	Intent    string `json:"intent"`
	Skill     string `json:"skill"`
	ReadOnly  bool   `json:"read_only"`
	Implement bool   `json:"implement"`
}

func classifyIntent(message string) intentRoute {
	text := strings.ToLower(strings.TrimSpace(message))
	has := func(words ...string) bool {
		for _, word := range words {
			if strings.Contains(text, word) {
				return true
			}
		}
		return false
	}
	switch {
	case has("升级", "upgrade", "迁移旧记录"):
		return intentRoute{"upgrade", "upgrade-ai-project", false, false}
	case has("验收通过", "验证通过", "验收不通过", "验证不通过", "acceptance passed", "acceptance failed"):
		return intentRoute{"acceptance-feedback", "diagnose-and-verify", false, false}
	case has("调研", "适合我们", "方案评估", "research", "evaluate"):
		return intentRoute{"research-only", "research-and-design-solution", false, false}
	case has("汇报", "进度", "查看计划", "当前开发计划", "status", "progress", "show plan"):
		return intentRoute{"status", "orchestrate-ai-delivery", true, false}
	case has("添加", "新增", "新功能", "实现", "add feature", "implement"):
		return intentRoute{"feature", "discover-product-goal", false, true}
	case has("修复", "报错", "bug", "失败", "fix"):
		return intentRoute{"bug", "diagnose-and-verify", false, true}
	case has("继续", "resume", "continue"):
		return intentRoute{"resume", "orchestrate-ai-delivery", false, false}
	default:
		return intentRoute{"clarify", "orchestrate-ai-delivery", true, false}
	}
}

func runRoute(args []string) error {
	fs := flag.NewFlagSet("route", flag.ContinueOnError)
	message := fs.String("message", "", "natural language request")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return printJSON(classifyIntent(*message))
}

func runContext(args []string) error {
	fs := flag.NewFlagSet("context", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	workID := fs.String("work", "", "include one task's acceptance instructions")
	history := fs.Bool("history", false, "include recent completion summaries")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	status, err := readStatus(root)
	if err != nil {
		return err
	}
	data, err := loadBoardData(root, status)
	if err != nil {
		return err
	}
	fmt.Printf("当前开发计划：%s\n", status.CurrentVersion)
	if goal := activeBoardGoal(data); goal != nil {
		fmt.Printf("目标：%s。%s\n", compactContextText(goal.Title, 80), compactContextText(goal.Outcome, 160))
	}
	if err := requireProjectCompatible(root); err != nil {
		fmt.Println("升级检查尚未完成；先整理并核对旧记录，再继续原计划。")
	}
	active, done := 0, 0
	for _, work := range data.WorkItems {
		if goal := activeBoardGoal(data); goal != nil && work.GoalID != nil && *work.GoalID != goal.ID {
			continue
		}
		if work.Status == "done" {
			done++
			continue
		}
		if work.Status == "cancelled" {
			continue
		}
		active++
		if active <= 8 {
			fmt.Printf("- %s：%s；%s\n", compactContextText(work.Title, 100), humanStatus(work.Status), workTestSummary(data, work.ID))
		}
	}
	if active > 8 {
		fmt.Printf("另有 %d 项未完成工作，详见当前计划。\n", active-8)
	}
	fmt.Printf("已完成 %d 项；当前未完成 %d 项。下一步：%s。\n", done, active, compactContextText(humanAction(status.NextAction), 180))
	if *workID != "" {
		work, err := readWorkItem(root, *workID)
		if err != nil {
			return err
		}
		fmt.Printf("\n%s\n", work.Title)
		if work.Acceptance != nil {
			a := work.Acceptance
			fmt.Println(a.Instructions)
			for i, step := range a.Steps {
				if i < len(a.Expected) {
					fmt.Printf("%d. %s → 应看到：%s\n", i+1, step, a.Expected[i])
				}
			}
			for _, limit := range a.KnownLimitations {
				fmt.Println("已知限制：" + limit)
			}
		}
		if work.CompletionSummary != "" {
			fmt.Println("完成结果：" + work.CompletionSummary)
		}
	}
	if *history {
		count := 0
		for i := len(data.WorkItems) - 1; i >= 0 && count < 10; i-- {
			w := data.WorkItems[i]
			if w.Status == "done" {
				fmt.Printf("已完成：%s。%s\n", w.Title, w.CompletionSummary)
				count++
			}
		}
	}
	return nil
}

func compactContextText(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	chars := []rune(text)
	if len(chars) > limit {
		return string(chars[:limit]) + "…"
	}
	return text
}
