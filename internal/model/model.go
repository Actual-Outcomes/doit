package model

import (
	"encoding/json"
	"time"
)

// Status represents the workflow state of an issue.
type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusDeferred   Status = "deferred"
	StatusClosed     Status = "closed"
	StatusPinned     Status = "pinned"
	StatusHooked     Status = "hooked"
)

// IssueType discriminates the kind of issue.
type IssueType string

const (
	TypeBug      IssueType = "bug"
	TypeFeature  IssueType = "feature"
	TypeTask     IssueType = "task"
	TypeEpic     IssueType = "epic"
	TypeChore    IssueType = "chore"
	TypeDecision IssueType = "decision"
	TypeMessage  IssueType = "message"
	TypeMolecule IssueType = "molecule"
	TypeEvent    IssueType = "event"
)

// DependencyType represents the kind of relationship between two issues.
type DependencyType string

const (
	DepBlocks            DependencyType = "blocks"
	DepConditionalBlocks DependencyType = "conditional-blocks"
	DepWaitsFor          DependencyType = "waits-for"
	DepParentChild       DependencyType = "parent-child"
	DepRelated           DependencyType = "related"
	DepRelatesTo         DependencyType = "relates-to"
	DepDiscoveredFrom    DependencyType = "discovered-from"
	DepCausedBy          DependencyType = "caused-by"
	DepRepliesTo         DependencyType = "replies-to"
	DepDuplicates        DependencyType = "duplicates"
	DepSupersedes        DependencyType = "supersedes"
	DepAuthoredBy        DependencyType = "authored-by"
	DepAssignedTo        DependencyType = "assigned-to"
	DepApprovedBy        DependencyType = "approved-by"
	DepAttests           DependencyType = "attests"
	DepValidates         DependencyType = "validates"
	DepTracks            DependencyType = "tracks"
	DepUntil             DependencyType = "until"
	DepDelegatedFrom     DependencyType = "delegated-from"
)

// MolType is the molecule execution strategy.
type MolType string

const (
	MolSwarm  MolType = "swarm"
	MolPatrol MolType = "patrol"
	MolWork   MolType = "work"
)

// WorkType controls how molecule children compete for work.
type WorkType string

const (
	WorkMutex           WorkType = "mutex"
	WorkOpenCompetition WorkType = "open_competition"
)

// WispType categorizes ephemeral child issues used by molecules.
type WispType string

const (
	WispHeartbeat  WispType = "heartbeat"
	WispPing       WispType = "ping"
	WispPatrol     WispType = "patrol"
	WispGCReport   WispType = "gc_report"
	WispRecovery   WispType = "recovery"
	WispError      WispType = "error"
	WispEscalation WispType = "escalation"
)

// AgentState tracks the lifecycle of an agent-type issue.
type AgentState string

const (
	AgentIdle     AgentState = "idle"
	AgentSpawning AgentState = "spawning"
	AgentRunning  AgentState = "running"
	AgentWorking  AgentState = "working"
	AgentStuck    AgentState = "stuck"
	AgentDone     AgentState = "done"
	AgentStopped  AgentState = "stopped"
	AgentDead     AgentState = "dead"
)

// EventType categorizes audit trail entries.
type EventType string

const (
	EventCreated           EventType = "created"
	EventUpdated           EventType = "updated"
	EventStatusChanged     EventType = "status_changed"
	EventCommented         EventType = "commented"
	EventClosed            EventType = "closed"
	EventReopened          EventType = "reopened"
	EventDependencyAdded   EventType = "dependency_added"
	EventDependencyRemoved EventType = "dependency_removed"
	EventLabelAdded        EventType = "label_added"
	EventLabelRemoved      EventType = "label_removed"
	EventCompacted         EventType = "compacted"
)

// Issue is the universal work item. Every task, bug, epic, message, molecule,
// and agent-bead is an Issue with different IssueType and field usage.
type Issue struct {
	// Identity
	ID          string `json:"id" db:"id"`
	ContentHash string `json:"content_hash,omitempty" db:"content_hash"`

	// Core fields
	Title              string    `json:"title" db:"title"`
	Description        string    `json:"description,omitempty" db:"description"`
	Design             string    `json:"design,omitempty" db:"design"`
	AcceptanceCriteria string    `json:"acceptance_criteria,omitempty" db:"acceptance_criteria"`
	Notes              string    `json:"notes,omitempty" db:"notes"`
	SpecID             string    `json:"spec_id,omitempty" db:"spec_id"`
	Status             Status    `json:"status" db:"status"`
	Priority           int       `json:"priority" db:"priority"`
	IssueType          IssueType `json:"issue_type" db:"issue_type"`

	// Assignment
	Assignee string `json:"assignee,omitempty" db:"assignee"`
	Owner    string `json:"owner,omitempty" db:"owner"`

	// Time estimates
	EstimatedMinutes *int `json:"estimated_minutes,omitempty" db:"estimated_minutes"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	CreatedBy string     `json:"created_by,omitempty" db:"created_by"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	ClosedAt  *time.Time `json:"closed_at,omitempty" db:"closed_at"`
	DueAt     *time.Time `json:"due_at,omitempty" db:"due_at"`
	DeferUntil *time.Time `json:"defer_until,omitempty" db:"defer_until"`

	// Closure
	CloseReason     string `json:"close_reason,omitempty" db:"close_reason"`
	ClosedBySession string `json:"closed_by_session,omitempty" db:"closed_by_session"`

	// External references
	ExternalRef  *string `json:"external_ref,omitempty" db:"external_ref"`
	SourceSystem string  `json:"source_system,omitempty" db:"source_system"`
	SourceRepo   string  `json:"source_repo,omitempty" db:"source_repo"`

	// Metadata
	Metadata json.RawMessage `json:"metadata,omitempty" db:"metadata"`

	// Compaction (memory decay)
	CompactionLevel   int        `json:"compaction_level" db:"compaction_level"`
	CompactedAt       *time.Time `json:"compacted_at,omitempty" db:"compacted_at"`
	CompactedAtCommit *string    `json:"compacted_at_commit,omitempty" db:"compacted_at_commit"`
	OriginalSize      int        `json:"original_size,omitempty" db:"original_size"`

	// Messaging
	Sender    string `json:"sender,omitempty" db:"sender"`
	Ephemeral bool   `json:"ephemeral,omitempty" db:"ephemeral"`

	// Molecule fields
	MolType      MolType  `json:"mol_type,omitempty" db:"mol_type"`
	WorkType     WorkType `json:"work_type,omitempty" db:"work_type"`
	Crystallizes bool     `json:"crystallizes,omitempty" db:"crystallizes"`

	// Wisp fields
	WispType WispType `json:"wisp_type,omitempty" db:"wisp_type"`

	// Display/template
	Pinned     bool `json:"pinned,omitempty" db:"pinned"`
	IsTemplate bool `json:"is_template,omitempty" db:"is_template"`

	// Quality
	QualityScore *float32 `json:"quality_score,omitempty" db:"quality_score"`

	// Event fields
	EventKind string `json:"event_kind,omitempty" db:"event_kind"`
	Actor     string `json:"actor,omitempty" db:"actor"`
	Target    string `json:"target,omitempty" db:"target"`
	Payload   string `json:"payload,omitempty" db:"payload"`

	// Await/gate fields
	AwaitType string        `json:"await_type,omitempty" db:"await_type"`
	AwaitID   string        `json:"await_id,omitempty" db:"await_id"`
	Timeout   time.Duration `json:"timeout,omitempty" db:"timeout_ns"`

	// Agent fields
	AgentState   AgentState `json:"agent_state,omitempty" db:"agent_state"`
	LastActivity *time.Time `json:"last_activity,omitempty" db:"last_activity"`
	RoleType     string     `json:"role_type,omitempty" db:"role_type"`
	Rig          string     `json:"rig,omitempty" db:"rig"`
	HookBead     string     `json:"hook_bead,omitempty" db:"hook_bead"`
	RoleBead     string     `json:"role_bead,omitempty" db:"role_bead"`

	// Denormalized fields (populated by queries, not stored directly)
	Labels       []string     `json:"labels,omitempty" db:"-"`
	Dependencies []Dependency `json:"dependencies,omitempty" db:"-"`
	ParentID     string       `json:"parent_id,omitempty" db:"-"`
}

// Dependency represents a directed relationship between two issues.
type Dependency struct {
	IssueID     string         `json:"issue_id" db:"issue_id"`
	DependsOnID string         `json:"depends_on_id" db:"depends_on_id"`
	Type        DependencyType `json:"type" db:"type"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	CreatedBy   string         `json:"created_by,omitempty" db:"created_by"`
	Metadata    json.RawMessage `json:"metadata,omitempty" db:"metadata"`
	ThreadID    string         `json:"thread_id,omitempty" db:"thread_id"`
}

// Comment is a discussion entry on an issue.
type Comment struct {
	ID        int64     `json:"id" db:"id"`
	IssueID   string    `json:"issue_id" db:"issue_id"`
	Author    string    `json:"author" db:"author"`
	Text      string    `json:"text" db:"text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Event is an audit trail entry for an issue.
type Event struct {
	ID        int64     `json:"id" db:"id"`
	IssueID   string    `json:"issue_id" db:"issue_id"`
	EventType EventType `json:"event_type" db:"event_type"`
	Actor     string    `json:"actor" db:"actor"`
	OldValue  string    `json:"old_value,omitempty" db:"old_value"`
	NewValue  string    `json:"new_value,omitempty" db:"new_value"`
	Comment   string    `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ChildCounter tracks the last child number for hierarchical IDs.
type ChildCounter struct {
	ParentID  string `json:"parent_id" db:"parent_id"`
	LastChild int    `json:"last_child" db:"last_child"`
}

// TreeNode represents an issue in a dependency tree traversal.
type TreeNode struct {
	Issue     Issue  `json:"issue"`
	Depth     int    `json:"depth"`
	ParentID  string `json:"parent_id,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
}

// CompactionSnapshot preserves original issue content before memory decay.
type CompactionSnapshot struct {
	ID        int64     `json:"id" db:"id"`
	IssueID   string    `json:"issue_id" db:"issue_id"`
	Level     int       `json:"level" db:"level"`
	Summary   string    `json:"summary" db:"summary"`
	Original  string    `json:"original" db:"original"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// IssueFilter provides comprehensive filtering for issue queries.
type IssueFilter struct {
	Status       *Status    `json:"status,omitempty"`
	StatusNot    []Status   `json:"status_not,omitempty"`
	Priority     *int       `json:"priority,omitempty"`
	IssueType    *IssueType `json:"issue_type,omitempty"`
	IssueTypeNot []IssueType `json:"issue_type_not,omitempty"`
	Assignee     *string    `json:"assignee,omitempty"`
	Owner        *string    `json:"owner,omitempty"`
	ParentID     *string    `json:"parent_id,omitempty"`
	Labels       []string   `json:"labels,omitempty"`       // AND match
	LabelsAny    []string   `json:"labels_any,omitempty"`   // OR match
	Search       *string    `json:"search,omitempty"`        // title/description search
	CreatedAfter *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter *time.Time `json:"updated_after,omitempty"`
	Ephemeral    *bool      `json:"ephemeral,omitempty"`
	Pinned       *bool      `json:"pinned,omitempty"`
	Overdue      *bool      `json:"overdue,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
	SortBy       string     `json:"sort_by,omitempty"` // "priority", "oldest", "updated", "hybrid"
}
