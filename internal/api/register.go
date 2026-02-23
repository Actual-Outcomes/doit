package api

import "github.com/modelcontextprotocol/go-sdk/mcp"

// RegisterTools registers all doit MCP tools on the server.
func RegisterTools(server *mcp.Server, h *Handlers) {
	// --- Issue CRUD ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_create_issue",
		Description: "Create a new work item (task, bug, feature, epic, etc). " +
			"Returns the created issue with its hash-based ID. " +
			"Use --parent to create a hierarchical child (e.g. epic.1).",
	}, h.CreateIssue)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_get_issue",
		Description: "Get full details of an issue including labels, dependencies, and parent. " +
			"Use when you need the complete picture of a work item.",
	}, h.GetIssue)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_update_issue",
		Description: "Update fields on an existing issue. Only specified fields are changed. " +
			"Use claim=true to atomically set assignee and status to in_progress.",
	}, h.UpdateIssue)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_issues",
		Description: "List issues with filtering by status, type, priority, assignee, and labels. " +
			"Supports sorting by priority, oldest, updated, or hybrid. " +
			"Use project slug to scope results to a single project.",
	}, h.ListIssues)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_delete_issue",
		Description: "Delete an issue. Cascades to dependencies, labels, comments, and events.",
	}, h.DeleteIssue)

	// --- Ready detection ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_ready",
		Description: "List issues ready for work — open, not blocked, not deferred. " +
			"Call this to find the next task to work on. " +
			"Use project slug to scope results to a single project.",
	}, h.Ready)

	// --- Dependencies ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_add_dependency",
		Description: "Add a dependency between two issues. " +
			"The 'blocks' type prevents the dependent issue from appearing in ready work.",
	}, h.AddDependency)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_remove_dependency",
		Description: "Remove a dependency between two issues.",
	}, h.RemoveDependency)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_dependencies",
		Description: "List dependencies for an issue. Direction: upstream (what it depends on), " +
			"downstream (what depends on it), or both.",
	}, h.ListDependencies)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_dependency_tree",
		Description: "Walk the parent-child hierarchy tree from a root issue. " +
			"Shows nested tasks at each depth level.",
	}, h.DependencyTree)

	// --- Comments ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_add_comment",
		Description: "Add a comment to an issue.",
	}, h.AddComment)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_comments",
		Description: "List comments on an issue, ordered by creation time.",
	}, h.ListComments)

	// --- Labels ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_add_label",
		Description: "Add a label to an issue.",
	}, h.AddLabel)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_remove_label",
		Description: "Remove a label from an issue.",
	}, h.RemoveLabel)

	// --- Messaging ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_send_message",
		Description: "Send a message to another agent. Messages are issue-type 'message' " +
			"with sender/assignee. Use thread_id to reply to an existing message.",
	}, h.SendMessage)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_messages",
		Description: "List messages, optionally filtered by recipient or unread status.",
	}, h.ListMessages)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_mark_message_read",
		Description: "Mark a message as read (sets status to closed).",
	}, h.MarkMessageRead)

	// --- Compaction ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_compact",
		Description: "Run semantic compaction on old closed issues. " +
			"Summarizes issues to save context window tokens. Default threshold: 7 days.",
	}, h.Compact)

	// --- Projects ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "doit_create_project",
		Description: "Create a project within your tenant for organizing issues.",
	}, h.CreateProject)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "doit_list_projects",
		Description: "List projects in your tenant.",
	}, h.ListProjects)

	// --- Tenant Management (admin only) ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_create_tenant",
		Description: "Create a new tenant. Requires admin API key. " +
			"Each tenant gets isolated data.",
	}, h.CreateTenant)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_tenants",
		Description: "List all tenants. Requires admin API key.",
	}, h.ListTenants)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_create_api_key",
		Description: "Generate a new API key for a tenant. Requires admin API key. " +
			"The raw key is returned once and cannot be retrieved again.",
	}, h.CreateAPIKey)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_revoke_api_key",
		Description: "Revoke an API key by its 8-character prefix. Requires admin API key.",
	}, h.RevokeAPIKey)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_api_keys",
		Description: "List all API keys for a tenant. Requires admin API key.",
	}, h.ListAPIKeys)
}
