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

// CompactLesson is a minimal representation of a Lesson for list responses.
type CompactLesson struct {
	ID       string       `json:"id"`
	Title    string       `json:"title"`
	Status   LessonStatus `json:"status"`
	Severity int          `json:"severity"`
	Expert   string       `json:"expert,omitempty"`
}

// ToCompact converts a Lesson to its compact form.
func (l *Lesson) ToCompact() CompactLesson {
	return CompactLesson{
		ID:       l.ID,
		Title:    l.Title,
		Status:   l.Status,
		Severity: l.Severity,
		Expert:   l.Expert,
	}
}

// ToCompactLessonList converts a slice of Lessons to CompactLessons.
func ToCompactLessonList(lessons []Lesson) []CompactLesson {
	out := make([]CompactLesson, len(lessons))
	for i := range lessons {
		out[i] = lessons[i].ToCompact()
	}
	return out
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
