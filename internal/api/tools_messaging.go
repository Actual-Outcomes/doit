package api

import (
	"context"
	"fmt"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type sendMessageArgs struct {
	Body      string `json:"body"`
	To        string `json:"to"`
	ThreadID  string `json:"thread_id"`
	Ephemeral bool   `json:"ephemeral"`
}

func (h *Handlers) SendMessage(ctx context.Context, _ *mcp.CallToolRequest, args sendMessageArgs) (*mcp.CallToolResult, any, error) {
	id, err := h.store.GenerateID(ctx, "msg")
	if err != nil {
		return errResult(fmt.Errorf("generating ID: %w", err))
	}

	title := args.Body
	if len(title) > 80 {
		title = title[:77] + "..."
	}

	issue, err := h.store.CreateIssue(ctx, store.CreateIssueInput{
		ID:          id,
		Title:       title,
		Description: args.Body,
		Status:      model.StatusOpen,
		Priority:    2,
		IssueType:   model.TypeMessage,
		Assignee:    args.To,
		Ephemeral:   args.Ephemeral,
	})
	if err != nil {
		return errResult(err)
	}

	if args.ThreadID != "" {
		_, err = h.store.AddDependency(ctx, store.AddDependencyInput{
			IssueID:     issue.ID,
			DependsOnID: args.ThreadID,
			Type:        model.DepRepliesTo,
			ThreadID:    args.ThreadID,
		})
		if err != nil {
			return errResult(fmt.Errorf("threading: %w", err))
		}
	}

	return jsonResult(issue)
}

type listMessagesArgs struct {
	To     string `json:"to"`
	Unread bool   `json:"unread"`
	Limit  int    `json:"limit"`
}

func (h *Handlers) ListMessages(ctx context.Context, _ *mcp.CallToolRequest, args listMessagesArgs) (*mcp.CallToolResult, any, error) {
	msgType := model.TypeMessage
	filter := model.IssueFilter{
		IssueType: &msgType,
		Limit:     args.Limit,
		SortBy:    "updated",
	}
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if args.To != "" {
		filter.Assignee = &args.To
	}
	if args.Unread {
		s := model.StatusOpen
		filter.Status = &s
	}

	issues, err := h.store.ListIssues(ctx, filter)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(issues)
}

type markReadArgs struct {
	ID string `json:"id"`
}

func (h *Handlers) MarkMessageRead(ctx context.Context, _ *mcp.CallToolRequest, args markReadArgs) (*mcp.CallToolResult, any, error) {
	s := model.StatusClosed
	issue, err := h.store.UpdateIssue(ctx, args.ID, store.UpdateIssueInput{Status: &s})
	if err != nil {
		return errResult(err)
	}
	return jsonResult(issue)
}
