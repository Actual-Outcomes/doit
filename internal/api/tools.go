package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// resolveProjectSlug takes a project identifier (slug or UUID) and returns
// the project UUID string. If the input looks like a UUID (36 chars with dashes),
// it is returned as-is; otherwise it is resolved as a slug via the store.
func resolveProjectSlug(ctx context.Context, s store.Store, input string) (string, error) {
	// UUID passthrough: 36 chars with dashes (e.g. "d0f96271-467b-41f0-9793-0f5150fc9a6d")
	if len(input) == 36 && input[8] == '-' && input[13] == '-' && input[18] == '-' && input[23] == '-' {
		return input, nil
	}
	p, err := s.GetProjectBySlug(ctx, input)
	if err != nil {
		return "", fmt.Errorf("resolving project %q: %w", input, err)
	}
	return p.ID.String(), nil
}

func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, _ := json.MarshalIndent(v, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func errResult(err error) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		IsError: true,
	}, nil, nil
}

// strSet returns true if the pointer holds a meaningful value —
// not nil, not empty, and not the literal string "null" that some
// MCP clients send in place of JSON null.
func strSet(s *string) bool {
	return s != nil && *s != "" && *s != "null"
}

// Response size protection constants.
const (
	maxResponseChars = 50_000
	defaultLimit     = 50
	hardCapNoProject = 20
)

// listResponse wraps list results with metadata for agent protection.
type listResponse struct {
	Count         int    `json:"count"`
	HasMore       bool   `json:"has_more,omitempty"`
	Items         any    `json:"items"`
	AutoCompacted bool   `json:"auto_compacted,omitempty"`
	Message       string `json:"message,omitempty"`
}

// protectedListResult wraps list items in a response envelope with size protection.
// If the serialized response exceeds maxResponseChars, it auto-compacts using compactFn.
// Pass compactFn=nil if there is no compact form for this type.
func protectedListResult(items any, count int, hasMore bool, compactFn func() any) (*mcp.CallToolResult, any, error) {
	resp := listResponse{
		Count:   count,
		HasMore: hasMore,
		Items:   items,
	}
	data, _ := json.MarshalIndent(resp, "", "  ")
	if len(data) > maxResponseChars && compactFn != nil {
		resp.Items = compactFn()
		resp.AutoCompacted = true
		resp.Message = "Response exceeded size limit; results returned in compact mode. Use the get endpoint for full details."
		data, _ = json.MarshalIndent(resp, "", "  ")
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

// compactDefault returns the effective compact setting, defaulting to true.
func compactDefault(compact *bool) bool {
	if compact == nil {
		return true
	}
	return *compact
}

// applyListDefaults resolves limit and enforces hard cap when compact=false without project.
func applyListDefaults(limit int, compact bool, hasProject bool) int {
	if limit == 0 {
		limit = defaultLimit
	}
	if !compact && !hasProject && limit > hardCapNoProject {
		limit = hardCapNoProject
	}
	return limit
}

type createIssueArgs struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Design             string   `json:"design"`
	AcceptanceCriteria string   `json:"acceptance_criteria"`
	Notes              string   `json:"notes"`
	Priority           int      `json:"priority"`
	IssueType          string   `json:"issue_type"`
	Assignee           string   `json:"assignee"`
	Owner              string   `json:"owner"`
	ParentID           string   `json:"parent_id"`
	Project            string   `json:"project"`
	Labels             []string `json:"labels"`
	Ephemeral          bool     `json:"ephemeral"`
}

func (h *Handlers) CreateIssue(ctx context.Context, _ *mcp.CallToolRequest, args createIssueArgs) (*mcp.CallToolResult, any, error) {
	var id string
	var err error
	if args.ParentID != "" {
		id, err = h.store.NextChildID(ctx, args.ParentID)
	} else {
		id, err = h.store.GenerateID(ctx, "")
	}
	if err != nil {
		return errResult(fmt.Errorf("generating ID: %w", err))
	}

	issueType := model.IssueType(args.IssueType)
	if issueType == "" {
		issueType = model.TypeTask
	}

	createdBy := args.Owner
	if createdBy == "" {
		createdBy = args.Assignee
	}
	if createdBy == "" {
		createdBy = "system"
	}

	var projectID string
	if args.Project != "" {
		projectID, err = resolveProjectSlug(ctx, h.store, args.Project)
		if err != nil {
			return errResult(err)
		}
	}

	issue, err := h.store.CreateIssue(ctx, store.CreateIssueInput{
		ID:                 id,
		Title:              args.Title,
		Description:        args.Description,
		Design:             args.Design,
		AcceptanceCriteria: args.AcceptanceCriteria,
		Notes:              args.Notes,
		Status:             model.StatusOpen,
		Priority:           args.Priority,
		IssueType:          issueType,
		Assignee:           args.Assignee,
		Owner:              args.Owner,
		CreatedBy:          createdBy,
		ProjectID:          projectID,
		ParentID:           args.ParentID,
		Labels:             args.Labels,
		Ephemeral:          args.Ephemeral,
	})
	if err != nil {
		return errResult(err)
	}

	return jsonResult(issue)
}

type getIssueArgs struct {
	ID string `json:"id"`
}

func (h *Handlers) GetIssue(ctx context.Context, _ *mcp.CallToolRequest, args getIssueArgs) (*mcp.CallToolResult, any, error) {
	issue, err := h.store.GetIssue(ctx, args.ID)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(issue)
}

type updateIssueArgs struct {
	ID          string  `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *int    `json:"priority"`
	Assignee    *string `json:"assignee"`
	Owner       *string `json:"owner"`
	Claim       bool    `json:"claim"`
	Pinned      *bool   `json:"pinned"`
	Notes       *string `json:"notes"`
}

func (h *Handlers) UpdateIssue(ctx context.Context, _ *mcp.CallToolRequest, args updateIssueArgs) (*mcp.CallToolResult, any, error) {
	// Filter out literal "null" strings that arrive from MCP client serialization.
	// go-sdk/mcp marks all struct fields as required, so clients send null for
	// fields they don't want to change. The JSON round-trip can turn *string null
	// into the literal string "null".
	filterNull := func(s *string) *string {
		if s != nil && *s == "null" {
			return nil
		}
		return s
	}

	input := store.UpdateIssueInput{
		Title:       filterNull(args.Title),
		Description: filterNull(args.Description),
		Priority:    args.Priority,
		Assignee:    filterNull(args.Assignee),
		Owner:       filterNull(args.Owner),
		Pinned:      args.Pinned,
		Notes:       filterNull(args.Notes),
	}

	if args.Status != nil && *args.Status != "null" {
		s := model.Status(*args.Status)
		input.Status = &s
	}

	if args.Claim {
		assignee := "agent"
		input.Assignee = &assignee
		s := model.StatusInProgress
		input.Status = &s
	}

	issue, err := h.store.UpdateIssue(ctx, args.ID, input)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(issue)
}

type listIssuesArgs struct {
	Status    string  `json:"status"`
	IssueType string  `json:"issue_type"`
	Priority  *int    `json:"priority"`
	Assignee  string  `json:"assignee"`
	Project   *string `json:"project"`
	Limit     int     `json:"limit"`
	SortBy    string  `json:"sort_by"`
	Compact   *bool   `json:"compact,omitempty"`
	Pinned    bool    `json:"pinned,omitempty"`
}

func (h *Handlers) ListIssues(ctx context.Context, _ *mcp.CallToolRequest, args listIssuesArgs) (*mcp.CallToolResult, any, error) {
	compact := compactDefault(args.Compact)
	hasProject := strSet(args.Project)
	limit := applyListDefaults(args.Limit, compact, hasProject)

	filter := model.IssueFilter{
		Limit:  limit + 1, // fetch one extra to detect truncation
		SortBy: args.SortBy,
	}
	if args.Status != "" {
		s := model.Status(args.Status)
		filter.Status = &s
	}
	if args.IssueType != "" {
		t := model.IssueType(args.IssueType)
		filter.IssueType = &t
	}
	if args.Priority != nil {
		filter.Priority = args.Priority
	}
	if args.Assignee != "" {
		filter.Assignee = &args.Assignee
	}
	if hasProject {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		filter.ProjectID = &resolved
	}
	if args.Pinned {
		t := true
		filter.Pinned = &t
	}

	issues, err := h.store.ListIssues(ctx, filter)
	if err != nil {
		return errResult(err)
	}

	hasMore := len(issues) > limit
	if hasMore {
		issues = issues[:limit]
	}

	if compact {
		compactItems := model.ToCompactList(issues)
		return protectedListResult(compactItems, len(compactItems), hasMore, nil)
	}
	return protectedListResult(issues, len(issues), hasMore, func() any {
		return model.ToCompactList(issues)
	})
}

type deleteIssueArgs struct {
	ID string `json:"id"`
}

func (h *Handlers) DeleteIssue(ctx context.Context, _ *mcp.CallToolRequest, args deleteIssueArgs) (*mcp.CallToolResult, any, error) {
	if err := h.store.DeleteIssue(ctx, args.ID); err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]string{"deleted": args.ID})
}

type readyArgs struct {
	Limit   int     `json:"limit"`
	Project *string `json:"project"`
	Compact *bool   `json:"compact,omitempty"`
}

func (h *Handlers) Ready(ctx context.Context, _ *mcp.CallToolRequest, args readyArgs) (*mcp.CallToolResult, any, error) {
	compact := compactDefault(args.Compact)
	hasProject := strSet(args.Project)
	limit := applyListDefaults(args.Limit, compact, hasProject)

	filter := model.IssueFilter{Limit: limit + 1}
	if hasProject {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		filter.ProjectID = &resolved
	}
	issues, err := h.store.ListReady(ctx, filter)
	if err != nil {
		return errResult(err)
	}

	hasMore := len(issues) > limit
	if hasMore {
		issues = issues[:limit]
	}

	if compact {
		compactItems := model.ToCompactList(issues)
		return protectedListResult(compactItems, len(compactItems), hasMore, nil)
	}
	return protectedListResult(issues, len(issues), hasMore, func() any {
		return model.ToCompactList(issues)
	})
}

type addDepArgs struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
}

func (h *Handlers) AddDependency(ctx context.Context, _ *mcp.CallToolRequest, args addDepArgs) (*mcp.CallToolResult, any, error) {
	depType := model.DependencyType(args.Type)
	if depType == "" {
		depType = model.DepBlocks
	}
	dep, err := h.store.AddDependency(ctx, store.AddDependencyInput{
		IssueID:     args.IssueID,
		DependsOnID: args.DependsOnID,
		Type:        depType,
	})
	if err != nil {
		return errResult(err)
	}
	return jsonResult(dep)
}

type removeDepArgs struct {
	IssueID     string `json:"issue_id"`
	DependsOnID string `json:"depends_on_id"`
}

func (h *Handlers) RemoveDependency(ctx context.Context, _ *mcp.CallToolRequest, args removeDepArgs) (*mcp.CallToolResult, any, error) {
	if err := h.store.RemoveDependency(ctx, args.IssueID, args.DependsOnID); err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]string{"removed": args.IssueID + " → " + args.DependsOnID})
}

type listDepsArgs struct {
	IssueID   string `json:"issue_id"`
	Direction string `json:"direction"`
}

func (h *Handlers) ListDependencies(ctx context.Context, _ *mcp.CallToolRequest, args listDepsArgs) (*mcp.CallToolResult, any, error) {
	deps, err := h.store.ListDependencies(ctx, args.IssueID, args.Direction)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(deps)
}

type depTreeArgs struct {
	RootID   string `json:"root_id"`
	MaxDepth int    `json:"max_depth"`
}

func (h *Handlers) DependencyTree(ctx context.Context, _ *mcp.CallToolRequest, args depTreeArgs) (*mcp.CallToolResult, any, error) {
	depth := args.MaxDepth
	if depth == 0 {
		depth = 3
	}
	nodes, err := h.store.GetDependencyTree(ctx, args.RootID, depth)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(nodes)
}

type addCommentArgs struct {
	IssueID string `json:"issue_id"`
	Author  string `json:"author"`
	Text    string `json:"text"`
}

func (h *Handlers) AddComment(ctx context.Context, _ *mcp.CallToolRequest, args addCommentArgs) (*mcp.CallToolResult, any, error) {
	comment, err := h.store.AddComment(ctx, args.IssueID, args.Author, args.Text)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(comment)
}

type listCommentsArgs struct {
	IssueID string `json:"issue_id"`
}

func (h *Handlers) ListComments(ctx context.Context, _ *mcp.CallToolRequest, args listCommentsArgs) (*mcp.CallToolResult, any, error) {
	comments, err := h.store.ListComments(ctx, args.IssueID)
	if err != nil {
		return errResult(err)
	}
	return protectedListResult(comments, len(comments), false, nil)
}

type labelArgs struct {
	IssueID string `json:"issue_id"`
	Label   string `json:"label"`
}

func (h *Handlers) AddLabel(ctx context.Context, _ *mcp.CallToolRequest, args labelArgs) (*mcp.CallToolResult, any, error) {
	if err := h.store.AddLabel(ctx, args.IssueID, args.Label); err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]string{"added": args.Label})
}

func (h *Handlers) RemoveLabel(ctx context.Context, _ *mcp.CallToolRequest, args labelArgs) (*mcp.CallToolResult, any, error) {
	if err := h.store.RemoveLabel(ctx, args.IssueID, args.Label); err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]string{"removed": args.Label})
}
