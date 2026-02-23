package model

import (
	"time"

	"github.com/google/uuid"
)

// LessonStatus represents the workflow state of a lesson.
type LessonStatus string

const (
	LessonOpen     LessonStatus = "open"
	LessonResolved LessonStatus = "resolved"
)

// Lesson records a mistake and its correction for continuous improvement.
type Lesson struct {
	ID         string       `json:"id"`
	TenantID   uuid.UUID    `json:"tenant_id"`
	ProjectID  string       `json:"project_id,omitempty"`
	IssueID    string       `json:"issue_id,omitempty"`
	Title      string       `json:"title"`
	Mistake    string       `json:"mistake"`
	Correction string       `json:"correction"`
	Expert     string       `json:"expert,omitempty"`
	Components []string     `json:"components"`
	Severity   int          `json:"severity"`
	Status     LessonStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	CreatedBy  string       `json:"created_by,omitempty"`
	ResolvedAt *time.Time   `json:"resolved_at,omitempty"`
	ResolvedBy string       `json:"resolved_by,omitempty"`
}

// LessonFilter provides filtering for lesson queries.
type LessonFilter struct {
	ProjectID *string
	Status    *LessonStatus
	Expert    *string
	Component *string // filter by component (ANY match)
	Severity  *int
	Limit     int
}
