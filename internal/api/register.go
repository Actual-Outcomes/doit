package api

import "github.com/modelcontextprotocol/go-sdk/mcp"

// RegisterAgentTools registers agent-facing MCP tools (25 tools).
func RegisterAgentTools(server *mcp.Server, h *Handlers) {
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
			"Use project slug to scope results to a single project. " +
			"Set compact=true for minimal responses that save context window tokens. " +
			"Set pinned=true to retrieve only pinned issues for fast orientation.",
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
			"Use project slug to scope results to a single project. " +
			"Set compact=true for minimal responses that save context window tokens.",
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

	// --- Lessons ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_record_lesson",
		Description: "Record a lesson learned — a mistake and its correction. " +
			"Used for continuous improvement. Tag with components and expert role.",
	}, h.RecordLesson)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_lessons",
		Description: "List lessons learned, filtered by project, status, expert, component, or severity. " +
			"Review before starting work to avoid repeating mistakes.",
	}, h.ListLessons)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_resolve_lesson",
		Description: "Mark a lesson as resolved after the correction has been applied.",
	}, h.ResolveLesson)

	// --- Retries ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_record_retry",
		Description: "Record a retry attempt for an issue. Attempt number is auto-computed. " +
			"Status: failed, succeeded, abandoned, escalated. " +
			"Use to track retry history for informed retry/escalate/abandon decisions.",
	}, h.RecordRetry)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_retries",
		Description: "List retry attempts for an issue, ordered by attempt number. " +
			"Shows attempt count, status, error, and agent for each attempt. " +
			"Use to check retry history before deciding to retry, escalate, or abandon.",
	}, h.ListRetries)

	// --- Flags (Human Escalation) ---

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_raise_flag",
		Description: "Raise a flag to escalate an issue for human decision. " +
			"Types: structural_concern, feature_concern, red_flag, human_decision, security_concern. " +
			"Severity: 1=critical, 2=blocking, 3=warning. " +
			"Issues with open severity 1-2 flags are excluded from doit_ready() output.",
	}, h.RaiseFlag)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_list_flags",
		Description: "List escalation flags. Filter by project, status (open/acknowledged/resolved), " +
			"severity, or issue_id. Call at session start to check for unresolved escalations.",
	}, h.ListFlags)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_resolve_flag",
		Description: "Resolve an escalation flag with a resolution message. " +
			"Records who resolved it and when.",
	}, h.ResolveFlag)

}

// RegisterAdminTools registers admin-only MCP tools (10 tools).
func RegisterAdminTools(server *mcp.Server, h *Handlers) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_create_tenant",
		Description: "Create a new tenant. Requires admin API key. " +
			"Each tenant gets isolated data.",
	}, h.CreateTenant)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_update_tenant",
		Description: "Update a tenant's name or slug. Requires admin API key. " +
			"Accepts tenant slug as identifier. Provide name and/or slug to change.",
	}, h.UpdateTenant)

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

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_update_project",
		Description: "Update a project's name or slug. Requires admin API key. " +
			"Accepts project ID or slug as identifier.",
	}, h.AdminUpdateProject)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_delete_project",
		Description: "Delete a project. Requires admin API key. " +
			"Rejects if issues still reference the project. Accepts project slug.",
	}, h.AdminDeleteProject)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_delete_tenant",
		Description: "Delete a tenant and its API keys. Requires admin API key. " +
			"Rejects if projects still exist (delete them first). Accepts tenant slug.",
	}, h.DeleteTenant)

	mcp.AddTool(server, &mcp.Tool{
		Name: "doit_rotate_admin_key",
		Description: "Generate a new admin API key and store its hash in the database. " +
			"The raw key is returned once. The env var key still works as fallback.",
	}, h.RotateAdminKey)
}
