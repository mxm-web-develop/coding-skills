package main

type boardGoal struct {
	ID                 string   `json:"id"`
	Status             string   `json:"status"`
	Title              string   `json:"title"`
	Problem            string   `json:"problem"`
	Outcome            string   `json:"outcome"`
	InScope            []string `json:"in_scope"`
	NonGoals           []string `json:"non_goals"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Risks              []string `json:"risks"`
	TargetRelease      string   `json:"target_release"`
}

type boardRequirement struct {
	ID                 string   `json:"id"`
	GoalID             string   `json:"goal_id"`
	Status             string   `json:"status"`
	Statement          string   `json:"statement"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	TestIDs            []string `json:"test_ids"`
}

type boardMilestone struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Outcome        string   `json:"outcome"`
	RequirementIDs []string `json:"requirement_ids"`
	ExitGates      []string `json:"exit_gates"`
	TargetRelease  string   `json:"target_release"`
}

type boardPlan struct {
	ID          string           `json:"id"`
	GoalID      string           `json:"goal_id"`
	Status      string           `json:"status"`
	Title       string           `json:"title"`
	Milestones  []boardMilestone `json:"milestones"`
	WorkItemIDs []string         `json:"work_item_ids"`
}

type boardDecision struct {
	ID             string   `json:"id"`
	Status         string   `json:"status"`
	Title          string   `json:"title"`
	Decision       string   `json:"decision"`
	Consequences   []string `json:"consequences"`
	RequirementIDs []string `json:"requirement_ids"`
	WorkItemIDs    []string `json:"work_item_ids"`
}

type boardTestSpec struct {
	ID             string   `json:"id"`
	RequirementIDs []string `json:"requirement_ids"`
	WorkItemID     string   `json:"work_item_id"`
	Level          string   `json:"level"`
	Purpose        string   `json:"purpose"`
	Status         string   `json:"status"`
	EvidenceIDs    []string `json:"evidence_ids"`
}

type boardRelease struct {
	ID              string   `json:"id"`
	Status          string   `json:"status"`
	PreviousVersion string   `json:"previous_version"`
	Version         string   `json:"version"`
	WorkItemIDs     []string `json:"work_item_ids"`
	EvidenceIDs     []string `json:"evidence_ids"`
	Summary         string   `json:"summary"`
	KnownIssues     []string `json:"known_issues"`
	Migration       string   `json:"migration"`
	Rollback        string   `json:"rollback"`
	Tag             *string  `json:"tag"`
	UpdatedAt       string   `json:"updated_at"`
}

type boardTechnology struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type boardArchitecture struct {
	Style       string   `json:"style"`
	Constraints []string `json:"constraints"`
}

type boardVisualTesting struct {
	Required  bool     `json:"required"`
	Tool      string   `json:"tool"`
	Browsers  []string `json:"browsers"`
	Viewports []string `json:"viewports"`
}

type boardEngineeringProfile struct {
	Languages         []boardTechnology  `json:"languages"`
	Frameworks        []boardTechnology  `json:"frameworks"`
	Architecture      boardArchitecture  `json:"architecture"`
	SelectedPlaybooks []string           `json:"selected_playbooks"`
	VisualTesting     boardVisualTesting `json:"visual_testing"`
	Unknowns          []string           `json:"unknowns"`
}

type boardData struct {
	Status       projectStatus
	Goals        []boardGoal
	Requirements []boardRequirement
	Plans        []boardPlan
	WorkItems    []WorkItem
	Decisions    []boardDecision
	Tests        []boardTestSpec
	Evidence     []Evidence
	Releases     []boardRelease
	Engineering  *boardEngineeringProfile
}

type versionProgress struct {
	Version     string
	GoalTitle   string
	Total       int
	Done        int
	InProgress  int
	Review      int
	Blocked     int
	Pending     int
	PassedTests int
	FailedTests int
	OtherTests  int
}
