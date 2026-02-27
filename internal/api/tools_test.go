package api

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// mockStore implements store.Store for handler tests.
type mockStore struct {
	issues  map[string]*model.Issue
	deps    []model.Dependency
	labels  map[string][]string
	project *model.Project
}

func newMockStore() *mockStore {
	return &mockStore{
		issues: make(map[string]*model.Issue),
		labels: make(map[string][]string),
	}
}

func (m *mockStore) CreateIssue(_ context.Context, input store.CreateIssueInput) (*model.Issue, error) {
	issue := &model.Issue{
		ID:        input.ID,
		Title:     input.Title,
		Status:    input.Status,
		Priority:  input.Priority,
		IssueType: input.IssueType,
		Assignee:  input.Assignee,
		Owner:     input.Owner,
		ProjectID: input.ProjectID,
	}
	m.issues[input.ID] = issue
	return issue, nil
}

func (m *mockStore) GetIssue(_ context.Context, id string) (*model.Issue, error) {
	issue, ok := m.issues[id]
	if !ok {
		return nil, fmt.Errorf("issue %q not found", id)
	}
	return issue, nil
}

func (m *mockStore) UpdateIssue(_ context.Context, id string, input store.UpdateIssueInput) (*model.Issue, error) {
	issue, ok := m.issues[id]
	if !ok {
		return nil, fmt.Errorf("issue %q not found", id)
	}
	if input.Title != nil {
		issue.Title = *input.Title
	}
	if input.Status != nil {
		issue.Status = *input.Status
	}
	if input.Assignee != nil {
		issue.Assignee = *input.Assignee
	}
	if input.Pinned != nil {
		issue.Pinned = *input.Pinned
	}
	return issue, nil
}

func (m *mockStore) ListIssues(_ context.Context, filter model.IssueFilter) ([]model.Issue, error) {
	var out []model.Issue
	for _, issue := range m.issues {
		if filter.Status != nil && issue.Status != *filter.Status {
			continue
		}
		if filter.Pinned != nil && issue.Pinned != *filter.Pinned {
			continue
		}
		out = append(out, *issue)
	}
	return out, nil
}

func (m *mockStore) DeleteIssue(_ context.Context, id string) error {
	if _, ok := m.issues[id]; !ok {
		return fmt.Errorf("issue %q not found", id)
	}
	delete(m.issues, id)
	return nil
}

func (m *mockStore) ListReady(_ context.Context, _ model.IssueFilter) ([]model.Issue, error) {
	var out []model.Issue
	for _, issue := range m.issues {
		if issue.Status == model.StatusOpen {
			out = append(out, *issue)
		}
	}
	return out, nil
}

func (m *mockStore) AddDependency(_ context.Context, input store.AddDependencyInput) (*model.Dependency, error) {
	dep := &model.Dependency{
		IssueID:     input.IssueID,
		DependsOnID: input.DependsOnID,
		Type:        input.Type,
	}
	m.deps = append(m.deps, *dep)
	return dep, nil
}

func (m *mockStore) RemoveDependency(_ context.Context, _, _ string) error { return nil }

func (m *mockStore) ListDependencies(_ context.Context, issueID string, _ string) ([]model.Dependency, error) {
	var out []model.Dependency
	for _, d := range m.deps {
		if d.IssueID == issueID || d.DependsOnID == issueID {
			out = append(out, d)
		}
	}
	return out, nil
}

func (m *mockStore) GetDependencyTree(_ context.Context, _ string, _ int) ([]model.TreeNode, error) {
	return nil, nil
}

func (m *mockStore) NextChildID(_ context.Context, parentID string) (string, error) {
	return parentID + ".1", nil
}

func (m *mockStore) AddLabel(_ context.Context, issueID, label string) error {
	m.labels[issueID] = append(m.labels[issueID], label)
	return nil
}

func (m *mockStore) RemoveLabel(_ context.Context, issueID, label string) error {
	labels := m.labels[issueID]
	for i, l := range labels {
		if l == label {
			m.labels[issueID] = append(labels[:i], labels[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockStore) ListLabels(_ context.Context, issueID string) ([]string, error) {
	return m.labels[issueID], nil
}

func (m *mockStore) AddComment(_ context.Context, issueID, author, text string) (*model.Comment, error) {
	return &model.Comment{ID: 1, IssueID: issueID, Author: author, Text: text}, nil
}

func (m *mockStore) ListComments(_ context.Context, _ string) ([]model.Comment, error) {
	return nil, nil
}

func (m *mockStore) AddEvent(_ context.Context, _ store.AddEventInput) (*model.Event, error) {
	return &model.Event{}, nil
}

func (m *mockStore) ListEvents(_ context.Context, _ string, _ int) ([]model.Event, error) {
	return nil, nil
}

func (m *mockStore) SaveCompactionSnapshot(_ context.Context, _ string, _ int, _, _ string) error {
	return nil
}

func (m *mockStore) GetCompactionSnapshots(_ context.Context, _ string) ([]model.CompactionSnapshot, error) {
	return nil, nil
}

func (m *mockStore) CountIssuesByStatus(_ context.Context) (map[string]int, error) {
	return nil, nil
}

func (m *mockStore) GenerateID(_ context.Context, _ string) (string, error) {
	return "doit-test1", nil
}

func (m *mockStore) RecordLesson(_ context.Context, _ store.RecordLessonInput) (*model.Lesson, error) {
	return &model.Lesson{}, nil
}

func (m *mockStore) ListLessons(_ context.Context, _ model.LessonFilter) ([]model.Lesson, error) {
	return nil, nil
}

func (m *mockStore) ResolveLesson(_ context.Context, _ string, _ string) (*model.Lesson, error) {
	return &model.Lesson{}, nil
}

func (m *mockStore) GenerateLessonID(_ context.Context) (string, error) {
	return "lesson-test1", nil
}

func (m *mockStore) CreateProject(_ context.Context, name, slug string) (*model.Project, error) {
	p := &model.Project{ID: uuid.New(), Name: name, Slug: slug}
	m.project = p
	return p, nil
}

func (m *mockStore) GetProjectBySlug(_ context.Context, slug string) (*model.Project, error) {
	if m.project != nil && m.project.Slug == slug {
		return m.project, nil
	}
	return nil, fmt.Errorf("project %q not found", slug)
}

func (m *mockStore) ListProjects(_ context.Context) ([]model.Project, error) {
	if m.project != nil {
		return []model.Project{*m.project}, nil
	}
	return nil, nil
}

func (m *mockStore) CreateTenant(_ context.Context, name, slug string) (*model.Tenant, error) {
	return &model.Tenant{ID: uuid.New(), Name: name, Slug: slug}, nil
}

func (m *mockStore) ListTenants(_ context.Context) ([]model.Tenant, error) { return nil, nil }

func (m *mockStore) ResolveAPIKey(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, fmt.Errorf("not found")
}

func (m *mockStore) CreateAPIKey(_ context.Context, _, _, _, _ string) (*model.APIKeyInfo, error) {
	return &model.APIKeyInfo{}, nil
}

func (m *mockStore) RevokeAPIKey(_ context.Context, _ string) error { return nil }

func (m *mockStore) ListAPIKeys(_ context.Context, _ string) ([]model.APIKeyInfo, error) {
	return nil, nil
}

func (m *mockStore) UpdateProject(_ context.Context, projectID string, name, slug *string) (*model.Project, error) {
	p := &model.Project{ID: uuid.MustParse(projectID)}
	if name != nil {
		p.Name = *name
	}
	if slug != nil {
		p.Slug = *slug
	}
	return p, nil
}

func (m *mockStore) Close() {}

// --- Tests ---

func TestCreateIssue(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.CreateIssue(context.Background(), nil, createIssueArgs{
		Title:     "Test task",
		IssueType: "task",
		Priority:  2,
		Owner:     "dave",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	// Verify the issue was stored
	if len(ms.issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(ms.issues))
	}
}

func TestGetIssue(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["doit-abc"] = &model.Issue{ID: "doit-abc", Title: "Found it", Status: model.StatusOpen}

	result, _, err := h.GetIssue(context.Background(), nil, getIssueArgs{ID: "doit-abc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !contains(text, "Found it") {
		t.Errorf("response should contain issue title, got: %s", text)
	}
}

func TestGetIssue_NotFound(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.GetIssue(context.Background(), nil, getIssueArgs{ID: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for missing issue")
	}
}

func TestListIssues_Compact(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{
		ID:          "a",
		Title:       "Task A",
		Status:      model.StatusOpen,
		IssueType:   model.TypeTask,
		Description: "full description here",
	}

	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{Compact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	// Compact should NOT include description
	if contains(text, "full description here") {
		t.Error("compact response should not include description")
	}
	// But should include title
	if !contains(text, "Task A") {
		t.Error("compact response should include title")
	}
}

func TestListIssues_Full(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{
		ID:          "a",
		Title:       "Task A",
		Status:      model.StatusOpen,
		Description: "full description here",
	}

	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !contains(text, "full description here") {
		t.Error("full response should include description")
	}
}

func TestListIssues_PinnedFilter(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{ID: "a", Title: "Pinned", Status: model.StatusOpen, Pinned: true}
	ms.issues["b"] = &model.Issue{ID: "b", Title: "Not pinned", Status: model.StatusOpen, Pinned: false}

	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{Pinned: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !contains(text, "Pinned") {
		t.Error("should include pinned issue")
	}

	// Parse response to verify only one issue returned
	var issues []model.Issue
	if err := json.Unmarshal([]byte(text), &issues); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("expected 1 pinned issue, got %d", len(issues))
	}
}

func TestReady_Compact(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{
		ID:          "a",
		Title:       "Ready task",
		Status:      model.StatusOpen,
		Description: "should not appear",
	}

	result, _, err := h.Ready(context.Background(), nil, readyArgs{Compact: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if contains(text, "should not appear") {
		t.Error("compact ready should not include description")
	}
	if !contains(text, "Ready task") {
		t.Error("compact ready should include title")
	}
}

func TestDeleteIssue(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["doit-del"] = &model.Issue{ID: "doit-del", Title: "Delete me"}

	result, _, err := h.DeleteIssue(context.Background(), nil, deleteIssueArgs{ID: "doit-del"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	if len(ms.issues) != 0 {
		t.Error("issue should have been deleted from store")
	}
}

func TestDeleteIssue_NotFound(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.DeleteIssue(context.Background(), nil, deleteIssueArgs{ID: "nonexistent"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for missing issue")
	}
}

func TestAddDependency(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.AddDependency(context.Background(), nil, addDepArgs{
		IssueID:     "a",
		DependsOnID: "b",
		Type:        "blocks",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	if len(ms.deps) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(ms.deps))
	}
}

func TestAddDependency_DefaultType(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	_, _, err := h.AddDependency(context.Background(), nil, addDepArgs{
		IssueID:     "a",
		DependsOnID: "b",
		Type:        "", // should default to "blocks"
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ms.deps[0].Type != model.DepBlocks {
		t.Errorf("type = %q, want %q", ms.deps[0].Type, model.DepBlocks)
	}
}

func TestAddComment(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.AddComment(context.Background(), nil, addCommentArgs{
		IssueID: "a",
		Author:  "dave",
		Text:    "looks good",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestAddLabel(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.AddLabel(context.Background(), nil, labelArgs{
		IssueID: "a",
		Label:   "urgent",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
	if len(ms.labels["a"]) != 1 || ms.labels["a"][0] != "urgent" {
		t.Errorf("labels = %v, want [urgent]", ms.labels["a"])
	}
}

func TestUpdateIssue_Claim(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["x"] = &model.Issue{ID: "x", Title: "Claimable", Status: model.StatusOpen}

	result, _, err := h.UpdateIssue(context.Background(), nil, updateIssueArgs{
		ID:    "x",
		Claim: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}

	issue := ms.issues["x"]
	if issue.Status != model.StatusInProgress {
		t.Errorf("status = %q, want in_progress", issue.Status)
	}
	if issue.Assignee != "agent" {
		t.Errorf("assignee = %q, want agent", issue.Assignee)
	}
}

// helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
