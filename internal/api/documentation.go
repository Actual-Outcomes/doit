package api

import (
	"net/http"
	"strings"

	"github.com/Actual-Outcomes/doit/internal/version"
)

// DocumentationHandler serves a self-contained HTML page with project documentation.
func DocumentationHandler() http.HandlerFunc {
	html := strings.ReplaceAll(documentationHTML, "{{VERSION}}", version.Number)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

const documentationHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Doit — AI Agent Work Planner & Tracker</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
      line-height: 1.6;
      color: #1a1a2e;
      background: #f8f9fa;
      padding: 2rem;
      max-width: 900px;
      margin: 0 auto;
    }
    h1 { font-size: 2rem; margin-bottom: 0.5rem; color: #16213e; }
    h2 { font-size: 1.5rem; margin-top: 2rem; margin-bottom: 0.75rem; color: #16213e; border-bottom: 2px solid #e2e8f0; padding-bottom: 0.3rem; }
    h3 { font-size: 1.15rem; margin-top: 1.5rem; margin-bottom: 0.5rem; color: #334155; }
    p { margin-bottom: 0.75rem; }
    code {
      background: #e2e8f0;
      padding: 0.15em 0.4em;
      border-radius: 3px;
      font-size: 0.9em;
      font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', monospace;
    }
    pre {
      background: #1e293b;
      color: #e2e8f0;
      padding: 1rem;
      border-radius: 8px;
      overflow-x: auto;
      margin-bottom: 1rem;
    }
    pre code { background: none; color: inherit; padding: 0; }
    ul, ol { margin-bottom: 0.75rem; padding-left: 1.5rem; }
    li { margin-bottom: 0.3rem; }
    strong { color: #0f172a; }
    table { width: 100%; border-collapse: collapse; margin-bottom: 1rem; }
    th, td { text-align: left; padding: 0.5rem 0.75rem; border-bottom: 1px solid #e2e8f0; }
    th { background: #f1f5f9; font-weight: 600; }
    hr { border: none; border-top: 1px solid #e2e8f0; margin: 2rem 0; }
    .subtitle { color: #64748b; margin-bottom: 2rem; font-size: 1.1rem; }
    .badge {
      display: inline-block;
      padding: 0.15em 0.5em;
      border-radius: 4px;
      font-size: 0.8em;
      font-weight: 600;
    }
    .badge-open { background: #dbeafe; color: #1e40af; }
    .badge-closed { background: #d1fae5; color: #065f46; }
    .badge-blocked { background: #fce7f3; color: #9d174d; }
    .badge-deferred { background: #fef3c7; color: #92400e; }
    .badge-task { background: #e2e8f0; color: #475569; }
    .badge-bug { background: #fee2e2; color: #991b1b; }
    .badge-feature { background: #ede9fe; color: #6d28d9; }
    .badge-epic { background: #dbeafe; color: #1e40af; }
  </style>
</head>
<body>

<h1>Doit — AI Agent Work Planner & Tracker</h1>
<p class="subtitle">Plan, track, and coordinate work across AI agent sessions &middot; v{{VERSION}}</p>

<h2>What is Doit?</h2>
<p>Doit is an <strong>MCP server</strong> that gives AI coding assistants persistent work tracking. It manages <strong>issues</strong> (tasks, bugs, features, epics), <strong>dependencies</strong>, and <strong>comments</strong>. For agent-to-agent messaging, use <a href="https://herald.aoendpoint.com/documentation">TheHerald</a>.</p>
<p>Think of it as a lightweight issue tracker purpose-built for AI agents — with hash-based IDs, hierarchical tasks, dependency-aware "ready" detection, and semantic compaction for memory decay.</p>

<h2>Quick Start</h2>

<h3>1. Configure MCP Connection</h3>
<p>Add to your project's <code>.mcp.json</code>. Agents use <code>/mcp</code>; admins use <code>/admin/mcp</code>:</p>
<pre><code>{
  "mcpServers": {
    "doit": {
      "type": "http",
      "url": "https://din.aoendpoint.com/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_TENANT_API_KEY"
      }
    },
    "doit-admin": {
      "type": "http",
      "url": "https://din.aoendpoint.com/admin/mcp",
      "headers": {
        "Authorization": "Bearer YOUR_ADMIN_API_KEY"
      }
    }
  }
}</code></pre>

<h3>2. Add Rules to CLAUDE.md</h3>
<p>Add the following to your project's <code>CLAUDE.md</code>:</p>
<pre><code>## Doit — Work Tracking

This project uses Doit for persistent work tracking via MCP.

### Workflow
- Call doit_list_projects to find your project slug
- Call doit_ready with project slug to find available work
- Call doit_get_issue for full details before starting
- Call doit_update_issue with claim=true to start work
- Call doit_update_issue with status=closed when done
- Call doit_create_issue with project slug for new work items
- Call doit_add_dependency to track blockers</code></pre>

<h2>Agent Tools (25)</h2>
<p>Available on <code>POST /mcp</code> — authenticated with any API key (tenant or admin).</p>

<h3>Issue CRUD</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_create_issue</code></td><td>Create a new work item (task, bug, feature, epic, etc). Returns the created issue with its hash-based ID. Use <code>parent_id</code> to create a hierarchical child (e.g. epic.1). Use <code>project</code> (slug) to assign to a project.</td></tr>
  <tr><td><code>doit_get_issue</code></td><td>Get full details of an issue including labels, dependencies, and parent.</td></tr>
  <tr><td><code>doit_update_issue</code></td><td>Update fields on an existing issue. Only specified fields are changed. Use claim=true to atomically set assignee and status to in_progress.</td></tr>
  <tr><td><code>doit_list_issues</code></td><td>List issues with filtering by status, type, priority, assignee, and labels. Supports sorting by priority, oldest, updated, or hybrid. Use <code>project</code> (slug) to scope results. Set <code>pinned=true</code> to retrieve only pinned issues. Returns <code>{count, has_more, items}</code> envelope. Defaults: <code>compact=true</code>, <code>limit=50</code>. Without project filter and <code>compact=false</code>, hard cap at 20 items. Oversized responses auto-compact.</td></tr>
  <tr><td><code>doit_delete_issue</code></td><td>Delete an issue. Cascades to dependencies, labels, comments, and events.</td></tr>
</table>

<h3>Ready Detection</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_ready</code></td><td>List issues ready for work — open, not blocked, not deferred. Call this to find the next task to work on. Use <code>project</code> (slug) to scope results. Returns <code>{count, has_more, items}</code> envelope. Defaults: <code>compact=true</code>, <code>limit=50</code>. Without project filter and <code>compact=false</code>, hard cap at 20 items.</td></tr>
</table>

<h3>Dependencies</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_add_dependency</code></td><td>Add a dependency between two issues. The 'blocks' type prevents the dependent from appearing in ready work.</td></tr>
  <tr><td><code>doit_remove_dependency</code></td><td>Remove a dependency between two issues.</td></tr>
  <tr><td><code>doit_list_dependencies</code></td><td>List dependencies for an issue. Direction: upstream, downstream, or both.</td></tr>
  <tr><td><code>doit_dependency_tree</code></td><td>Walk the parent-child hierarchy tree from a root issue.</td></tr>
</table>

<h3>Comments</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_add_comment</code></td><td>Add a comment to an issue.</td></tr>
  <tr><td><code>doit_list_comments</code></td><td>List comments on an issue, ordered by creation time.</td></tr>
</table>

<h3>Labels</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_add_label</code></td><td>Add a label to an issue.</td></tr>
  <tr><td><code>doit_remove_label</code></td><td>Remove a label from an issue.</td></tr>
</table>

<h3>Projects</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_create_project</code></td><td>Create a project within your tenant for organizing issues.</td></tr>
  <tr><td><code>doit_list_projects</code></td><td>List projects in your tenant. Returns project slugs for use with <code>project</code> filters.</td></tr>
</table>

<h3>Compaction</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_compact</code></td><td>Run semantic compaction on old closed issues. Summarizes issues to save context window tokens. Default threshold: 7 days.</td></tr>
</table>

<h3>Lessons Learned</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_record_lesson</code></td><td>Record a lesson learned — a mistake and its correction. Required: <code>title</code>, <code>mistake</code>, <code>correction</code>. Optional: <code>project</code> (slug), <code>issue_id</code>, <code>expert</code>, <code>components</code>, <code>severity</code>, <code>created_by</code>.</td></tr>
  <tr><td><code>doit_list_lessons</code></td><td>List lessons learned. All filters optional: <code>project</code> (slug), <code>status</code>, <code>expert</code>, <code>component</code>, <code>severity</code>, <code>limit</code>, <code>compact</code>. Returns <code>{count, has_more, items}</code> envelope. Defaults: <code>compact=true</code>, <code>limit=50</code>.</td></tr>
  <tr><td><code>doit_resolve_lesson</code></td><td>Mark a lesson as resolved. Required: <code>id</code>. Optional: <code>resolved_by</code>.</td></tr>
</table>

<h3>Retries (Operational Memory)</h3>
<p>Structured retry tracking for informed retry/escalate/abandon decisions. Attempt numbers are auto-computed per issue.
Retry history persists across sessions so agents can check how many times a task has been attempted.</p>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_record_retry</code></td><td>Record a retry attempt. Required: <code>issue_id</code>, <code>status</code> (failed/succeeded/abandoned/escalated), <code>error</code>. Optional: <code>project</code> (slug), <code>agent</code>, <code>created_by</code>. Attempt number auto-increments.</td></tr>
  <tr><td><code>doit_list_retries</code></td><td>List retry attempts for an issue, ordered by attempt number. Optional: <code>status</code>, <code>limit</code>.</td></tr>
</table>

<h3>Flags (Human Escalation)</h3>
<p>Durable escalation flags tied to work items. Issues with open severity 1-2 flags are excluded from <code>doit_ready()</code> output.
Flags survive session reset and are queryable across agent sessions.</p>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_raise_flag</code></td><td>Raise an escalation flag. Required: <code>issue_id</code>, <code>type</code>, <code>severity</code> (1=critical, 2=blocking, 3=warning), <code>summary</code>. Optional: <code>context</code> (JSON), <code>project</code> (slug), <code>created_by</code>.</td></tr>
  <tr><td><code>doit_list_flags</code></td><td>List escalation flags. All filters optional: <code>project</code>, <code>status</code> (open/acknowledged/resolved), <code>severity</code>, <code>issue_id</code>, <code>limit</code>.</td></tr>
  <tr><td><code>doit_resolve_flag</code></td><td>Resolve a flag with a decision. Required: <code>id</code>, <code>resolution</code>. Optional: <code>resolved_by</code>.</td></tr>
</table>

<h2>Admin Tools (10)</h2>
<p>Available on <code>POST /admin/mcp</code> — requires admin API key. Tenant keys receive 403.</p>

<h3>Tenant Management</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_create_tenant</code></td><td>Create a new tenant. Each tenant gets isolated data.</td></tr>
  <tr><td><code>doit_update_tenant</code></td><td>Update a tenant's name or slug. Accepts tenant slug as identifier. Provide <code>name</code> and/or <code>slug</code> to change.</td></tr>
  <tr><td><code>doit_list_tenants</code></td><td>List all tenants.</td></tr>
  <tr><td><code>doit_delete_tenant</code></td><td>Delete a tenant and its API keys. Rejects if projects still exist (delete them first). Accepts tenant slug.</td></tr>
  <tr><td><code>doit_create_api_key</code></td><td>Generate a new API key for a tenant. The raw key is returned once and cannot be retrieved again.</td></tr>
  <tr><td><code>doit_revoke_api_key</code></td><td>Revoke an API key by its 8-character prefix.</td></tr>
  <tr><td><code>doit_list_api_keys</code></td><td>List all API keys for a tenant.</td></tr>
</table>

<h3>Project Management</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_update_project</code></td><td>Update a project's name or slug. Accepts project ID or current slug as <code>project</code> identifier. Provide <code>name</code> and/or <code>slug</code> to change.</td></tr>
  <tr><td><code>doit_delete_project</code></td><td>Delete a project. Rejects if issues still reference the project. Accepts project slug.</td></tr>
</table>

<h3>Admin Key Management</h3>
<table>
  <tr><th>Tool</th><th>Description</th></tr>
  <tr><td><code>doit_rotate_admin_key</code></td><td>Generate a new admin API key and store its hash in the database. The raw key is returned once. The env var key (<code>API_KEY</code>) still works as fallback.</td></tr>
</table>

<h2>Data Model</h2>

<h3>Issue Statuses</h3>
<p>
  <span class="badge badge-open">open</span>
  <span class="badge" style="background:#fef3c7;color:#92400e;">in_progress</span>
  <span class="badge badge-blocked">blocked</span>
  <span class="badge badge-deferred">deferred</span>
  <span class="badge badge-closed">closed</span>
  <span class="badge" style="background:#e2e8f0;color:#475569;">pinned</span>
  <span class="badge" style="background:#e2e8f0;color:#475569;">hooked</span>
</p>

<h3>Issue Types</h3>
<p>
  <span class="badge badge-task">task</span>
  <span class="badge badge-bug">bug</span>
  <span class="badge badge-feature">feature</span>
  <span class="badge badge-epic">epic</span>
  <span class="badge" style="background:#e2e8f0;color:#475569;">chore</span>
  <span class="badge" style="background:#e2e8f0;color:#475569;">decision</span>
  <span class="badge" style="background:#e2e8f0;color:#475569;">molecule</span>
  <span class="badge" style="background:#e2e8f0;color:#475569;">event</span>
</p>

<h3>Lesson Statuses</h3>
<p>
  <span class="badge badge-open">open</span>
  <span class="badge badge-closed">resolved</span>
</p>

<h3>Flag Types</h3>
<p><code>structural_concern</code> &middot; <code>feature_concern</code> &middot; <code>red_flag</code> &middot; <code>human_decision</code> &middot; <code>security_concern</code></p>

<h3>Flag Statuses</h3>
<p>
  <span class="badge badge-open">open</span>
  <span class="badge" style="background:#fef3c7;color:#92400e;">acknowledged</span>
  <span class="badge badge-closed">resolved</span>
</p>

<h3>Flag Severity</h3>
<p><strong>1</strong> = critical &middot; <strong>2</strong> = blocking &middot; <strong>3</strong> = warning. Severity 1&ndash;2 flags on an issue exclude it from <code>doit_ready()</code>.</p>

<h3>Retry Statuses</h3>
<p><code>failed</code> &middot; <code>succeeded</code> &middot; <code>abandoned</code> &middot; <code>escalated</code></p>

<h3>Dependency Types (19)</h3>
<p><code>blocks</code> &middot; <code>conditional-blocks</code> &middot; <code>waits-for</code> &middot; <code>parent-child</code> &middot; <code>related</code> &middot; <code>relates-to</code> &middot; <code>discovered-from</code> &middot; <code>caused-by</code> &middot; <code>replies-to</code> &middot; <code>duplicates</code> &middot; <code>supersedes</code> &middot; <code>authored-by</code> &middot; <code>assigned-to</code> &middot; <code>approved-by</code> &middot; <code>attests</code> &middot; <code>validates</code> &middot; <code>tracks</code> &middot; <code>until</code> &middot; <code>delegated-from</code></p>

<h3>ID Format</h3>
<p>Doit uses <strong>hash-based IDs</strong> with a configurable prefix (default: <code>doit</code>). IDs are generated from a SHA-256 hash of a random seed, using the shortest unique prefix (minimum 7 chars).</p>
<p>Hierarchical children use dotted notation: <code>doit-abc1234.1</code>, <code>doit-abc1234.2</code>, etc.</p>

<h3>Tenant Isolation</h3>
<p>Every issue is scoped to the tenant of the authenticated API key. The <code>tenant_id</code> is automatically set on creation and enforced on all reads, updates, and deletes. Tenants cannot see or modify each other's issues.</p>

<h3>Priority</h3>
<p>Integer 0&ndash;4 where 0 is critical and 4 is backlog. Default: 2 (medium).</p>

<h2>Response Format</h2>

<h3>List Response Envelope</h3>
<p>All list endpoints (<code>doit_list_issues</code>, <code>doit_ready</code>, <code>doit_list_lessons</code>, <code>doit_list_comments</code>) return a response envelope:</p>
<pre><code>{
  "count": 12,        // Number of items in this response
  "has_more": true,   // True if more items exist beyond the limit
  "items": [...]      // The actual items (compact or full)
}</code></pre>

<h3>Response Size Protection</h3>
<p>List endpoints include server-side protection against oversized responses that can crash agent sessions:</p>
<ul>
  <li><strong>Default compact=true</strong> — all list endpoints return compact results by default. Set <code>compact=false</code> for full details.</li>
  <li><strong>Default limit=50</strong> — maximum items returned per call.</li>
  <li><strong>Hard cap at 20</strong> — when <code>compact=false</code> without a <code>project</code> filter, limit is capped at 20 to prevent oversized responses.</li>
  <li><strong>Auto-compact fallback</strong> — if a response exceeds 50K characters, it is automatically re-serialized in compact mode with <code>"auto_compacted": true</code> and a message.</li>
</ul>

<h2>Key Concepts</h2>

<h3>Ready Detection</h3>
<p>An issue is "ready" when it is <code>open</code>, has no unresolved <code>blocks</code> dependencies, and is not deferred to the future. Use <code>doit_ready</code> to find work.</p>

<h3>Hierarchical Tasks</h3>
<p>Issues can be nested: epic &rarr; task &rarr; subtask. Use <code>parent</code> when creating an issue to make it a child. Children get auto-numbered IDs like <code>parent.1</code>, <code>parent.2</code>.</p>

<h3>Semantic Compaction</h3>
<p>Old closed issues can be compacted to save context window tokens. The original content is preserved in a snapshot. Use <code>doit_compact</code> to trigger.</p>

<h3>Agent Messaging &rarr; TheHerald</h3>
<p>Agent-to-agent messaging is handled by <a href="https://herald.aoendpoint.com/documentation"><strong>TheHerald</strong></a>, a dedicated messaging MCP server. Herald provides typed messages (DO, ASK, TELL, HAND), conversation threading, signals (ACK, CLAIM, BLOCK, REJECT), and agent identity. Add Herald to your <code>.mcp.json</code> alongside Doit.</p>

<hr>

<h2>API</h2>
<table>
  <tr><th>Endpoint</th><th>Auth</th><th>Description</th></tr>
  <tr><td><code>GET /health</code></td><td>None</td><td>Health check</td></tr>
  <tr><td><code>POST /mcp</code></td><td>Bearer token</td><td>Agent MCP server (20 tools)</td></tr>
  <tr><td><code>POST /admin/mcp</code></td><td>Admin key</td><td>Admin MCP server (10 tools)</td></tr>
  <tr><td><code>GET /documentation</code></td><td>None</td><td>This page</td></tr>
  <tr><td><code>GET /ui/</code></td><td>Session cookie</td><td>Web UI (login with API key)</td></tr>
  <tr><td><code>GET /ui/admin/</code></td><td>Admin session</td><td>Admin UI (tenants, API keys, projects)</td></tr>
</table>

<hr>
<p style="color: #94a3b8; font-size: 0.85rem; margin-top: 2rem;">Doit MCP Server v{{VERSION}} — Built by Actual Outcomes</p>

</body>
</html>`
