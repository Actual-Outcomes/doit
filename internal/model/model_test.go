package model

import (
	"testing"
)

func TestToCompact(t *testing.T) {
	issue := Issue{
		ID:          "doit-abc",
		Title:       "Test issue",
		IssueType:   TypeTask,
		Status:      StatusOpen,
		Priority:    2,
		Assignee:    "agent",
		Owner:       "dave",
		Description: "should be excluded",
		Design:      "should be excluded",
		Notes:       "should be excluded",
		ProjectID:   "proj-123",
		Labels:      []string{"bug", "urgent"},
	}

	compact := issue.ToCompact()

	if compact.ID != issue.ID {
		t.Errorf("ID = %q, want %q", compact.ID, issue.ID)
	}
	if compact.Title != issue.Title {
		t.Errorf("Title = %q, want %q", compact.Title, issue.Title)
	}
	if compact.IssueType != issue.IssueType {
		t.Errorf("IssueType = %q, want %q", compact.IssueType, issue.IssueType)
	}
	if compact.Status != issue.Status {
		t.Errorf("Status = %q, want %q", compact.Status, issue.Status)
	}
	if compact.Priority != issue.Priority {
		t.Errorf("Priority = %d, want %d", compact.Priority, issue.Priority)
	}
	if compact.Assignee != issue.Assignee {
		t.Errorf("Assignee = %q, want %q", compact.Assignee, issue.Assignee)
	}
	if compact.Owner != issue.Owner {
		t.Errorf("Owner = %q, want %q", compact.Owner, issue.Owner)
	}
	if compact.ProjectID != issue.ProjectID {
		t.Errorf("ProjectID = %q, want %q", compact.ProjectID, issue.ProjectID)
	}
	if len(compact.Labels) != 2 || compact.Labels[0] != "bug" {
		t.Errorf("Labels = %v, want [bug urgent]", compact.Labels)
	}
}

func TestToCompactList(t *testing.T) {
	issues := []Issue{
		{ID: "a", Title: "First", Status: StatusOpen, IssueType: TypeTask},
		{ID: "b", Title: "Second", Status: StatusClosed, IssueType: TypeBug},
	}

	compacts := ToCompactList(issues)

	if len(compacts) != 2 {
		t.Fatalf("len = %d, want 2", len(compacts))
	}
	if compacts[0].ID != "a" || compacts[1].ID != "b" {
		t.Errorf("IDs = [%q, %q], want [a, b]", compacts[0].ID, compacts[1].ID)
	}
}

func TestToCompactListEmpty(t *testing.T) {
	compacts := ToCompactList([]Issue{})
	if len(compacts) != 0 {
		t.Fatalf("len = %d, want 0", len(compacts))
	}
}
