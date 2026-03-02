package model

import "time"

// RetryStatus represents the outcome of a retry attempt.
type RetryStatus string

const (
	RetryFailed    RetryStatus = "failed"
	RetrySucceeded RetryStatus = "succeeded"
	RetryAbandoned RetryStatus = "abandoned"
	RetryEscalated RetryStatus = "escalated"
)

// Retry records one attempt at executing a task.
type Retry struct {
	ID        string      `json:"id"`
	TenantID  string      `json:"tenant_id,omitempty"`
	ProjectID string      `json:"project_id,omitempty"`
	IssueID   string      `json:"issue_id"`
	Attempt   int         `json:"attempt"`
	Status    RetryStatus `json:"status"`
	Error     string      `json:"error,omitempty"`
	Agent     string      `json:"agent,omitempty"`
	StartedAt time.Time   `json:"started_at"`
	EndedAt   *time.Time  `json:"ended_at,omitempty"`
	CreatedBy string      `json:"created_by,omitempty"`
}

// RetryFilter provides filtering for retry queries.
type RetryFilter struct {
	IssueID *string
	Status  *RetryStatus
	Limit   int
}
