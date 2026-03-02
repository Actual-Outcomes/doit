package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type raiseFlagArgs struct {
	IssueID   string           `json:"issue_id"`
	Type      string           `json:"type"`
	Severity  int              `json:"severity"`
	Summary   string           `json:"summary"`
	Context   *json.RawMessage `json:"context,omitempty"`
	Project   *string          `json:"project,omitempty"`
	CreatedBy *string          `json:"created_by,omitempty"`
}

func (h *Handlers) RaiseFlag(ctx context.Context, _ *mcp.CallToolRequest, args raiseFlagArgs) (*mcp.CallToolResult, any, error) {
	input := store.RaiseFlagInput{
		IssueID:  args.IssueID,
		Type:     args.Type,
		Severity: args.Severity,
		Summary:  args.Summary,
	}

	if args.Context != nil {
		input.Context = *args.Context
	}
	if strSet(args.Project) {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		input.ProjectID = resolved
	}
	if args.CreatedBy != nil {
		input.CreatedBy = *args.CreatedBy
	}

	flag, err := h.store.RaiseFlag(ctx, input)
	if err != nil {
		return errResult(fmt.Errorf("raising flag: %w", err))
	}
	return jsonResult(flag)
}

type listFlagsArgs struct {
	Project  *string `json:"project,omitempty"`
	Status   *string `json:"status,omitempty"`
	Severity *int    `json:"severity,omitempty"`
	IssueID  *string `json:"issue_id,omitempty"`
	Limit    *int    `json:"limit,omitempty"`
}

func (h *Handlers) ListFlags(ctx context.Context, _ *mcp.CallToolRequest, args listFlagsArgs) (*mcp.CallToolResult, any, error) {
	filter := model.FlagFilter{}

	if strSet(args.Project) {
		resolved, err := resolveProjectSlug(ctx, h.store, *args.Project)
		if err != nil {
			return errResult(err)
		}
		filter.ProjectID = &resolved
	}
	if strSet(args.Status) {
		s := model.FlagStatus(*args.Status)
		filter.Status = &s
	}
	if args.Severity != nil {
		filter.Severity = args.Severity
	}
	if strSet(args.IssueID) {
		filter.IssueID = args.IssueID
	}
	if args.Limit != nil {
		filter.Limit = *args.Limit
	}

	flags, err := h.store.ListFlags(ctx, filter)
	if err != nil {
		return errResult(fmt.Errorf("listing flags: %w", err))
	}
	return jsonResult(flags)
}

type resolveFlagArgs struct {
	ID         string  `json:"id"`
	Resolution string  `json:"resolution"`
	ResolvedBy *string `json:"resolved_by,omitempty"`
}

func (h *Handlers) ResolveFlag(ctx context.Context, _ *mcp.CallToolRequest, args resolveFlagArgs) (*mcp.CallToolResult, any, error) {
	resolvedBy := ""
	if args.ResolvedBy != nil {
		resolvedBy = *args.ResolvedBy
	}

	flag, err := h.store.ResolveFlag(ctx, args.ID, args.Resolution, resolvedBy)
	if err != nil {
		return errResult(fmt.Errorf("resolving flag: %w", err))
	}
	return jsonResult(flag)
}
