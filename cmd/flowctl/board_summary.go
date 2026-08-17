package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func activeBoardGoal(data boardData) *boardGoal {
	for index := range data.Goals {
		if data.Goals[index].ID == data.Status.ActiveGoal {
			return &data.Goals[index]
		}
	}
	for index := range data.Goals {
		if data.Goals[index].Status == "active" || data.Goals[index].Status == "accepted" {
			return &data.Goals[index]
		}
	}
	return nil
}

func boardGoalByID(data boardData, id string) *boardGoal {
	for index := range data.Goals {
		if data.Goals[index].ID == id {
			return &data.Goals[index]
		}
	}
	return nil
}

func boardWorkByID(data boardData, id string) *WorkItem {
	for index := range data.WorkItems {
		if data.WorkItems[index].ID == id {
			return &data.WorkItems[index]
		}
	}
	return nil
}

func boardEvidenceByID(data boardData, id string) *Evidence {
	for index := range data.Evidence {
		if data.Evidence[index].ID == id {
			return &data.Evidence[index]
		}
	}
	return nil
}

func versionForWork(data boardData, work WorkItem) string {
	for _, plan := range data.Plans {
		if !contains(plan.WorkItemIDs, work.ID) {
			continue
		}
		for _, milestone := range plan.Milestones {
			if milestone.TargetRelease != "" && intersects(milestone.RequirementIDs, work.RequirementIDs) {
				return milestone.TargetRelease
			}
		}
		if goal := boardGoalByID(data, plan.GoalID); goal != nil && goal.TargetRelease != "" {
			return goal.TargetRelease
		}
	}
	if work.GoalID != nil {
		if goal := boardGoalByID(data, *work.GoalID); goal != nil && goal.TargetRelease != "" {
			return goal.TargetRelease
		}
	}
	return firstNonEmpty(data.Status.CurrentVersion, "未规划")
}

func versionProgressRows(data boardData) []versionProgress {
	rows := map[string]*versionProgress{}
	ensure := func(version string) *versionProgress {
		version = firstNonEmpty(version, "未规划")
		if rows[version] == nil {
			rows[version] = &versionProgress{Version: version}
		}
		return rows[version]
	}
	for _, goal := range data.Goals {
		if goal.Status == "superseded" || goal.Status == "archived" || goal.Status == "rejected" {
			continue
		}
		if goal.TargetRelease != "" {
			ensure(goal.TargetRelease).GoalTitle = goal.Title
		}
	}
	for _, work := range data.WorkItems {
		if work.Status == "cancelled" {
			continue
		}
		row := ensure(versionForWork(data, work))
		if row.GoalTitle == "" && work.GoalID != nil {
			if goal := boardGoalByID(data, *work.GoalID); goal != nil {
				row.GoalTitle = goal.Title
			}
		}
		row.Total++
		switch work.Status {
		case "done":
			row.Done++
		case "in_progress":
			row.InProgress++
		case "ready_for_review":
			row.Review++
		case "blocked":
			row.Blocked++
		default:
			row.Pending++
		}
		passed, failed, other := evidenceCountsForWork(data, work.ID)
		row.PassedTests += passed
		row.FailedTests += failed
		row.OtherTests += other
	}
	result := make([]versionProgress, 0, len(rows))
	for _, row := range rows {
		result = append(result, *row)
	}
	sort.Slice(result, func(i, j int) bool {
		return compareVersions(result[i].Version, result[j].Version) > 0
	})
	return result
}

func evidenceCountsForWork(data boardData, workID string) (passed, failed, other int) {
	for _, evidence := range data.Evidence {
		if evidence.WorkItemID != workID {
			continue
		}
		switch evidence.Result {
		case "passed":
			passed++
		case "failed":
			failed++
		default:
			other++
		}
	}
	for _, test := range data.Tests {
		if test.WorkItemID != workID || boardTestHasEvidence(data, test) {
			continue
		}
		other++
	}
	return passed, failed, other
}

func boardTestHasEvidence(data boardData, test boardTestSpec) bool {
	for _, id := range test.EvidenceIDs {
		if boardEvidenceByID(data, id) != nil {
			return true
		}
	}
	for _, evidence := range data.Evidence {
		if evidence.TestID == test.ID {
			return true
		}
	}
	return false
}

func evidenceCountsForIDs(data boardData, ids []string) (passed, failed, other int) {
	for _, id := range ids {
		evidence := boardEvidenceByID(data, id)
		if evidence == nil {
			other++
			continue
		}
		switch evidence.Result {
		case "passed":
			passed++
		case "failed":
			failed++
		default:
			other++
		}
	}
	return passed, failed, other
}

func testEvidenceSummary(data boardData, test boardTestSpec) string {
	passed, failed, other := evidenceCountsForIDs(data, test.EvidenceIDs)
	if len(test.EvidenceIDs) == 0 {
		for _, evidence := range data.Evidence {
			if evidence.TestID != test.ID {
				continue
			}
			switch evidence.Result {
			case "passed":
				passed++
			case "failed":
				failed++
			default:
				other++
			}
		}
	}
	if passed+failed+other == 0 {
		return "尚无执行证据"
	}
	return fmt.Sprintf("通过 %d / 失败 %d / 待确认 %d", passed, failed, other)
}

func workTestSummary(data boardData, workID string) string {
	passed, failed, other := evidenceCountsForWork(data, workID)
	if passed+failed+other == 0 {
		count := 0
		for _, test := range data.Tests {
			if test.WorkItemID == workID {
				count++
			}
		}
		if count == 0 {
			return "未定义"
		}
		return fmt.Sprintf("待执行 %d", count)
	}
	return fmt.Sprintf("通过 %d / 失败 %d / 待确认 %d", passed, failed, other)
}

func overallSummary(data boardData) string {
	goal := activeBoardGoal(data)
	goalText := "尚未确认新的产品目标"
	if goal != nil {
		goalText = fmt.Sprintf("正在推进“%s”", goal.Title)
	}
	current := currentMajorProgress(data)
	active := current.InProgress + current.Review
	return fmt.Sprintf(
		"项目当前处于 **%s** 大版本，当前开发版本是 **%s**，%s。本轮任务中 **%d 个已完成、%d 个正在处理、%d 个阻塞**；测试结果为 **%d 项通过、%d 项失败、%d 项待确认**。下一步：**%s**。",
		majorVersion(data.Status.CurrentVersion), data.Status.CurrentVersion, goalText,
		current.Done, active, current.Blocked,
		current.PassedTests, current.FailedTests, current.OtherTests,
		humanAction(data.Status.NextAction),
	)
}

func currentMajorProgress(data boardData) versionProgress {
	current := versionProgress{Version: majorVersion(data.Status.CurrentVersion)}
	for _, row := range versionProgressRows(data) {
		if !sameMajorVersion(row.Version, data.Status.CurrentVersion) {
			continue
		}
		current.Total += row.Total
		current.Done += row.Done
		current.InProgress += row.InProgress
		current.Review += row.Review
		current.Blocked += row.Blocked
		current.Pending += row.Pending
		current.PassedTests += row.PassedTests
		current.FailedTests += row.FailedTests
		current.OtherTests += row.OtherTests
	}
	return current
}

func majorVersion(version string) string {
	value := strings.TrimPrefix(strings.TrimSpace(version), "v")
	if value == "" || value == "unknown" {
		return "未知"
	}
	parts := strings.Split(value, ".")
	if len(parts) == 0 || parts[0] == "" {
		return "未知"
	}
	if _, err := strconv.Atoi(parts[0]); err != nil {
		return version
	}
	return "V" + parts[0]
}

func compareVersions(left, right string) int {
	leftParts, leftOK := versionNumbers(left)
	rightParts, rightOK := versionNumbers(right)
	if leftOK && rightOK {
		for index := 0; index < 3; index++ {
			if leftParts[index] != rightParts[index] {
				return leftParts[index] - rightParts[index]
			}
		}
		return 0
	}
	return strings.Compare(left, right)
}

func versionNumbers(version string) ([3]int, bool) {
	result := [3]int{}
	value := strings.TrimPrefix(strings.TrimSpace(version), "v")
	value = strings.SplitN(value, "-", 2)[0]
	value = strings.SplitN(value, "+", 2)[0]
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return result, false
	}
	for index, part := range parts {
		number, err := strconv.Atoi(part)
		if err != nil {
			return result, false
		}
		result[index] = number
	}
	return result, true
}

func sameMajorVersion(left, right string) bool {
	leftParts, leftOK := versionNumbers(left)
	rightParts, rightOK := versionNumbers(right)
	return leftOK && rightOK && leftParts[0] == rightParts[0]
}

func humanStatus(status string) string {
	labels := map[string]string{
		"draft": "草稿", "active": "进行中", "accepted": "已确认", "superseded": "已替代",
		"archived": "已归档", "rejected": "已拒绝", "cancelled": "已取消", "ready": "待开始",
		"in_progress": "开发中", "blocked": "阻塞", "ready_for_review": "等待检查", "done": "已完成",
		"planned": "已规划", "released": "已发布", "withdrawn": "已撤回", "red": "未通过",
		"green": "已通过", "retired": "已停用", "passed": "通过", "failed": "失败", "unverified": "待确认",
		"proposed": "等待确认", "approved": "已确认", "executing": "处理中", "applied": "已完成",
		"partial": "部分完成", "rolled-back": "已恢复",
	}
	if label := labels[status]; label != "" {
		return label
	}
	return "待确认"
}

func humanPhase(phase string) string {
	labels := map[string]string{
		"intake": "需求接收", "baselining": "项目基线确认", "document_audit": "历史文档盘点",
		"workspace_audit": "检查工作区内容", "workspace_cleanup_planning": "确认整理范围", "workspace_cleaning": "整理项目工作区",
		"engineering_profiling": "技术栈识别", "goal_alignment": "目标对齐", "planning": "开发规划",
		"researching": "技术调研", "test_specification": "测试设计", "implementing": "开发实现",
		"verifying": "测试验证", "reviewing": "代码检查", "integrating": "整理提交",
		"releasing": "版本发布", "syncing": "状态同步", "completed": "已完成",
	}
	if label := labels[phase]; label != "" {
		return label
	}
	return "待确认"
}

func humanKind(kind string) string {
	labels := map[string]string{"feature": "功能", "bug": "缺陷", "refactor": "重构", "test": "测试", "docs": "文档", "research": "调研", "chore": "维护"}
	return firstNonEmpty(labels[kind], "其他")
}

func humanPriority(priority string) string {
	labels := map[string]string{"critical": "紧急", "high": "高", "medium": "中", "low": "低"}
	return firstNonEmpty(labels[priority], "未设置")
}

func humanOwner(owner string) string {
	labels := map[string]string{
		"codex-agent":  "Codex",
		"cursor-agent": "Cursor",
		"claude-agent": "Claude Code",
	}
	return firstNonEmpty(labels[owner], owner, "未分配")
}

func humanEngineeringGuidance(playbooks []string) string {
	if len(playbooks) == 0 {
		return "尚未确认"
	}
	return "已根据当前语言、框架和项目结构采用匹配的开发与测试规范"
}

func humanArchitectureStyle(style string) string {
	labels := map[string]string{
		"feature modules":    "按产品功能划分模块",
		"layered":            "按职责分层组织",
		"hexagonal":          "核心业务与外部系统分离",
		"monolith":           "单体应用",
		"modular monolith":   "模块化单体应用",
		"microservices":      "多个独立服务协作",
		"service-oriented":   "按服务职责拆分",
		"clean architecture": "核心业务与技术实现分层",
	}
	return firstNonEmpty(labels[strings.ToLower(strings.TrimSpace(style))], style, "尚未确认")
}

func humanTestLevel(level string) string {
	labels := map[string]string{
		"unit": "单元", "component": "模块", "integration": "模块联动", "contract": "接口兼容",
		"end_to_end": "完整流程", "accessibility": "无障碍体验", "visual": "界面效果", "migration": "数据迁移",
		"performance": "性能", "security": "安全", "smoke": "基本功能",
	}
	return firstNonEmpty(labels[level], "其他")
}

func humanAction(action string) string {
	if action == "" || action == "not_recorded" {
		return "等待安排下一项工作"
	}
	labels := map[string]string{
		"adopt_existing_project":       "完成既有项目基线和历史文档盘点",
		"discover_product_goal":        "讨论并确认首个产品目标",
		"profile_project_engineering":  "识别项目技术环境并确认开发与测试方式",
		"plan_product_delivery":        "把需求安排成版本、开发阶段和具体任务",
		"research_and_design_solution": "比较可行方案并确认实现方式",
		"specify_tests":                "确认测试范围和通过条件",
		"implement_work_item":          "继续当前开发任务",
		"diagnose_and_verify":          "定位问题并完成回归测试",
		"review_change":                "检查本次代码改动",
		"integrate_git_change":         "整理并提交已完成的改动",
		"manage_release":               "整理版本内容并准备发布",
		"sync_project_knowledge":       "更新项目进度和版本记录",
		"clean_project_workspace":      "核对并整理项目工作区",
		"review_visual_evidence":       "检查功能测试和界面效果",
	}
	if label := labels[action]; label != "" {
		return label
	}
	return "查看项目记录并确认下一步"
}

func mdCell(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return firstNonEmpty(strings.TrimSpace(value), "—")
}

func joinBoardText(values []string) string {
	if len(values) == 0 {
		return "—"
	}
	return strings.Join(values, "；")
}

func intersects(left, right []string) bool {
	for _, candidate := range left {
		if contains(right, candidate) {
			return true
		}
	}
	return false
}
