package store

import (
	"context"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/google/uuid"
)

// Store defines the persistence interface for doit.
type Store interface {
	// Issues
	CreateIssue(ctx context.Context, input CreateIssueInput) (*model.Issue, error)
	GetIssue(ctx context.Context, id string) (*model.Issue, error)
	UpdateIssue(ctx context.Context, id string, input UpdateIssueInput) (*model.Issue, error)
	ListIssues(ctx context.Context, filter model.IssueFilter) ([]model.Issue, error)
	DeleteIssue(ctx context.Context, id string) error

	// Ready detection
	ListReady(ctx context.Context, filter model.IssueFilter) ([]model.Issue, error)

	// Dependencies
	AddDependency(ctx context.Context, input AddDependencyInput) (*model.Dependency, error)
	RemoveDependency(ctx context.Context, issueID, dependsOnID string) error
	ListDependencies(ctx context.Context, issueID string, direction string) ([]model.Dependency, error)
	GetDependencyTree(ctx context.Context, rootID string, maxDepth int) ([]model.TreeNode, error)

	// Hierarchical IDs
	NextChildID(ctx context.Context, parentID string) (string, error)

	// Labels
	AddLabel(ctx context.Context, issueID, label string) error
	RemoveLabel(ctx context.Context, issueID, label string) error
	ListLabels(ctx context.Context, issueID string) ([]string, error)

	// Comments
	AddComment(ctx context.Context, issueID, author, text string) (*model.Comment, error)
	ListComments(ctx context.Context, issueID string) ([]model.Comment, error)

	// Events (audit trail)
	AddEvent(ctx context.Context, input AddEventInput) (*model.Event, error)
	ListEvents(ctx context.Context, issueID string, limit int) ([]model.Event, error)

	// Compaction
	SaveCompactionSnapshot(ctx context.Context, issueID string, level int, summary, original string) error
	GetCompactionSnapshots(ctx context.Context, issueID string) ([]model.CompactionSnapshot, error)

	// Aggregation
	CountIssuesByStatus(ctx context.Context) (map[string]int, error)

	// ID generation
	GenerateID(ctx context.Context, prefix string) (string, error)

	// Lessons
	RecordLesson(ctx context.Context, input RecordLessonInput) (*model.Lesson, error)
	ListLessons(ctx context.Context, filter model.LessonFilter) ([]model.Lesson, error)
	ResolveLesson(ctx context.Context, id string, resolvedBy string) (*model.Lesson, error)
	GenerateLessonID(ctx context.Context) (string, error)

	// Projects
	CreateProject(ctx context.Context, name, slug string) (*model.Project, error)
	GetProjectBySlug(ctx context.Context, slug string) (*model.Project, error)
	ListProjects(ctx context.Context) ([]model.Project, error)
	UpdateProject(ctx context.Context, projectID string, name, slug *string) (*model.Project, error)

	// Tenants
	CreateTenant(ctx context.Context, name, slug string) (*model.Tenant, error)
	ListTenants(ctx context.Context) ([]model.Tenant, error)
	ResolveAPIKey(ctx context.Context, keyHash string) (uuid.UUID, error)
	CreateAPIKey(ctx context.Context, tenantSlug, label, keyHash, prefix string) (*model.APIKeyInfo, error)
	RevokeAPIKey(ctx context.Context, prefix string) error
	ListAPIKeys(ctx context.Context, tenantSlug string) ([]model.APIKeyInfo, error)

	Close()
}

// CreateIssueInput holds the fields for creating a new issue.
type CreateIssueInput struct {
	ID                 string              // pre-generated ID (hash-based or child)
	Title              string
	Description        string
	Design             string
	AcceptanceCriteria string
	Notes              string
	Status             model.Status
	Priority           int
	IssueType          model.IssueType
	Assignee           string
	Owner              string
	CreatedBy          string
	ProjectID          string // project to assign the issue to
	ParentID           string // if set, creates parent-child dependency
	Labels             []string
	Ephemeral          bool
	MolType            model.MolType
	WorkType           model.WorkType
	WispType           model.WispType
}

// UpdateIssueInput holds optional fields for updating an issue.
// Nil pointer = no change; non-nil = set to this value.
type UpdateIssueInput struct {
	Title              *string
	Description        *string
	Design             *string
	AcceptanceCriteria *string
	Notes              *string
	Status             *model.Status
	Priority           *int
	Assignee           *string
	Owner              *string
	DueAt              *string // ISO 8601 or relative like "+6h"
	DeferUntil         *string
	CloseReason        *string
	Pinned             *bool
	ExternalRef        *string
}

// AddDependencyInput holds the fields for creating a dependency.
type AddDependencyInput struct {
	IssueID     string
	DependsOnID string
	Type        model.DependencyType
	CreatedBy   string
	ThreadID    string
}

// RecordLessonInput holds the fields for recording a lesson learned.
type RecordLessonInput struct {
	Title      string
	Mistake    string
	Correction string
	ProjectID  string
	IssueID    string
	Expert     string
	Components []string
	Severity   int
	CreatedBy  string
}

// AddEventInput holds the fields for creating an audit event.
type AddEventInput struct {
	IssueID   string
	EventType model.EventType
	Actor     string
	OldValue  string
	NewValue  string
	Comment   string
}
