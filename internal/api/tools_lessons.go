package api

import (
	"context"
	"fmt"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type recordLessonArgs struct {
	Title      string   `json:"title"`
	Mistake    string   `json:"mistake"`
	Correction string   `json:"correction"`
	Project    *string  `json:"project,omitempty"`
	IssueID    *string  `json:"issue_id,omitempty"`
	Expert     *string  `json:"expert,omitempty"`
	Components []string `json:"components,omitempty"`
	Severity   *int     `json:"severity,omitempty"`
	CreatedBy  *string  `json:"created_by,omitempty"`
}

func (h *Handlers) RecordLesson(ctx context.Context, _ *mcp.CallToolRequest, args recordLessonArgs) (*mcp.CallToolResult, any, error) {
	input := store.RecordLessonInput{
		Title:      args.Title,
		Mistake:    args.Mistake,
		Correction: args.Correction,
		Components: args.Components,
	}

	if strSet(args.Project) {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		input.ProjectID = resolved
	}
	if strSet(args.IssueID) {
		input.IssueID = *args.IssueID
	}
	if strSet(args.Expert) {
		input.Expert = *args.Expert
	}
	if args.Severity != nil {
		input.Severity = *args.Severity
	}
	if strSet(args.CreatedBy) {
		input.CreatedBy = *args.CreatedBy
	}

	lesson, err := h.store.RecordLesson(ctx, input)
	if err != nil {
		return errResult(fmt.Errorf("recording lesson: %w", err))
	}
	return jsonResult(lesson)
}

type listLessonsArgs struct {
	Project   *string `json:"project,omitempty"`
	Status    *string `json:"status,omitempty"`
	Expert    *string `json:"expert,omitempty"`
	Component *string `json:"component,omitempty"`
	Severity  *int    `json:"severity,omitempty"`
	Limit     *int    `json:"limit,omitempty"`
}

func (h *Handlers) ListLessons(ctx context.Context, _ *mcp.CallToolRequest, args listLessonsArgs) (*mcp.CallToolResult, any, error) {
	filter := model.LessonFilter{}

	if strSet(args.Project) {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		filter.ProjectID = &resolved
	}
	if strSet(args.Status) {
		s := model.LessonStatus(*args.Status)
		filter.Status = &s
	}
	if strSet(args.Expert) {
		filter.Expert = args.Expert
	}
	if strSet(args.Component) {
		filter.Component = args.Component
	}
	if args.Severity != nil {
		filter.Severity = args.Severity
	}
	if args.Limit != nil {
		filter.Limit = *args.Limit
	}

	lessons, err := h.store.ListLessons(ctx, filter)
	if err != nil {
		return errResult(fmt.Errorf("listing lessons: %w", err))
	}
	return jsonResult(lessons)
}

type resolveLessonArgs struct {
	ID         string  `json:"id"`
	ResolvedBy *string `json:"resolved_by,omitempty"`
}

func (h *Handlers) ResolveLesson(ctx context.Context, _ *mcp.CallToolRequest, args resolveLessonArgs) (*mcp.CallToolResult, any, error) {
	resolvedBy := ""
	if args.ResolvedBy != nil {
		resolvedBy = *args.ResolvedBy
	}

	lesson, err := h.store.ResolveLesson(ctx, args.ID, resolvedBy)
	if err != nil {
		return errResult(fmt.Errorf("resolving lesson: %w", err))
	}
	return jsonResult(lesson)
}
