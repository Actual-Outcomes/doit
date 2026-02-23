package api

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createProjectArgs struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handlers) CreateProject(ctx context.Context, _ *mcp.CallToolRequest, args createProjectArgs) (*mcp.CallToolResult, any, error) {
	project, err := h.store.CreateProject(ctx, args.Name, args.Slug)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(project)
}

type listProjectsArgs struct{}

func (h *Handlers) ListProjects(ctx context.Context, _ *mcp.CallToolRequest, _ listProjectsArgs) (*mcp.CallToolResult, any, error) {
	projects, err := h.store.ListProjects(ctx)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(projects)
}
