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
	Project    *string  `json:"project"`
	IssueID    *string  `json:"issue_id"`
	Expert     *string  `json:"expert"`
	Components []string `json:"components"`
	Severity   *int     `json:"severity"`
	CreatedBy  *string  `json:"created_by"`
}

func (h *Handlers) RecordLesson(ctx context.Context, _ *mcp.CallToolRequest, args recordLessonArgs) (*mcp.CallToolResult, any, error) {
	input := store.RecordLessonInput{
		Title:      args.Title,
		Mistake:    args.Mistake,
		Correction: args.Correction,
		Components: args.Components,
	}

	if args.Project != nil && *args.Project != "" {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		input.ProjectID = resolved
	}
	if args.IssueID != nil {
		input.IssueID = *args.IssueID
	}
	if args.Expert != nil {
		input.Expert = *args.Expert
	}
	if args.Severity != nil {
		input.Severity = *args.Severity
	}
	if args.CreatedBy != nil {
		input.CreatedBy = *args.CreatedBy
	}

	lesson, err := h.store.RecordLesson(ctx, input)
	if err != nil {
		return errResult(fmt.Errorf("recording lesson: %w", err))
	}
	return jsonResult(lesson)
}

type listLessonsArgs struct {
	Project   *string `json:"project"`
	Status    *string `json:"status"`
	Expert    *string `json:"expert"`
	Component *string `json:"component"`
	Severity  *int    `json:"severity"`
	Limit     *int    `json:"limit"`
}

func (h *Handlers) ListLessons(ctx context.Context, _ *mcp.CallToolRequest, args listLessonsArgs) (*mcp.CallToolResult, any, error) {
	filter := model.LessonFilter{}

	if args.Project != nil && *args.Project != "" {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		filter.ProjectID = &resolved
	}
	if args.Status != nil && *args.Status != "" {
		s := model.LessonStatus(*args.Status)
		filter.Status = &s
	}
	if args.Expert != nil && *args.Expert != "" {
		filter.Expert = args.Expert
	}
	if args.Component != nil && *args.Component != "" {
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
	ResolvedBy *string `json:"resolved_by"`
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
