package api

import (
	"context"
	"fmt"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type recordRetryArgs struct {
	IssueID   string  `json:"issue_id"`
	Status    string  `json:"status"`
	Error     string  `json:"error"`
	Project   *string `json:"project,omitempty"`
	Agent     *string `json:"agent,omitempty"`
	CreatedBy *string `json:"created_by,omitempty"`
}

func (h *Handlers) RecordRetry(ctx context.Context, _ *mcp.CallToolRequest, args recordRetryArgs) (*mcp.CallToolResult, any, error) {
	input := store.RecordRetryInput{
		IssueID: args.IssueID,
		Status:  args.Status,
		Error:   args.Error,
	}

	if strSet(args.Project) {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		input.ProjectID = resolved
	}
	if strSet(args.Agent) {
		input.Agent = *args.Agent
	}
	if strSet(args.CreatedBy) {
		input.CreatedBy = *args.CreatedBy
	}

	retry, err := h.store.RecordRetry(ctx, input)
	if err != nil {
		return errResult(fmt.Errorf("recording retry: %w", err))
	}
	return jsonResult(retry)
}

type listRetriesArgs struct {
	IssueID string  `json:"issue_id"`
	Status  *string `json:"status,omitempty"`
	Limit   *int    `json:"limit,omitempty"`
}

func (h *Handlers) ListRetries(ctx context.Context, _ *mcp.CallToolRequest, args listRetriesArgs) (*mcp.CallToolResult, any, error) {
	filter := model.RetryFilter{}

	if strSet(args.Status) {
		s := model.RetryStatus(*args.Status)
		filter.Status = &s
	}
	if args.Limit != nil {
		filter.Limit = *args.Limit
	}

	retries, err := h.store.ListRetries(ctx, args.IssueID, filter)
	if err != nil {
		return errResult(fmt.Errorf("listing retries: %w", err))
	}
	return jsonResult(retries)
}
