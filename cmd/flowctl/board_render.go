package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func runRenderBoard(args []string) error {
	fs := flag.NewFlagSet("render-board", flag.ContinueOnError)
	rootArg := fs.String("root", "", "project root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := resolveRoot(*rootArg, true)
	if err != nil {
		return err
	}
	if err := renderBoard(root); err != nil {
		return err
	}
	fmt.Printf("Rendered human board at %s\n", filepath.Join(root, "docs", "board"))
	return nil
}

func renderBoard(root string) error {
	status, err := readStatus(root)
	if err != nil {
		return err
	}
	if !status.Initialized {
		return errors.New("project is not initialized")
	}
	data, err := loadBoardData(root, status)
	if err != nil {
		return err
	}
	boardDir := filepath.Join(root, "docs", "board")
	if err := os.MkdirAll(boardDir, 0o755); err != nil {
		return err
	}
	files := map[string]string{
		"STATUS.md":        renderStatusBoard(data),
		"ROADMAP.md":       renderRoadmapBoard(data),
		"CURRENT_STATE.md": renderCurrentStateBoard(data),
		"RELEASES.md":      renderReleasesBoard(data),
	}
	for name, content := range files {
		if err := writeBoardFile(filepath.Join(boardDir, name), content); err != nil {
			return err
		}
	}
	return nil
}

func renderStatusBoard(data boardData) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "# 项目状态\n\n%s\n\n", overallSummary(data))
	builder.WriteString("## 当前概览\n\n")
	builder.WriteString("| 项目 | 当前大版本 | 当前开发版本 | 所处阶段 | 总体状态 | 最近更新 |\n")
	builder.WriteString("|---|---|---|---|---|---|\n")
	fmt.Fprintf(&builder, "| %s | %s | %s | %s | %s | %s |\n\n",
		mdCell(data.Status.ProjectName), mdCell(majorVersion(data.Status.CurrentVersion)), mdCell(data.Status.CurrentVersion),
		mdCell(humanPhase(data.Status.Phase)), mdCell(humanStatus(data.Status.Status)), mdCell(data.Status.UpdatedAt))

	builder.WriteString("## 版本进度\n\n")
	builder.WriteString("| 小版本 | 目标 | 任务进度 | 开发中 | 待评审 | 阻塞 | 测试证据 |\n")
	builder.WriteString("|---|---|---:|---:|---:|---:|---|\n")
	versions := versionProgressRows(data)
	displayedVersions := 0
	for _, row := range versions {
		if !sameMajorVersion(row.Version, data.Status.CurrentVersion) && row.InProgress+row.Review+row.Blocked+row.Pending == 0 {
			continue
		}
		fmt.Fprintf(&builder, "| %s | %s | %d / %d | %d | %d | %d | 通过 %d / 失败 %d / 待确认 %d |\n",
			mdCell(row.Version), mdCell(row.GoalTitle), row.Done, row.Total, row.InProgress, row.Review, row.Blocked,
			row.PassedTests, row.FailedTests, row.OtherTests)
		displayedVersions++
	}
	if displayedVersions == 0 {
		builder.WriteString("| 未规划 | 尚未形成版本计划 | 0 / 0 | 0 | 0 | 0 | 尚无证据 |\n")
	}

	builder.WriteString("\n## 开发任务\n\n")
	builder.WriteString("| 小版本 | 开发任务 | 类型 | 优先级 | 状态 | 负责人 | 测试状态 |\n")
	builder.WriteString("|---|---|---|---|---|---|---|\n")
	visibleWork := 0
	for _, work := range data.WorkItems {
		if work.Status == "cancelled" {
			continue
		}
		if work.Status == "done" && !sameMajorVersion(versionForWork(data, work), data.Status.CurrentVersion) {
			continue
		}
		owner := "未分配"
		if work.Owner != nil && *work.Owner != "" {
			owner = *work.Owner
		}
		fmt.Fprintf(&builder, "| %s | %s | %s | %s | %s | %s | %s |\n",
			mdCell(versionForWork(data, work)), mdCell(work.Title), mdCell(humanKind(work.Kind)), mdCell(humanPriority(work.Priority)),
			mdCell(humanStatus(work.Status)), mdCell(owner), mdCell(workTestSummary(data, work.ID)))
		visibleWork++
	}
	if visibleWork == 0 {
		builder.WriteString("| — | 尚未创建开发任务 | — | — | — | — | — |\n")
	}

	builder.WriteString("\n## 测试与验收\n\n")
	builder.WriteString("| 测试项 | 类型 | 验证目的 | 对应任务 | 规格状态 | 执行证据 |\n")
	builder.WriteString("|---|---|---|---|---|---|\n")
	visibleTests := 0
	for _, test := range data.Tests {
		workTitle := test.WorkItemID
		if work := boardWorkByID(data, test.WorkItemID); work != nil {
			if work.Status == "done" && !sameMajorVersion(versionForWork(data, *work), data.Status.CurrentVersion) {
				continue
			}
			workTitle = work.Title
		}
		fmt.Fprintf(&builder, "| %s | %s | %s | %s | %s | %s |\n",
			mdCell(test.ID), mdCell(humanTestLevel(test.Level)), mdCell(test.Purpose), mdCell(workTitle),
			mdCell(humanStatus(test.Status)), mdCell(testEvidenceSummary(data, test)))
		visibleTests++
	}
	if visibleTests == 0 {
		builder.WriteString("| — | — | 尚未定义测试项 | — | — | — |\n")
	}

	builder.WriteString("\n## 阻塞与下一步\n\n")
	blocked := 0
	for _, work := range data.WorkItems {
		if work.Status != "blocked" {
			continue
		}
		reason := "未记录原因"
		if work.BlockedReason != nil && *work.BlockedReason != "" {
			reason = *work.BlockedReason
		}
		fmt.Fprintf(&builder, "- **%s**：%s\n", work.Title, reason)
		blocked++
	}
	if blocked == 0 {
		builder.WriteString("- 当前没有已记录的阻塞任务。\n")
	}
	fmt.Fprintf(&builder, "- 下一步：**%s**。\n", humanAction(data.Status.NextAction))
	builder.WriteString("\n> 本页由 `.ai-flow/` 自动生成。请修改机读事实，不要直接修改看板结论。\n")
	return builder.String()
}

func renderRoadmapBoard(data boardData) string {
	var builder strings.Builder
	builder.WriteString("# 产品路线图\n\n")
	goal := activeBoardGoal(data)
	if goal == nil {
		builder.WriteString("当前还没有确认中的产品目标。完成目标讨论后，这里会展示目标、版本和里程碑。\n")
	} else {
		fmt.Fprintf(&builder, "当前正在推进 **%s**，目标版本为 **%s**。%s\n\n", goal.Title, firstNonEmpty(goal.TargetRelease, "待确认"), goal.Outcome)
		fmt.Fprintf(&builder, "- 要解决的问题：%s\n", goal.Problem)
		fmt.Fprintf(&builder, "- 验收重点：%s\n", joinBoardText(goal.AcceptanceCriteria))
		fmt.Fprintf(&builder, "- 不在本轮范围：%s\n", joinBoardText(goal.NonGoals))
	}

	builder.WriteString("\n## 版本与里程碑\n\n")
	builder.WriteString("| 版本 | 里程碑 | 预期结果 | 任务进度 | 状态 | 完成门禁 |\n")
	builder.WriteString("|---|---|---|---:|---|---|\n")
	milestoneCount := 0
	activeGoal := activeBoardGoal(data)
	for _, plan := range data.Plans {
		if plan.Status == "archived" || plan.Status == "rejected" || plan.Status == "superseded" {
			continue
		}
		if activeGoal != nil && plan.GoalID != activeGoal.ID {
			continue
		}
		planGoal := boardGoalByID(data, plan.GoalID)
		for _, milestone := range plan.Milestones {
			version := milestone.TargetRelease
			if version == "" && planGoal != nil {
				version = planGoal.TargetRelease
			}
			total, done, blocked := milestoneProgress(data, plan, milestone)
			state := "待开始"
			if blocked > 0 {
				state = "阻塞"
			} else if total > 0 && done == total {
				state = "已完成"
			} else if total > 0 {
				state = "进行中"
			}
			fmt.Fprintf(&builder, "| %s | %s | %s | %d / %d | %s | %s |\n",
				mdCell(version), mdCell(milestone.Title), mdCell(milestone.Outcome), done, total, state, mdCell(joinBoardText(milestone.ExitGates)))
			milestoneCount++
		}
	}
	if milestoneCount == 0 {
		builder.WriteString("| 待确认 | 尚未拆分里程碑 | — | 0 / 0 | 待规划 | — |\n")
	}

	builder.WriteString("\n## 接下来要做\n\n")
	fmt.Fprintf(&builder, "%s。\n", humanAction(data.Status.NextAction))
	builder.WriteString("\n> 详细计划和依赖关系保存在 `.ai-flow/plans/`，本页只展示用户需要关注的内容。\n")
	return builder.String()
}

func renderCurrentStateBoard(data boardData) string {
	var builder strings.Builder
	builder.WriteString("# 当前产品与技术状态\n\n")
	fmt.Fprintf(&builder, "项目当前运行在 **%s**，处于 **%s** 阶段。", data.Status.CurrentVersion, humanPhase(data.Status.Phase))
	if goal := activeBoardGoal(data); goal != nil {
		fmt.Fprintf(&builder, "当前目标是 **%s**：%s", goal.Title, goal.Outcome)
	}
	builder.WriteString("\n\n## 已确认需求\n\n")
	builder.WriteString("| 需求 | 内容 | 状态 | 验收条件 |\n")
	builder.WriteString("|---|---|---|---|\n")
	requirementCount := 0
	activeGoal := activeBoardGoal(data)
	for _, requirement := range data.Requirements {
		if requirement.Status == "archived" || requirement.Status == "superseded" || requirement.Status == "rejected" {
			continue
		}
		if activeGoal != nil && requirement.GoalID != activeGoal.ID {
			continue
		}
		fmt.Fprintf(&builder, "| %s | %s | %s | %s |\n", mdCell(requirement.ID), mdCell(requirement.Statement), mdCell(humanStatus(requirement.Status)), mdCell(joinBoardText(requirement.AcceptanceCriteria)))
		requirementCount++
	}
	if requirementCount == 0 {
		builder.WriteString("| — | 尚未确认需求 | — | — |\n")
	}

	builder.WriteString("\n## 开发方案决策\n\n")
	builder.WriteString("| 决策 | 采用方案 | 主要影响 | 状态 |\n")
	builder.WriteString("|---|---|---|---|\n")
	decisionCount := 0
	for _, decision := range data.Decisions {
		if decision.Status == "archived" || decision.Status == "superseded" || decision.Status == "rejected" {
			continue
		}
		fmt.Fprintf(&builder, "| %s | %s | %s | %s |\n", mdCell(decision.Title), mdCell(decision.Decision), mdCell(joinBoardText(decision.Consequences)), mdCell(humanStatus(decision.Status)))
		decisionCount++
	}
	if decisionCount == 0 {
		builder.WriteString("| — | 尚未记录技术方案决策 | — | — |\n")
	}

	builder.WriteString("\n## 技术栈与工程约束\n\n")
	if data.Engineering == nil {
		builder.WriteString("工程画像尚未生成。完成技术栈识别后，这里会显示语言、框架、架构和视觉测试要求。\n")
	} else {
		builder.WriteString("| 项目 | 当前选择 |\n|---|---|\n")
		fmt.Fprintf(&builder, "| 编程语言 | %s |\n", mdCell(technologyList(data.Engineering.Languages)))
		fmt.Fprintf(&builder, "| 框架 | %s |\n", mdCell(technologyList(data.Engineering.Frameworks)))
		fmt.Fprintf(&builder, "| 架构方式 | %s |\n", mdCell(data.Engineering.Architecture.Style))
		fmt.Fprintf(&builder, "| 开发 Playbook | %s |\n", mdCell(joinBoardText(data.Engineering.SelectedPlaybooks)))
		visual := "不要求"
		if data.Engineering.VisualTesting.Required {
			visual = fmt.Sprintf("要求；工具：%s；浏览器：%s；视口：%s", firstNonEmpty(data.Engineering.VisualTesting.Tool, "项目现有工具"), joinBoardText(data.Engineering.VisualTesting.Browsers), joinBoardText(data.Engineering.VisualTesting.Viewports))
		}
		fmt.Fprintf(&builder, "| 视觉验证 | %s |\n", mdCell(visual))
	}

	builder.WriteString("\n## 当前边界与风险\n\n")
	if goal := activeBoardGoal(data); goal != nil {
		fmt.Fprintf(&builder, "- 本轮不做：%s\n", joinBoardText(goal.NonGoals))
		fmt.Fprintf(&builder, "- 已知风险：%s\n", joinBoardText(goal.Risks))
	} else {
		builder.WriteString("- 尚未记录当前目标的边界和风险。\n")
	}
	if data.Engineering != nil && len(data.Engineering.Unknowns) > 0 {
		fmt.Fprintf(&builder, "- 工程未知项：%s\n", joinBoardText(data.Engineering.Unknowns))
	}
	builder.WriteString("\n> 本页只展示当前有效需求和决策；已替代方案保存在 `.ai-flow/archive/`。\n")
	return builder.String()
}

func renderReleasesBoard(data boardData) string {
	var builder strings.Builder
	builder.WriteString("# 版本记录\n\n")
	if len(data.Releases) == 0 {
		fmt.Fprintf(&builder, "目前还没有可验证的发布记录。当前开发版本是 **%s**；发布后这里会显示变更、测试证据、已知问题和回滚方式。\n", data.Status.CurrentVersion)
		return builder.String()
	}
	releases := append([]boardRelease(nil), data.Releases...)
	sort.Slice(releases, func(i, j int) bool { return compareVersions(releases[i].Version, releases[j].Version) > 0 })
	builder.WriteString("| 版本 | 状态 | 主要变化 | 开发任务 | 验证结果 | 更新时间 |\n")
	builder.WriteString("|---|---|---|---:|---|---|\n")
	for _, release := range releases {
		passed, failed, other := evidenceCountsForIDs(data, release.EvidenceIDs)
		fmt.Fprintf(&builder, "| %s | %s | %s | %d | 通过 %d / 失败 %d / 待确认 %d | %s |\n",
			mdCell(release.Version), mdCell(humanStatus(release.Status)), mdCell(release.Summary), len(release.WorkItemIDs), passed, failed, other, mdCell(release.UpdatedAt))
	}
	for _, release := range releases {
		fmt.Fprintf(&builder, "\n## %s\n\n", release.Version)
		fmt.Fprintf(&builder, "%s\n\n", release.Summary)
		fmt.Fprintf(&builder, "- 上一个版本：%s\n", firstNonEmpty(release.PreviousVersion, "无"))
		fmt.Fprintf(&builder, "- 包含任务：%s\n", releaseWorkTitles(data, release.WorkItemIDs))
		fmt.Fprintf(&builder, "- 已知问题：%s\n", joinBoardText(release.KnownIssues))
		fmt.Fprintf(&builder, "- 迁移说明：%s\n", firstNonEmpty(release.Migration, "无需迁移"))
		fmt.Fprintf(&builder, "- 回滚方式：%s\n", release.Rollback)
	}
	builder.WriteString("\n> 发布结论来自 `.ai-flow/releases/` 和关联 Evidence，不依据聊天描述生成。\n")
	return builder.String()
}

func milestoneProgress(data boardData, plan boardPlan, milestone boardMilestone) (total, done, blocked int) {
	for _, workID := range plan.WorkItemIDs {
		work := boardWorkByID(data, workID)
		if work == nil || !intersects(work.RequirementIDs, milestone.RequirementIDs) || work.Status == "cancelled" {
			continue
		}
		total++
		if work.Status == "done" {
			done++
		}
		if work.Status == "blocked" {
			blocked++
		}
	}
	return total, done, blocked
}

func technologyList(items []boardTechnology) string {
	values := make([]string, 0, len(items))
	for _, item := range items {
		value := item.Name
		if item.Version != "" {
			value += " " + item.Version
		}
		values = append(values, value)
	}
	return joinBoardText(values)
}

func releaseWorkTitles(data boardData, ids []string) string {
	titles := make([]string, 0, len(ids))
	for _, id := range ids {
		if work := boardWorkByID(data, id); work != nil {
			titles = append(titles, work.Title)
		} else {
			titles = append(titles, id)
		}
	}
	return joinBoardText(titles)
}

func writeBoardFile(path, content string) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".flowctl-board-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if err := temp.Chmod(0o644); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.WriteString(content); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}
