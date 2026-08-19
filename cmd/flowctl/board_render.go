package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	files := expectedBoardFiles(data)
	for name, content := range files {
		if err := writeBoardFile(filepath.Join(boardDir, name), content); err != nil {
			return err
		}
	}
	planFiles := expectedPlanFiles(data)
	if len(planFiles) > 0 {
		plansDir := filepath.Join(boardDir, "plans")
		if err := os.MkdirAll(plansDir, 0o755); err != nil {
			return err
		}
		for name, content := range planFiles {
			if err := writeBoardFile(filepath.Join(boardDir, name), content); err != nil {
				return err
			}
		}
	}
	return nil
}

func expectedBoardFiles(data boardData) map[string]string {
	files := map[string]string{
		"STATUS.md":        renderStatusBoard(data),
		"ROADMAP.md":       renderRoadmapBoard(data),
		"CURRENT_STATE.md": renderCurrentStateBoard(data),
		"RELEASES.md":      renderReleasesBoard(data),
	}
	if index := renderPlanIndex(data); index != "" {
		files["PLANS.md"] = index
	}
	return files
}

// expectedPlanFiles returns per-version plan documents under docs/board/plans/.
// Plans without a target version are skipped. Plans whose target version has
// already been released (a boardRelease with the same Version) are also skipped
// because the published release artifacts already cover them.
func expectedPlanFiles(data boardData) map[string]string {
	files := map[string]string{}
	released := map[string]bool{}
	for _, release := range data.Releases {
		if release.Version != "" {
			released[release.Version] = true
		}
	}
	seen := map[string]bool{}
	for _, plan := range data.Plans {
		version := planVersionLabel(plan, data)
		if version == "" || released[version] || seen[version] {
			continue
		}
		seen[version] = true
		files["plans/"+version+".md"] = renderPlanDoc(data, plan)
	}
	return files
}

// planVersionLabel returns the user-visible version string for a plan, in
// priority order: explicit target_release on any of its milestones, the goal
// it belongs to, or the plan itself. Returns "" when no version can be
// determined (the caller should skip emitting a per-version document).
func planVersionLabel(plan boardPlan, data boardData) string {
	for _, milestone := range plan.Milestones {
		if milestone.TargetRelease != "" {
			return normalizeVersionLabel(milestone.TargetRelease)
		}
	}
	if goal := boardGoalByID(data, plan.GoalID); goal != nil && goal.TargetRelease != "" {
		return normalizeVersionLabel(goal.TargetRelease)
	}
	return ""
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
	builder.WriteString("| 小版本 | 目标 | 任务进度 | 开发中 | 等待检查 | 阻塞 | 测试结果 |\n")
	builder.WriteString("|---|---|---:|---:|---:|---:|---|\n")
	versions := versionProgressRows(data)
	displayedVersions := 0
	for _, row := range versions {
		if !sameMajorVersion(row.Version, data.Status.CurrentVersion) && row.InProgress+row.Review+row.Blocked+row.Pending == 0 {
			continue
		}
		fmt.Fprintf(&builder, "| %s | %s%s | %d / %d | %d | %d | %d | 通过 %d / 失败 %d / 待确认 %d |\n",
			mdCell(row.Version), mdCell(row.GoalTitle), versionProgressTrace(data, row.Version), row.Done, row.Total, row.InProgress, row.Review, row.Blocked,
			row.PassedTests, row.FailedTests, row.OtherTests)
		displayedVersions++
	}
	if displayedVersions == 0 {
		builder.WriteString("| 未规划 | 尚未形成版本计划 | 0 / 0 | 0 | 0 | 0 | 尚无证据 |\n")
	}

	builder.WriteString("\n## 开发任务\n\n")
	builder.WriteString("| 小版本 | 开发任务 | 类型 | 优先级 | 状态 | 负责人 | 测试结果 |\n")
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
			owner = humanOwner(*work.Owner)
		}
		trace := hiddenTrace("work="+work.ID, optionalTrace("goal", work.GoalID), traceList("plan", planIDsForWork(data, work.ID)), traceList("requirement", work.RequirementIDs), traceList("evidence", work.EvidenceIDs))
		fmt.Fprintf(&builder, "| %s | %s%s | %s | %s | %s | %s | %s |\n",
			mdCell(versionForWork(data, work)), mdCell(work.Title), trace, mdCell(humanKind(work.Kind)), mdCell(humanPriority(work.Priority)),
			mdCell(humanStatus(work.Status)), mdCell(owner), mdCell(workTestSummary(data, work.ID)))
		visibleWork++
	}
	if visibleWork == 0 {
		builder.WriteString("| — | 尚未创建开发任务 | — | — | — | — | — |\n")
	}

	builder.WriteString("\n## 测试与验收\n\n")
	builder.WriteString("| 类型 | 检查内容 | 对应任务 | 当前状态 | 执行结果 |\n")
	builder.WriteString("|---|---|---|---|---|\n")
	visibleTests := 0
	for _, test := range data.Tests {
		workTitle := "未找到对应任务"
		if work := boardWorkByID(data, test.WorkItemID); work != nil {
			if work.Status == "done" && !sameMajorVersion(versionForWork(data, *work), data.Status.CurrentVersion) {
				continue
			}
			workTitle = work.Title
		}
		trace := hiddenTrace("test="+test.ID, "work="+test.WorkItemID, traceList("requirement", test.RequirementIDs), traceList("evidence", test.EvidenceIDs))
		fmt.Fprintf(&builder, "| %s | %s%s | %s | %s | %s |\n",
			mdCell(humanTestLevel(test.Level)), mdCell(test.Purpose), trace, mdCell(workTitle),
			mdCell(humanStatus(test.Status)), mdCell(testEvidenceSummary(data, test)))
		visibleTests++
	}
	if visibleTests == 0 {
		builder.WriteString("| — | 尚未定义测试项 | — | — | — |\n")
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
	builder.WriteString("\n> 本页根据项目中的最新开发和测试记录自动更新，请不要直接修改统计结论。\n")
	return builder.String()
}

func renderRoadmapBoard(data boardData) string {
	var builder strings.Builder
	builder.WriteString("# 产品路线图\n\n")
	goal := activeBoardGoal(data)
	if goal == nil {
		builder.WriteString("当前还没有确认中的产品目标。完成目标讨论后，这里会展示目标、版本和开发阶段。\n")
	} else {
		fmt.Fprintf(&builder, "当前正在推进 **%s**，目标版本为 **%s**。%s\n\n", goal.Title, firstNonEmpty(goal.TargetRelease, "待确认"), goal.Outcome)
		fmt.Fprintf(&builder, "- 要解决的问题：%s\n", goal.Problem)
		fmt.Fprintf(&builder, "- 验收重点：%s\n", joinBoardText(goal.AcceptanceCriteria))
		fmt.Fprintf(&builder, "- 不在本轮范围：%s\n", joinBoardText(goal.NonGoals))
	}

	builder.WriteString("\n## 版本与开发阶段\n\n")
	builder.WriteString("| 版本 | 开发阶段 | 预期结果 | 任务进度 | 状态 | 完成条件 |\n")
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
			trace := hiddenTrace("goal="+plan.GoalID, "plan="+plan.ID, "stage="+milestone.ID, traceList("requirement", milestone.RequirementIDs), traceList("work", milestoneWorkIDs(data, plan, milestone)))
			fmt.Fprintf(&builder, "| %s | %s%s | %s | %d / %d | %s | %s |\n",
				mdCell(version), mdCell(milestone.Title), trace, mdCell(milestone.Outcome), done, total, state, mdCell(joinBoardText(milestone.ExitGates)))
			milestoneCount++
		}
	}
	if milestoneCount == 0 {
		builder.WriteString("| 待确认 | 尚未安排开发阶段 | — | 0 / 0 | 待规划 | — |\n")
	}

	builder.WriteString("\n## 接下来要做\n\n")
	fmt.Fprintf(&builder, "%s。\n", humanAction(data.Status.NextAction))
	builder.WriteString("\n> 本页只展示当前目标、开发阶段和先后关系；详细技术安排由开发流程维护。\n")
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
	builder.WriteString("| 需要实现的内容 | 状态 | 验收条件 |\n")
	builder.WriteString("|---|---|---|\n")
	requirementCount := 0
	activeGoal := activeBoardGoal(data)
	for _, requirement := range data.Requirements {
		if requirement.Status == "archived" || requirement.Status == "superseded" || requirement.Status == "rejected" {
			continue
		}
		if activeGoal != nil && requirement.GoalID != activeGoal.ID {
			continue
		}
		trace := hiddenTrace("requirement="+requirement.ID, "goal="+requirement.GoalID, traceList("test", requirement.TestIDs))
		fmt.Fprintf(&builder, "| %s%s | %s | %s |\n", mdCell(requirement.Statement), trace, mdCell(humanStatus(requirement.Status)), mdCell(joinBoardText(requirement.AcceptanceCriteria)))
		requirementCount++
	}
	if requirementCount == 0 {
		builder.WriteString("| 尚未确认需求 | — | — |\n")
	}

	builder.WriteString("\n## 开发方案决策\n\n")
	builder.WriteString("| 需要确定的方向 | 当前建议或选择 | 为什么 | 主要影响 | 确认状态 |\n")
	builder.WriteString("|---|---|---|---|---|\n")
	decisionCount := 0
	for _, decision := range data.Decisions {
		if decision.Status == "archived" || decision.Status == "superseded" || decision.Status == "rejected" {
			continue
		}
		trace := hiddenTrace("decision="+decision.ID, traceList("requirement", decision.RequirementIDs), traceList("work", decision.WorkItemIDs))
		fmt.Fprintf(&builder, "| %s%s | %s | %s | %s | %s |\n", mdCell(decision.Title), trace,
			mdCell(decisionCurrentChoice(decision)), mdCell(decision.RecommendationReason),
			mdCell(joinBoardText(decision.Consequences)), mdCell(decisionConfirmationStatus(decision)))
		decisionCount++
	}
	if decisionCount == 0 {
		builder.WriteString("| — | 尚未记录技术方案决策 | — | — | — |\n")
	}

	prototypeCount := 0
	for _, decision := range data.Decisions {
		if decision.Status == "archived" || decision.Status == "superseded" || decision.Status == "rejected" {
			continue
		}
		for _, option := range decision.Options {
			if option.PrototypePath != nil && strings.TrimSpace(*option.PrototypePath) != "" {
				prototypeCount++
			}
		}
	}
	if prototypeCount > 0 {
		builder.WriteString("\n## 界面体验方案\n\n")
		builder.WriteString("这些页面用于确认布局、视觉和交互方向，不是正式产品代码。\n\n")
		builder.WriteString("| 体验方向 | 主要感受和验证重点 | 预览 | 当前选择 |\n")
		builder.WriteString("|---|---|---|---|\n")
		for _, decision := range data.Decisions {
			if decision.Status == "archived" || decision.Status == "superseded" || decision.Status == "rejected" {
				continue
			}
			for _, option := range decision.Options {
				if option.PrototypePath == nil || strings.TrimSpace(*option.PrototypePath) == "" {
					continue
				}
				focus := firstNonEmpty(option.PrototypeFocus, option.Summary, "等待补充体验说明")
				fmt.Fprintf(&builder, "| %s | %s | %s | %s |\n", mdCell(option.Name), mdCell(focus),
					prototypePreviewLink(*option.PrototypePath), mdCell(prototypeChoiceStatus(decision, option.Name)))
			}
		}
	}

	builder.WriteString("\n## 技术环境与开发约束\n\n")
	if data.Engineering == nil {
		builder.WriteString("项目的技术环境尚未确认。完成识别后，这里会显示语言、框架、代码组织方式和界面检查要求。\n")
	} else {
		builder.WriteString("| 项目 | 当前选择 |\n|---|---|\n")
		fmt.Fprintf(&builder, "| 编程语言 | %s |\n", mdCell(technologyList(data.Engineering.Languages)))
		fmt.Fprintf(&builder, "| 框架 | %s |\n", mdCell(technologyList(data.Engineering.Frameworks)))
		fmt.Fprintf(&builder, "| 代码组织方式 | %s |\n", mdCell(humanArchitectureStyle(data.Engineering.Architecture.Style)))
		fmt.Fprintf(&builder, "| 开发与测试规范 | %s |\n", mdCell(humanEngineeringGuidance(data.Engineering.SelectedPlaybooks)))
		visual := "当前版本不需要额外的界面效果检查"
		if data.Engineering.VisualTesting.Required {
			visual = fmt.Sprintf("需要；使用 %s 检查 %s 浏览器下的 %s 页面尺寸", firstNonEmpty(data.Engineering.VisualTesting.Tool, "项目现有工具"), joinBoardText(data.Engineering.VisualTesting.Browsers), joinBoardText(data.Engineering.VisualTesting.Viewports))
		}
		fmt.Fprintf(&builder, "| 界面效果检查 | %s |\n", mdCell(visual))
	}

	builder.WriteString("\n## 当前边界与风险\n\n")
	if goal := activeBoardGoal(data); goal != nil {
		fmt.Fprintf(&builder, "- 本轮不做：%s\n", joinBoardText(goal.NonGoals))
		fmt.Fprintf(&builder, "- 已知风险：%s\n", joinBoardText(goal.Risks))
	} else {
		builder.WriteString("- 尚未记录当前目标的边界和风险。\n")
	}
	if data.Engineering != nil && len(data.Engineering.Unknowns) > 0 {
		fmt.Fprintf(&builder, "- 尚未确认的技术问题：%s\n", joinBoardText(data.Engineering.Unknowns))
	}
	builder.WriteString("\n> 本页只展示当前仍然有效的需求和方案；旧方案已经归档，需要时仍可追溯。\n")
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
		trace := hiddenTrace("release="+release.ID, traceList("work", release.WorkItemIDs), traceList("evidence", release.EvidenceIDs))
		fmt.Fprintf(&builder, "| %s | %s | %s%s | %d | 通过 %d / 失败 %d / 待确认 %d | %s |\n",
			mdCell(release.Version), mdCell(humanStatus(release.Status)), mdCell(release.Summary), trace, len(release.WorkItemIDs), passed, failed, other, mdCell(release.UpdatedAt))
	}
	for _, release := range releases {
		fmt.Fprintf(&builder, "\n## %s\n\n", release.Version)
		fmt.Fprintf(&builder, "%s\n\n", release.Summary)
		fmt.Fprintf(&builder, "- 上一个版本：%s\n", firstNonEmpty(release.PreviousVersion, "无"))
		fmt.Fprintf(&builder, "- 包含任务：%s\n", releaseWorkTitles(data, release.WorkItemIDs))
		fmt.Fprintf(&builder, "- 已知问题：%s\n", joinBoardText(release.KnownIssues))
		fmt.Fprintf(&builder, "- 升级说明：%s\n", firstNonEmpty(release.Migration, "无需额外操作"))
		fmt.Fprintf(&builder, "- 恢复方式：%s\n", release.Rollback)
	}
	builder.WriteString("\n> 发布结论以实际版本记录和测试结果为准，不依据聊天中的口头描述生成。\n")
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
			titles = append(titles, "未命名任务")
		}
	}
	return joinBoardText(titles)
}

func hiddenTrace(parts ...string) string {
	values := []string{}
	for _, part := range parts {
		if part != "" {
			values = append(values, part)
		}
	}
	if len(values) == 0 {
		return ""
	}
	return "<!-- ai-flow-trace:" + strings.Join(values, " ") + " -->"
}

func optionalTrace(name string, value *string) string {
	if value == nil || *value == "" {
		return ""
	}
	return name + "=" + *value
}

func traceList(name string, values []string) string {
	if len(values) == 0 {
		return ""
	}
	return name + "=" + strings.Join(values, ",")
}

func milestoneWorkIDs(data boardData, plan boardPlan, milestone boardMilestone) []string {
	ids := []string{}
	for _, workID := range plan.WorkItemIDs {
		work := boardWorkByID(data, workID)
		if work != nil && intersects(work.RequirementIDs, milestone.RequirementIDs) && work.Status != "cancelled" {
			ids = append(ids, workID)
		}
	}
	return ids
}

func planIDsForWork(data boardData, workID string) []string {
	ids := []string{}
	for _, plan := range data.Plans {
		for _, candidate := range plan.WorkItemIDs {
			if candidate == workID {
				ids = append(ids, plan.ID)
				break
			}
		}
	}
	return ids
}

func versionProgressTrace(data boardData, version string) string {
	goalIDs := []string{}
	planIDs := []string{}
	workIDs := []string{}
	testIDs := []string{}
	evidenceIDs := []string{}
	for _, goal := range data.Goals {
		if goal.TargetRelease == version && goal.Status != "archived" && goal.Status != "superseded" && goal.Status != "rejected" {
			goalIDs = append(goalIDs, goal.ID)
		}
	}
	for _, plan := range data.Plans {
		matched := false
		if goal := boardGoalByID(data, plan.GoalID); goal != nil && goal.TargetRelease == version {
			matched = true
		}
		for _, milestone := range plan.Milestones {
			if milestone.TargetRelease == version {
				matched = true
				break
			}
		}
		if matched {
			planIDs = append(planIDs, plan.ID)
		}
	}
	for _, work := range data.WorkItems {
		if work.Status == "cancelled" || versionForWork(data, work) != version {
			continue
		}
		workIDs = append(workIDs, work.ID)
		evidenceIDs = append(evidenceIDs, work.EvidenceIDs...)
		for _, test := range data.Tests {
			if test.WorkItemID == work.ID {
				testIDs = append(testIDs, test.ID)
				evidenceIDs = append(evidenceIDs, test.EvidenceIDs...)
			}
		}
	}
	return hiddenTrace("version="+version, traceList("goal", goalIDs), traceList("plan", planIDs), traceList("work", workIDs), traceList("test", testIDs), traceList("evidence", evidenceIDs))
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


// renderPlanIndex produces docs/board/PLANS.md, an index of every per-version
// plan document. Returns "" when there are no plans to index, so the caller
// can omit the file entirely instead of writing an empty index.
func renderPlanIndex(data boardData) string {
	plans := make([]boardPlan, 0, len(data.Plans))
	for _, plan := range data.Plans {
		if planVersionLabel(plan, data) != "" {
			plans = append(plans, plan)
		}
	}
	if len(plans) == 0 {
		return ""
	}
	sort.SliceStable(plans, func(i, j int) bool {
		left := planVersionLabel(plans[i], data)
		right := planVersionLabel(plans[j], data)
		if left != right {
			return compareVersions(left, right) > 0
		}
		return plans[i].ID < plans[j].ID
	})
	released := map[string]bool{}
	for _, release := range data.Releases {
		if release.Version != "" {
			released[release.Version] = true
		}
	}
	var builder strings.Builder
	builder.WriteString("# 开发计划索引\n\n")
	builder.WriteString("本目录按目标版本号组织每一版的开发计划、阶段划分、技术选型与风险依赖。详细任务以自然语言写成，")
	builder.WriteString("与 `docs/board/STATUS.md` 中的实时进度、`docs/board/RELEASES.md` 中的发布历史保持一致。\n\n")
	builder.WriteString("## 当前计划\n\n")
	builder.WriteString("| 版本 | 目标 | 阶段数 | 任务数 | 状态 | 计划文件 |\n")
	builder.WriteString("|---|---|---:|---:|---|---|\n")
	activeCount := 0
	for _, plan := range plans {
		version := planVersionLabel(plan, data)
		goal := boardGoalByID(data, plan.GoalID)
		target := ""
		if goal != nil {
			target = goal.Title
		}
		if target == "" {
			target = plan.Title
		}
		state := humanPlanState(plan.Status)
		if released[version] {
			state = "已发布"
		}
		link := fmt.Sprintf("[%s](./plans/%s.md)", version, version)
		fmt.Fprintf(&builder, "| %s | %s | %d | %d | %s | %s |\n",
			mdCell(version), mdCell(target), len(plan.Milestones), len(plan.WorkItemIDs), mdCell(state), link)
		if !released[version] {
			activeCount++
		}
	}
	builder.WriteString("\n")
	if activeCount == 0 {
		builder.WriteString("> 当前没有进行中的开发计划。下一个版本的需求与方案讨论完成后会自动出现新条目。\n")
	} else {
		builder.WriteString("点击对应版本查看该版的完整计划。同一版本如有多份计划（例如调整后的二次拆分），以最新一份为准。\n")
	}
	return builder.String()
}

func humanPlanState(status string) string {
	switch status {
	case "active":
		return "进行中"
	case "draft":
		return "草稿"
	case "superseded":
		return "已替代"
	case "cancelled":
		return "已取消"
	}
	if status == "" {
		return "未记录"
	}
	return status
}

// normalizeVersionLabel ensures the per-version file name always carries a
// leading "v" so filenames sort correctly (v1.10.0 > v1.9.0).
func normalizeVersionLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "v") && !strings.HasPrefix(value, "V") {
		return "v" + value
	}
	if strings.HasPrefix(value, "V") {
		return "v" + value[1:]
	}
	return value
}

// renderPlanDoc produces a single per-version plan document. The output is a
// narrative document written for human readers (PM, dev team, stakeholders):
// each milestone is described by the user-visible outcome it produces, the
// task list is summarized in plain language, technical choices are explained
// in terms of strengths/weaknesses rather than ADR IDs, and any risk or
// dependency that affects this version is surfaced.
func renderPlanDoc(data boardData, plan boardPlan) string {
	version := planVersionLabel(plan, data)
	if version == "" {
		return ""
	}
	goal := boardGoalByID(data, plan.GoalID)

	var builder strings.Builder
	fmt.Fprintf(&builder, "# %s 开发计划\n\n", version)

	leadGoal := ""
	problem := ""
	outcome := ""
	if goal != nil {
		leadGoal = goal.Title
		problem = goal.Problem
		outcome = goal.Outcome
	}
	if leadGoal == "" {
		leadGoal = plan.Title
	}
	if leadGoal != "" {
		fmt.Fprintf(&builder, "> **本计划面向** %s。", leadGoal)
	}
	if problem != "" {
		fmt.Fprintf(&builder, " **要解决的问题**：%s", problem)
	}
	if outcome != "" {
		fmt.Fprintf(&builder, " **完成后能提供**：%s。", outcome)
	}
	builder.WriteString("\n\n")
	builder.WriteString("这份文档按目标版本组织，描述阶段划分、各项开发任务、技术选型与风险依赖，给团队成员、PM 和相关方阅读。\n\n")

	if goal != nil {
		if len(goal.InScope) > 0 {
			builder.WriteString("## 范围内\n\n")
			for _, item := range goal.InScope {
				builder.WriteString("- " + item + "\n")
			}
			builder.WriteString("\n")
		}
		if len(goal.NonGoals) > 0 {
			builder.WriteString("## 不在范围内\n\n")
			for _, item := range goal.NonGoals {
				builder.WriteString("- " + item + "\n")
			}
			builder.WriteString("\n")
		}
		if len(goal.AcceptanceCriteria) > 0 {
			builder.WriteString("## 验收要点\n\n")
			for _, item := range goal.AcceptanceCriteria {
				builder.WriteString("- " + item + "\n")
			}
			builder.WriteString("\n")
		}
	}

	if len(plan.Milestones) > 0 {
		builder.WriteString("## 阶段划分\n\n")
		for index, milestone := range plan.Milestones {
			title := milestone.Title
			if title == "" {
				title = fmt.Sprintf("阶段 %d", index+1)
			}
			fmt.Fprintf(&builder, "### %d. %s\n\n", index+1, title)
			if milestone.Outcome != "" {
				fmt.Fprintf(&builder, "**完成后能看到**：%s\n\n", milestone.Outcome)
			}
			if len(milestone.ExitGates) > 0 {
				builder.WriteString("**完成条件**：\n\n")
				for _, gate := range milestone.ExitGates {
					builder.WriteString("- " + gate + "\n")
				}
				builder.WriteString("\n")
			}
			milestoneWork := milestoneWorkBullets(data, plan, milestone)
			if milestoneWork != "" {
				builder.WriteString("**本阶段包含的开发任务**：\n\n")
				builder.WriteString(milestoneWork)
				builder.WriteString("\n")
			}
		}
	}

	if len(plan.WorkItemIDs) > 0 {
		builder.WriteString("## 开发任务清单\n\n")
		builder.WriteString("| 任务 | 类型 | 负责人 | 当前状态 | 测试结果 |\n")
		builder.WriteString("|---|---|---|---|---|\n")
		for _, workID := range plan.WorkItemIDs {
			work := boardWorkByID(data, workID)
			if work == nil {
				continue
			}
			owner := ""
			if work.Owner != nil {
				owner = *work.Owner
			}
			passed, failed, other := evidenceCountsForWork(data, work.ID)
			testSummary := testSummaryShort(passed, failed, other)
			fmt.Fprintf(&builder, "| %s | %s | %s | %s | %s |\n",
				mdCell(work.Title), mdCell(humanKind(work.Kind)),
				mdCell(ownerOrUnassigned(owner)), mdCell(humanStatus(work.Status)), mdCell(testSummary))
		}
		builder.WriteString("\n")
		builder.WriteString("完整任务记录在 `.ai-flow/work-items/`；当前进度同步在 `docs/board/STATUS.md`。\n\n")
	}

	chosen := decisionsForPlan(data, plan)
	if len(chosen) > 0 {
		builder.WriteString("## 技术选型\n\n")
		builder.WriteString("| 决策点 | 选择 | 原因 | 确认状态 |\n")
		builder.WriteString("|---|---|---|---|\n")
		for _, decision := range chosen {
			summary := decision.Decision
			if summary == "" {
				summary = decision.RecommendedOption
			}
			state := decisionConfirmationState(decision)
			fmt.Fprintf(&builder, "| %s | %s | %s | %s |\n",
				mdCell(decision.Title), mdCell(summary),
				mdCell(decision.RecommendationReason), mdCell(state))
		}
		builder.WriteString("\n")
	} else if data.Engineering != nil {
		if stack := technologyList(data.Engineering.Languages); stack != "" {
			fmt.Fprintf(&builder, "## 当前技术环境\n\n本计划基于仓库当前的技术环境：%s。\n\n", stack)
		}
	}

	if goal != nil && len(goal.Risks) > 0 {
		builder.WriteString("## 风险与依赖\n\n")
		for _, risk := range goal.Risks {
			builder.WriteString("- " + risk + "\n")
		}
		builder.WriteString("\n")
	}

	builder.WriteString("---\n\n")
	builder.WriteString("**相关材料**\n\n")
	builder.WriteString("- 当前进度：`docs/board/STATUS.md`\n")
	builder.WriteString("- 路线图：`docs/board/ROADMAP.md`\n")
	builder.WriteString("- 技术与体验决策明细：`docs/board/CURRENT_STATE.md`\n")
	builder.WriteString("- 已发布版本：`docs/board/RELEASES.md`\n")
	builder.WriteString("- 全部计划索引：`docs/board/PLANS.md`\n")
	return builder.String()
}

func ownerOrUnassigned(value string) string {
	if value == "" {
		return "未指派"
	}
	return value
}

func testSummaryShort(passed, failed, other int) string {
	if passed == 0 && failed == 0 && other == 0 {
		return "尚无测试记录"
	}
	parts := []string{}
	if passed > 0 {
		parts = append(parts, fmt.Sprintf("通过 %d", passed))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("失败 %d", failed))
	}
	if other > 0 {
		parts = append(parts, fmt.Sprintf("待确认 %d", other))
	}
	return strings.Join(parts, "，")
}

func decisionConfirmationState(decision boardDecision) string {
	if decision.Confirmation == nil {
		return "待确认"
	}
	switch decision.Confirmation.Status {
	case "confirmed":
		return "已确认"
	case "rejected":
		return "已驳回"
	case "superseded":
		return "已替代"
	}
	return "待确认"
}

// decisionsForPlan returns accepted or recommended decisions whose requirement
// IDs or work item IDs overlap with the plan, in the order they were made.
func decisionsForPlan(data boardData, plan boardPlan) []boardDecision {
	out := make([]boardDecision, 0, len(data.Decisions))
	for _, decision := range data.Decisions {
		if decision.Status == "superseded" || decision.Status == "cancelled" {
			continue
		}
		if intersectsAny(decision.RequirementIDs, planGoalRequirementIDs(data, plan)) {
			out = append(out, decision)
			continue
		}
		if intersectsAny(decision.WorkItemIDs, plan.WorkItemIDs) {
			out = append(out, decision)
		}
	}
	return out
}

func planGoalRequirementIDs(data boardData, plan boardPlan) []string {
	goal := boardGoalByID(data, plan.GoalID)
	if goal == nil {
		return nil
	}
	ids := make([]string, 0)
	for _, req := range data.Requirements {
		if req.GoalID == goal.ID {
			ids = append(ids, req.ID)
		}
	}
	return ids
}

func intersectsAny(haystack []string, needles []string) bool {
	for _, h := range haystack {
		for _, n := range needles {
			if h == n {
				return true
			}
		}
	}
	return false
}

// milestoneWorkBullets returns a bullet list of the Work Item titles that
// belong to a milestone (matched by overlapping requirement IDs). Returns
// an empty string when no Work Item matches.
func milestoneWorkBullets(data boardData, plan boardPlan, milestone boardMilestone) string {
	var builder strings.Builder
	count := 0
	for _, workID := range plan.WorkItemIDs {
		work := boardWorkByID(data, workID)
		if work == nil || work.Status == "cancelled" {
			continue
		}
		if !intersects(work.RequirementIDs, milestone.RequirementIDs) {
			continue
		}
		count++
		fmt.Fprintf(&builder, "- %s\n", work.Title)
	}
	if count == 0 {
		return ""
	}
	return strings.TrimRight(builder.String(), "\n")
}


// forbiddenSectionRefPattern matches the tokens that must never appear as the
// primary text in a generated board document: bare §N / 第 N 节 / Phase N /
// Module N / Step N references. Per the user-communication-contract
// (禁止漏词表, v0.4.2), generated boards must restate the content in plain
// language or attach a clickable link, never show the raw number.
var forbiddenSectionRefPattern = regexp.MustCompile(`§\s*\d+|第\s*\d+\s*节|\bPhase\s*\d+|\bModule\s*\d+|\bStep\s*\d+`)

// lintBoardFile returns the unique forbidden section-ref tokens found in
// the given rendered Markdown body. HTML comments are stripped first so
// machine IDs can still live there for traceability without tripping the
// lint.
func lintBoardFile(body string) []string {
	stripped := stripHTMLCommentsForLint(body)
	matches := forbiddenSectionRefPattern.FindAllString(stripped, -1)
	seen := map[string]struct{}{}
	violations := []string{}
	for _, m := range matches {
		if _, ok := seen[m]; ok {
			continue
		}
		seen[m] = struct{}{}
		violations = append(violations, m)
	}
	return violations
}

func stripHTMLCommentsForLint(body string) string {
	var b strings.Builder
	b.Grow(len(body))
	for i := 0; i < len(body); {
		if i+3 < len(body) && body[i] == '<' && body[i+1] == '!' && body[i+2] == '-' && body[i+3] == '-' {
			end := strings.Index(body[i+4:], "-->")
			if end == -1 {
				break
			}
			i += 4 + end + 3
			continue
		}
		b.WriteByte(body[i])
		i++
	}
	return b.String()
}
