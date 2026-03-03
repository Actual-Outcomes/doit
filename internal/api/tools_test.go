package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

func (m *mockStore) RaiseFlag(_ context.Context, _ store.RaiseFlagInput) (*model.Flag, error) {
	return &model.Flag{}, nil
}

func (m *mockStore) ListFlags(_ context.Context, _ model.FlagFilter) ([]model.Flag, error) {
	return nil, nil
}

func (m *mockStore) ResolveFlag(_ context.Context, _ string, _, _ string) (*model.Flag, error) {
	return &model.Flag{}, nil
}

func (m *mockStore) GenerateFlagID(_ context.Context) (string, error) {
	return "flg-test1", nil
}

func (m *mockStore) RecordRetry(_ context.Context, input store.RecordRetryInput) (*model.Retry, error) {
	return &model.Retry{ID: "rty-test1", IssueID: input.IssueID, Attempt: 1, Status: model.RetryStatus(input.Status)}, nil
}

func (m *mockStore) ListRetries(_ context.Context, _ string, _ model.RetryFilter) ([]model.Retry, error) {
	return nil, nil
}

func (m *mockStore) GenerateRetryID(_ context.Context) (string, error) {
	return "rty-test1", nil
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

func (m *mockStore) UpdateTenant(_ context.Context, tenantID string, name, slug *string) (*model.Tenant, error) {
	t := &model.Tenant{ID: uuid.MustParse(tenantID)}
	if name != nil {
		t.Name = *name
	}
	if slug != nil {
		t.Slug = *slug
	}
	return t, nil
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

func (m *mockStore) ListAllProjects(_ context.Context) ([]model.Project, error) {
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

func (m *mockStore) DeleteProject(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) DeleteTenant(_ context.Context, _ string) error {
	return nil
}

func (m *mockStore) GetConfig(_ context.Context, key string) (string, error) {
	return "", fmt.Errorf("config key %q not found", key)
}

func (m *mockStore) SetConfig(_ context.Context, _, _ string) error {
	return nil
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

	compact := true
	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{Compact: &compact})
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

	compact := false
	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{Compact: &compact})
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

	// Parse response envelope to verify count
	var resp listResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse response envelope: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("expected count=1, got %d", resp.Count)
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

	compact := true
	result, _, err := h.Ready(context.Background(), nil, readyArgs{Compact: &compact})
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

// --- Null string handling tests ---
// These verify that literal "null" strings (sent by MCP clients for JSON null)
// are treated as unset and don't cause FK violations or incorrect filters.

func TestUpdateIssue_NullStrings(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["x"] = &model.Issue{ID: "x", Title: "Original", Status: model.StatusOpen, Assignee: "dave"}

	// All optional string fields set to "null" — should be treated as no-op
	nullStr := "null"
	result, _, err := h.UpdateIssue(context.Background(), nil, updateIssueArgs{
		ID:          "x",
		Title:       &nullStr,
		Description: &nullStr,
		Status:      &nullStr,
		Assignee:    &nullStr,
		Owner:       &nullStr,
		Notes:       &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	// Title should remain unchanged because filterNull converts "null" to nil
	issue := ms.issues["x"]
	if issue.Title != "Original" {
		t.Errorf("title = %q, want %q (null should be no-op)", issue.Title, "Original")
	}
}

func TestListIssues_NullProject(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{ID: "a", Title: "Task A", Status: model.StatusOpen}

	// Project set to "null" — should be treated as unset, not trigger slug resolution
	nullStr := "null"
	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{
		Project: &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestReady_NullProject(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{ID: "a", Title: "Ready", Status: model.StatusOpen}

	nullStr := "null"
	result, _, err := h.Ready(context.Background(), nil, readyArgs{
		Project: &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestRecordLesson_NullOptionalFields(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	// All optional fields set to "null" — should not be passed to store
	nullStr := "null"
	result, _, err := h.RecordLesson(context.Background(), nil, recordLessonArgs{
		Title:      "Test lesson",
		Mistake:    "Did something wrong",
		Correction: "Do it right",
		Project:    &nullStr,
		IssueID:    &nullStr,
		Expert:     &nullStr,
		CreatedBy:  &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestRecordLesson_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.RecordLesson(context.Background(), nil, recordLessonArgs{
		Title:      "Lesson title",
		Mistake:    "The mistake",
		Correction: "The correction",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestListLessons_NullFilters(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	// All filters set to "null" — should return unfiltered results
	nullStr := "null"
	result, _, err := h.ListLessons(context.Background(), nil, listLessonsArgs{
		Project:   &nullStr,
		Status:    &nullStr,
		Expert:    &nullStr,
		Component: &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestListLessons_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.ListLessons(context.Background(), nil, listLessonsArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestResolveLesson_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.ResolveLesson(context.Background(), nil, resolveLessonArgs{
		ID: "lsn-test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestRaiseFlag_NullOptionalFields(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	nullStr := "null"
	result, _, err := h.RaiseFlag(context.Background(), nil, raiseFlagArgs{
		IssueID:   "doit-test",
		Type:      "structural_concern",
		Severity:  2,
		Summary:   "Test flag",
		Project:   &nullStr,
		CreatedBy: &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestRaiseFlag_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.RaiseFlag(context.Background(), nil, raiseFlagArgs{
		IssueID:  "doit-test",
		Type:     "red_flag",
		Severity: 1,
		Summary:  "Critical concern",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestListFlags_NullFilters(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	nullStr := "null"
	result, _, err := h.ListFlags(context.Background(), nil, listFlagsArgs{
		Project: &nullStr,
		Status:  &nullStr,
		IssueID: &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestListFlags_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.ListFlags(context.Background(), nil, listFlagsArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestResolveFlag_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.ResolveFlag(context.Background(), nil, resolveFlagArgs{
		ID:         "flg-test",
		Resolution: "Fixed the concern",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestCreateProject_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.CreateProject(context.Background(), nil, createProjectArgs{
		Name: "Test Project",
		Slug: "test-project",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !contains(text, "test-project") {
		t.Errorf("response should contain slug, got: %s", text)
	}
}

func TestListProjects_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.ListProjects(context.Background(), nil, listProjectsArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

// --- Retry tests ---

func TestRecordRetry_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.RecordRetry(context.Background(), nil, recordRetryArgs{
		IssueID: "doit-test",
		Status:  "failed",
		Error:   "timeout after 30s",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if !contains(text, "rty-test1") {
		t.Errorf("response should contain retry ID, got: %s", text)
	}
}

func TestRecordRetry_NullOptionalFields(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	nullStr := "null"
	result, _, err := h.RecordRetry(context.Background(), nil, recordRetryArgs{
		IssueID:   "doit-test",
		Status:    "failed",
		Error:     "connection refused",
		Project:   &nullStr,
		Agent:     &nullStr,
		CreatedBy: &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

func TestListRetries_HappyPath(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.ListRetries(context.Background(), nil, listRetriesArgs{
		IssueID: "doit-test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}
}

func TestListRetries_NullFilters(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	nullStr := "null"
	result, _, err := h.ListRetries(context.Background(), nil, listRetriesArgs{
		IssueID: "doit-test",
		Status:  &nullStr,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("expected success, got error: %s", result.Content[0].(*mcp.TextContent).Text)
	}
}

// --- strSet unit tests ---

func TestStrSet(t *testing.T) {
	tests := []struct {
		name string
		val  *string
		want bool
	}{
		{"nil", nil, false},
		{"empty", strPtr(""), false},
		{"null string", strPtr("null"), false},
		{"real value", strPtr("doit"), true},
		{"whitespace", strPtr(" "), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strSet(tt.val)
			if got != tt.want {
				t.Errorf("strSet(%v) = %v, want %v", tt.val, got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool   { return &b }

// --- Response protection tests ---

func TestListIssues_DefaultCompact(t *testing.T) {
	// When compact is not set (nil), default should be compact=true
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{
		ID:          "a",
		Title:       "Task A",
		Status:      model.StatusOpen,
		Description: "should not appear in compact",
	}

	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if contains(text, "should not appear in compact") {
		t.Error("default compact=true should exclude description")
	}
	if !contains(text, "Task A") {
		t.Error("compact response should include title")
	}
}

func TestListIssues_ResponseEnvelope(t *testing.T) {
	// Verify response contains count and items fields
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{ID: "a", Title: "Task A", Status: model.StatusOpen}
	ms.issues["b"] = &model.Issue{ID: "b", Title: "Task B", Status: model.StatusOpen}

	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var resp listResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse response envelope: %v", err)
	}
	if resp.Count != 2 {
		t.Errorf("count = %d, want 2", resp.Count)
	}
	if resp.HasMore {
		t.Error("has_more should be false for small result set")
	}
}

func TestListIssues_HardCapNoProject(t *testing.T) {
	// When compact=false and no project filter, hard cap at 20
	ms := newMockStore()
	h := NewHandlers(ms)
	// Add 25 issues
	for i := 0; i < 25; i++ {
		id := fmt.Sprintf("issue-%d", i)
		ms.issues[id] = &model.Issue{ID: id, Title: fmt.Sprintf("Task %d", i), Status: model.StatusOpen}
	}

	compact := false
	result, _, err := h.ListIssues(context.Background(), nil, listIssuesArgs{Compact: &compact})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var resp listResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Count > hardCapNoProject {
		t.Errorf("count = %d, should be capped at %d without project filter", resp.Count, hardCapNoProject)
	}
	if !resp.HasMore {
		t.Error("has_more should be true when results are truncated")
	}
}

func TestReady_DefaultCompact(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{
		ID:          "a",
		Title:       "Ready task",
		Status:      model.StatusOpen,
		Description: "hidden by default",
	}

	result, _, err := h.Ready(context.Background(), nil, readyArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	if contains(text, "hidden by default") {
		t.Error("default compact=true should exclude description from ready")
	}
}

func TestReady_ResponseEnvelope(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)
	ms.issues["a"] = &model.Issue{ID: "a", Title: "Ready", Status: model.StatusOpen}

	result, _, err := h.Ready(context.Background(), nil, readyArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var resp listResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse envelope: %v", err)
	}
	if resp.Count != 1 {
		t.Errorf("count = %d, want 1", resp.Count)
	}
}

func TestListLessons_DefaultCompact(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.ListLessons(context.Background(), nil, listLessonsArgs{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatal("expected success")
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var resp listResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse envelope: %v", err)
	}
	// Empty result should still have the envelope
	if resp.Count != 0 {
		t.Errorf("count = %d, want 0", resp.Count)
	}
}

func TestListComments_ResponseEnvelope(t *testing.T) {
	ms := newMockStore()
	h := NewHandlers(ms)

	result, _, err := h.ListComments(context.Background(), nil, listCommentsArgs{IssueID: "x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var resp listResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse envelope: %v", err)
	}
}

func TestProtectedListResult_AutoCompact(t *testing.T) {
	// Create a large result that exceeds maxResponseChars
	issues := make([]model.Issue, 0, 100)
	for i := 0; i < 100; i++ {
		issues = append(issues, model.Issue{
			ID:          fmt.Sprintf("issue-%d", i),
			Title:       fmt.Sprintf("Task %d with a moderately long title for size testing", i),
			Description: fmt.Sprintf("This is a very long description that takes up lots of space. %s", strings.Repeat("x", 500)),
			Status:      model.StatusOpen,
			IssueType:   model.TypeTask,
		})
	}

	result, _, err := protectedListResult(issues, len(issues), false, func() any {
		return model.ToCompactList(issues)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := result.Content[0].(*mcp.TextContent).Text
	var resp listResponse
	if err := json.Unmarshal([]byte(text), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.AutoCompacted {
		t.Error("should have auto-compacted the oversized response")
	}
	if resp.Message == "" {
		t.Error("should include a message about auto-compaction")
	}
	// Compacted response should not contain the long descriptions
	if contains(text, "very long description") {
		t.Error("auto-compacted response should not contain full descriptions")
	}
}

func TestApplyListDefaults(t *testing.T) {
	// Default limit
	if got := applyListDefaults(0, true, false); got != defaultLimit {
		t.Errorf("default limit = %d, want %d", got, defaultLimit)
	}
	// Hard cap when compact=false, no project
	if got := applyListDefaults(50, false, false); got != hardCapNoProject {
		t.Errorf("hard cap = %d, want %d", got, hardCapNoProject)
	}
	// No cap when project is set
	if got := applyListDefaults(50, false, true); got != 50 {
		t.Errorf("with project = %d, want 50", got)
	}
	// No cap when compact=true
	if got := applyListDefaults(50, true, false); got != 50 {
		t.Errorf("compact=true = %d, want 50", got)
	}
}

func TestCompactDefault(t *testing.T) {
	if !compactDefault(nil) {
		t.Error("nil should default to true")
	}
	if !compactDefault(boolPtr(true)) {
		t.Error("true should be true")
	}
	if compactDefault(boolPtr(false)) {
		t.Error("false should be false")
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
