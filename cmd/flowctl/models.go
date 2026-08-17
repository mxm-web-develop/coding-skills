package main

type WorkItem struct {
	SchemaVersion      int      `json:"schema_version"`
	ID                 string   `json:"id"`
	Revision           int      `json:"revision"`
	Kind               string   `json:"kind"`
	Title              string   `json:"title"`
	Status             string   `json:"status"`
	Priority           string   `json:"priority"`
	GoalID             *string  `json:"goal_id"`
	RequirementIDs     []string `json:"requirement_ids"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Scope              []string `json:"scope"`
	Owner              *string  `json:"owner"`
	RunID              *string  `json:"run_id"`
	EvidenceIDs        []string `json:"evidence_ids"`
	BlockedReason      *string  `json:"blocked_reason"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

type HarnessRun struct {
	SchemaVersion int        `json:"schema_version"`
	ID            string     `json:"id"`
	Revision      int        `json:"revision"`
	WorkItemID    string     `json:"work_item_id"`
	Owner         string     `json:"owner"`
	Status        string     `json:"status"`
	Phase         string     `json:"phase"`
	GitSHA        string     `json:"git_sha"`
	CheckpointIDs []string   `json:"checkpoint_ids"`
	EvidenceIDs   []string   `json:"evidence_ids"`
	Budgets       RunBudgets `json:"budgets"`
	StartedAt     string     `json:"started_at"`
	UpdatedAt     string     `json:"updated_at"`
	CompletedAt   *string    `json:"completed_at"`
}

type RunBudgets struct {
	MaxElapsedMinutes int `json:"max_elapsed_minutes,omitempty"`
	MaxRetries        int `json:"max_retries,omitempty"`
	MaxChangedFiles   int `json:"max_changed_files,omitempty"`
}

type Checkpoint struct {
	SchemaVersion  int      `json:"schema_version"`
	ID             string   `json:"id"`
	RunID          string   `json:"run_id"`
	WorkItemID     string   `json:"work_item_id"`
	Sequence       int      `json:"sequence"`
	Phase          string   `json:"phase"`
	Summary        string   `json:"summary"`
	NextAction     string   `json:"next_action"`
	GitSHA         string   `json:"git_sha"`
	CompletedSteps []string `json:"completed_steps"`
	ChangedFiles   []string `json:"changed_files"`
	OpenQuestions  []string `json:"open_questions"`
	CreatedAt      string   `json:"created_at"`
}

type Evidence struct {
	SchemaVersion int               `json:"schema_version"`
	ID            string            `json:"id"`
	WorkItemID    string            `json:"work_item_id"`
	RunID         string            `json:"run_id"`
	TestID        string            `json:"test_id"`
	Source        string            `json:"source"`
	Trust         string            `json:"trust"`
	Result        string            `json:"result"`
	Command       []string          `json:"command"`
	ExitCode      int               `json:"exit_code"`
	GitSHA        string            `json:"git_sha"`
	Environment   map[string]string `json:"environment"`
	StartedAt     string            `json:"started_at"`
	EndedAt       string            `json:"ended_at"`
	LogPath       string            `json:"log_path"`
	LogSHA256     string            `json:"log_sha256"`
	ExternalURI   *string           `json:"external_uri"`
	CreatedAt     string            `json:"created_at"`
}

type WorkLease struct {
	SchemaVersion int      `json:"schema_version"`
	WorkItemID    string   `json:"work_item_id"`
	RunID         string   `json:"run_id"`
	Owner         string   `json:"owner"`
	Scope         []string `json:"scope"`
	AcquiredAt    string   `json:"acquired_at"`
	ExpiresAt     string   `json:"expires_at"`
}

type eventRecord struct {
	SchemaVersion int            `json:"schema_version"`
	ID            string         `json:"id"`
	Type          string         `json:"type"`
	ObjectType    string         `json:"object_type"`
	ObjectID      string         `json:"object_id"`
	RunID         string         `json:"run_id,omitempty"`
	Revision      int            `json:"revision,omitempty"`
	Timestamp     string         `json:"timestamp"`
	Data          map[string]any `json:"data,omitempty"`
}
