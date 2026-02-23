package ui

const baseLayout = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{.Title}} — Doit</title>
  <style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      line-height: 1.6;
      color: #1a1a2e;
      background: #f8f9fa;
    }
    a { color: #2563eb; text-decoration: none; }
    a:hover { text-decoration: underline; }

    /* Nav */
    .nav {
      background: #1e293b;
      color: #e2e8f0;
      padding: 0 2rem;
      display: flex;
      align-items: center;
      height: 56px;
      gap: 2rem;
    }
    .nav-brand {
      font-weight: 700;
      font-size: 1.2rem;
      color: #fff;
      text-decoration: none;
    }
    .nav-links { display: flex; gap: 1rem; flex: 1; }
    .nav-links a {
      color: #94a3b8;
      padding: 0.25rem 0.75rem;
      border-radius: 4px;
      font-size: 0.9rem;
      text-decoration: none;
    }
    .nav-links a:hover, .nav-links a.active {
      color: #fff;
      background: #334155;
    }
    .nav-project { margin-left: auto; }
    .nav-project select {
      background: #334155;
      color: #e2e8f0;
      border: 1px solid #475569;
      padding: 0.25rem 0.5rem;
      border-radius: 4px;
      font-size: 0.85rem;
      cursor: pointer;
    }
    .nav-project select:hover { border-color: #64748b; }
    .nav-right { margin-left: 0.75rem; }
    .nav-right form { display: inline; }
    .nav-right button {
      background: none;
      border: 1px solid #475569;
      color: #94a3b8;
      padding: 0.25rem 0.75rem;
      border-radius: 4px;
      cursor: pointer;
      font-size: 0.85rem;
    }
    .nav-right button:hover { color: #fff; border-color: #64748b; }

    /* Content */
    .content { max-width: 1100px; margin: 0 auto; padding: 2rem; }
    h1 { font-size: 1.8rem; margin-bottom: 1rem; color: #16213e; }
    h2 { font-size: 1.4rem; margin-top: 1.5rem; margin-bottom: 0.75rem; color: #16213e; }

    /* Cards */
    .stat-cards {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 1rem;
      margin-bottom: 2rem;
    }
    .stat-card {
      background: #fff;
      border-radius: 8px;
      padding: 1.25rem;
      box-shadow: 0 1px 3px rgba(0,0,0,0.1);
    }
    .stat-card .label { font-size: 0.85rem; color: #64748b; margin-bottom: 0.25rem; }
    .stat-card .value { font-size: 2rem; font-weight: 700; color: #16213e; }
    .stat-card.open .value { color: #2563eb; }
    .stat-card.progress .value { color: #d97706; }
    .stat-card.blocked .value { color: #dc2626; }
    .stat-card.ready .value { color: #059669; }

    /* Tables */
    table { width: 100%; border-collapse: collapse; background: #fff; border-radius: 8px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
    th, td { text-align: left; padding: 0.65rem 1rem; border-bottom: 1px solid #f1f5f9; }
    th { background: #f8fafc; font-weight: 600; font-size: 0.85rem; color: #475569; text-transform: uppercase; letter-spacing: 0.05em; }
    td { font-size: 0.9rem; }
    tr:hover td { background: #f8fafc; }

    /* Badges */
    .badge {
      display: inline-block;
      padding: 0.15em 0.55em;
      border-radius: 4px;
      font-size: 0.78em;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: 0.03em;
    }
    .badge-open { background: #dbeafe; color: #1e40af; }
    .badge-progress { background: #fef3c7; color: #92400e; }
    .badge-blocked { background: #fee2e2; color: #991b1b; }
    .badge-deferred { background: #e2e8f0; color: #475569; }
    .badge-closed { background: #d1fae5; color: #065f46; }
    .badge-default { background: #f1f5f9; color: #64748b; }
    .badge-bug { background: #fee2e2; color: #991b1b; }
    .badge-feature { background: #ede9fe; color: #6d28d9; }
    .badge-epic { background: #dbeafe; color: #1e40af; }
    .badge-p0 { background: #fee2e2; color: #991b1b; }
    .badge-p1 { background: #fef3c7; color: #92400e; }
    .badge-p2 { background: #e2e8f0; color: #475569; }
    .badge-p3 { background: #f1f5f9; color: #64748b; }
    .badge-p4 { background: #f1f5f9; color: #94a3b8; }

    /* Filters */
    .filters {
      display: flex;
      gap: 0.75rem;
      margin-bottom: 1.5rem;
      flex-wrap: wrap;
      align-items: center;
    }
    .filters select, .filters input {
      padding: 0.4rem 0.75rem;
      border: 1px solid #cbd5e1;
      border-radius: 6px;
      font-size: 0.9rem;
      background: #fff;
    }
    .filters button {
      padding: 0.4rem 1rem;
      background: #2563eb;
      color: #fff;
      border: none;
      border-radius: 6px;
      cursor: pointer;
      font-size: 0.9rem;
    }
    .filters button:hover { background: #1d4ed8; }
    .filters a.clear {
      padding: 0.4rem 0.75rem;
      color: #64748b;
      font-size: 0.85rem;
    }

    /* Detail */
    .detail-header { margin-bottom: 1.5rem; }
    .detail-header h1 { margin-bottom: 0.25rem; }
    .detail-meta { color: #64748b; font-size: 0.9rem; }
    .detail-meta code {
      background: #f1f5f9;
      padding: 0.1em 0.4em;
      border-radius: 3px;
      font-size: 0.85em;
    }
    .detail-body {
      background: #fff;
      border-radius: 8px;
      padding: 1.5rem;
      box-shadow: 0 1px 3px rgba(0,0,0,0.1);
      margin-bottom: 1.5rem;
    }
    .detail-body h3 { font-size: 1rem; color: #475569; margin-bottom: 0.5rem; margin-top: 1rem; }
    .detail-body h3:first-child { margin-top: 0; }
    .detail-body pre {
      background: #f8fafc;
      padding: 0.75rem;
      border-radius: 6px;
      overflow-x: auto;
      white-space: pre-wrap;
      font-size: 0.9rem;
    }
    .detail-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 0.5rem 2rem;
      margin-bottom: 1rem;
    }
    .detail-grid dt { font-size: 0.85rem; color: #64748b; }
    .detail-grid dd { font-size: 0.95rem; margin-bottom: 0.5rem; }
    .label-list { display: flex; gap: 0.5rem; flex-wrap: wrap; }
    .label-tag {
      background: #f1f5f9;
      color: #334155;
      padding: 0.1em 0.5em;
      border-radius: 3px;
      font-size: 0.85rem;
    }

    /* Comments */
    .comment {
      border-left: 3px solid #e2e8f0;
      padding: 0.5rem 1rem;
      margin-bottom: 0.75rem;
    }
    .comment-meta { font-size: 0.8rem; color: #64748b; margin-bottom: 0.25rem; }
    .comment-text { font-size: 0.9rem; }

    /* Login */
    .login-box {
      max-width: 400px;
      margin: 4rem auto;
      background: #fff;
      padding: 2rem;
      border-radius: 12px;
      box-shadow: 0 4px 6px rgba(0,0,0,0.1);
    }
    .login-box h1 { text-align: center; margin-bottom: 0.5rem; }
    .login-box .subtitle { text-align: center; color: #64748b; margin-bottom: 1.5rem; font-size: 0.9rem; }
    .login-box input[type="password"] {
      width: 100%;
      padding: 0.6rem 0.75rem;
      border: 1px solid #cbd5e1;
      border-radius: 6px;
      font-size: 1rem;
      margin-bottom: 1rem;
    }
    .login-box button {
      width: 100%;
      padding: 0.6rem;
      background: #2563eb;
      color: #fff;
      border: none;
      border-radius: 6px;
      font-size: 1rem;
      cursor: pointer;
      font-weight: 600;
    }
    .login-box button:hover { background: #1d4ed8; }
    .login-error { color: #dc2626; font-size: 0.9rem; margin-bottom: 1rem; text-align: center; }

    /* Error page */
    .error-box {
      max-width: 500px;
      margin: 4rem auto;
      text-align: center;
    }
    .error-box .code { font-size: 4rem; font-weight: 700; color: #cbd5e1; }
    .error-box .message { font-size: 1.1rem; color: #64748b; margin-top: 0.5rem; }

    /* Empty state */
    .empty { text-align: center; padding: 3rem; color: #94a3b8; font-size: 1rem; }

    /* Footer */
    .footer { text-align: center; color: #94a3b8; font-size: 0.8rem; padding: 2rem; }
  </style>
</head>
<body>
{{if .ShowNav}}
<nav class="nav">
  <a href="/ui/" class="nav-brand">Doit</a>
  <div class="nav-links">
    <a href="/ui/" {{if eq .NavActive "dashboard"}}class="active"{{end}}>Dashboard</a>
    <a href="/ui/issues" {{if eq .NavActive "issues"}}class="active"{{end}}>Issues</a>
    <a href="/ui/ready" {{if eq .NavActive "ready"}}class="active"{{end}}>Ready</a>
  </div>
  {{if .Projects}}
  <div class="nav-project">
    <form method="POST" action="/ui/project">
      <select name="project_id" onchange="this.form.submit()">
        <option value="">All Projects</option>
        {{range .Projects}}
        <option value="{{.ID}}" {{if eq (printf "%s" .ID) $.CurrentProject}}selected{{end}}>{{.Name}}</option>
        {{end}}
      </select>
    </form>
  </div>
  {{end}}
  <div class="nav-right">
    <form method="POST" action="/ui/logout">
      <button type="submit">Logout</button>
    </form>
  </div>
</nav>
{{end}}
<div class="content">
  {{template "page" .}}
</div>
<div class="footer">Doit &mdash; AI Agent Work Planner &middot; <a href="/documentation">Documentation</a></div>
</body>
</html>`

const loginPage = `{{define "page"}}
<div class="login-box">
  <h1>Doit</h1>
  <p class="subtitle">AI Agent Work Planner</p>
  {{if .Error}}<p class="login-error">{{.Error}}</p>{{end}}
  <form method="POST" action="/ui/login">
    <input type="password" name="api_key" placeholder="Enter your API key" autofocus>
    <button type="submit">Sign In</button>
  </form>
</div>
{{end}}`

const dashboardPage = `{{define "page"}}
<h1>Dashboard</h1>

<div class="stat-cards">
  <div class="stat-card open">
    <div class="label">Open</div>
    <div class="value">{{index .Counts "open"}}</div>
  </div>
  <div class="stat-card progress">
    <div class="label">In Progress</div>
    <div class="value">{{index .Counts "in_progress"}}</div>
  </div>
  <div class="stat-card blocked">
    <div class="label">Blocked</div>
    <div class="value">{{index .Counts "blocked"}}</div>
  </div>
  <div class="stat-card ready">
    <div class="label">Ready</div>
    <div class="value">{{len .Ready}}</div>
  </div>
</div>

{{if .Ready}}
<h2>Ready for Work</h2>
<table>
  <thead><tr><th>ID</th><th>Title</th><th>Type</th><th>Priority</th></tr></thead>
  <tbody>
  {{range .Ready}}
  <tr>
    <td><a href="/ui/issues/{{.ID}}"><code>{{.ID}}</code></a></td>
    <td><a href="/ui/issues/{{.ID}}">{{truncate .Title 60}}</a></td>
    <td><span class="badge badge-{{typeClass .IssueType}}">{{.IssueType}}</span></td>
    <td><span class="badge badge-p{{.Priority}}">{{priorityLabel .Priority}}</span></td>
  </tr>
  {{end}}
  </tbody>
</table>
{{end}}

{{if .Recent}}
<h2>Recently Updated</h2>
<table>
  <thead><tr><th>ID</th><th>Title</th><th>Status</th><th>Type</th><th>Updated</th></tr></thead>
  <tbody>
  {{range .Recent}}
  <tr>
    <td><a href="/ui/issues/{{.ID}}"><code>{{.ID}}</code></a></td>
    <td><a href="/ui/issues/{{.ID}}">{{truncate .Title 50}}</a></td>
    <td><span class="badge badge-{{statusClass .Status}}">{{replace (upper (string .Status)) "_" " "}}</span></td>
    <td><span class="badge badge-{{typeClass .IssueType}}">{{.IssueType}}</span></td>
    <td style="color:#64748b;font-size:0.85rem">{{.UpdatedAt.Format "Jan 2 15:04"}}</td>
  </tr>
  {{end}}
  </tbody>
</table>
{{end}}

{{if and (not .Ready) (not .Recent)}}
<div class="empty">No issues yet. Create some via the MCP tools!</div>
{{end}}
{{end}}`

const issuesPage = `{{define "page"}}
<h1>Issues</h1>

<form class="filters" method="GET" action="/ui/issues">
  <select name="status">
    <option value="">All Statuses</option>
    <option value="open" {{if eq .FilterStatus "open"}}selected{{end}}>Open</option>
    <option value="in_progress" {{if eq .FilterStatus "in_progress"}}selected{{end}}>In Progress</option>
    <option value="blocked" {{if eq .FilterStatus "blocked"}}selected{{end}}>Blocked</option>
    <option value="deferred" {{if eq .FilterStatus "deferred"}}selected{{end}}>Deferred</option>
    <option value="closed" {{if eq .FilterStatus "closed"}}selected{{end}}>Closed</option>
  </select>
  <select name="type">
    <option value="">All Types</option>
    <option value="task" {{if eq .FilterType "task"}}selected{{end}}>Task</option>
    <option value="bug" {{if eq .FilterType "bug"}}selected{{end}}>Bug</option>
    <option value="feature" {{if eq .FilterType "feature"}}selected{{end}}>Feature</option>
    <option value="epic" {{if eq .FilterType "epic"}}selected{{end}}>Epic</option>
    <option value="chore" {{if eq .FilterType "chore"}}selected{{end}}>Chore</option>
  </select>
  <select name="priority">
    <option value="">All Priorities</option>
    <option value="0" {{if eq .FilterPriority "0"}}selected{{end}}>P0 Critical</option>
    <option value="1" {{if eq .FilterPriority "1"}}selected{{end}}>P1 High</option>
    <option value="2" {{if eq .FilterPriority "2"}}selected{{end}}>P2 Medium</option>
    <option value="3" {{if eq .FilterPriority "3"}}selected{{end}}>P3 Low</option>
    <option value="4" {{if eq .FilterPriority "4"}}selected{{end}}>P4 Backlog</option>
  </select>
  <input type="text" name="assignee" placeholder="Assignee" value="{{.FilterAssignee}}">
  <button type="submit">Filter</button>
  <a class="clear" href="/ui/issues">Clear</a>
</form>

{{if .Issues}}
<table>
  <thead><tr><th>ID</th><th>Title</th><th>Status</th><th>Type</th><th>Priority</th><th>Assignee</th><th>Updated</th></tr></thead>
  <tbody>
  {{range .Issues}}
  <tr>
    <td><a href="/ui/issues/{{.ID}}"><code>{{.ID}}</code></a></td>
    <td><a href="/ui/issues/{{.ID}}">{{truncate .Title 45}}</a></td>
    <td><span class="badge badge-{{statusClass .Status}}">{{replace (upper (string .Status)) "_" " "}}</span></td>
    <td><span class="badge badge-{{typeClass .IssueType}}">{{.IssueType}}</span></td>
    <td><span class="badge badge-p{{.Priority}}">P{{.Priority}}</span></td>
    <td>{{if .Assignee}}{{.Assignee}}{{else}}<span style="color:#94a3b8">—</span>{{end}}</td>
    <td style="color:#64748b;font-size:0.85rem">{{.UpdatedAt.Format "Jan 2 15:04"}}</td>
  </tr>
  {{end}}
  </tbody>
</table>
{{else}}
<div class="empty">No issues match the current filters.</div>
{{end}}
{{end}}`

const issueDetailPage = `{{define "page"}}
<div class="detail-header">
  <h1>{{.Issue.Title}}</h1>
  <div class="detail-meta">
    <code>{{.Issue.ID}}</code>
    &middot;
    <span class="badge badge-{{statusClass .Issue.Status}}">{{replace (upper (string .Issue.Status)) "_" " "}}</span>
    &middot;
    <span class="badge badge-{{typeClass .Issue.IssueType}}">{{.Issue.IssueType}}</span>
    &middot;
    <span class="badge badge-p{{.Issue.Priority}}">{{priorityLabel .Issue.Priority}}</span>
  </div>
</div>

<div class="detail-body">
  <dl class="detail-grid">
    <dt>Assignee</dt>
    <dd>{{if .Issue.Assignee}}{{.Issue.Assignee}}{{else}}<span style="color:#94a3b8">Unassigned</span>{{end}}</dd>
    <dt>Owner</dt>
    <dd>{{if .Issue.Owner}}{{.Issue.Owner}}{{else}}<span style="color:#94a3b8">—</span>{{end}}</dd>
    <dt>Created</dt>
    <dd>{{.Issue.CreatedAt.Format "2006-01-02 15:04"}}{{if .Issue.CreatedBy}} by {{.Issue.CreatedBy}}{{end}}</dd>
    <dt>Updated</dt>
    <dd>{{.Issue.UpdatedAt.Format "2006-01-02 15:04"}}</dd>
    {{if .Issue.ParentID}}
    <dt>Parent</dt>
    <dd><a href="/ui/issues/{{.Issue.ParentID}}"><code>{{.Issue.ParentID}}</code></a></dd>
    {{end}}
  </dl>

  {{if .Issue.Labels}}
  <h3>Labels</h3>
  <div class="label-list">
    {{range .Issue.Labels}}<span class="label-tag">{{.}}</span>{{end}}
  </div>
  {{end}}

  {{if .Issue.Description}}
  <h3>Description</h3>
  <pre>{{.Issue.Description}}</pre>
  {{end}}

  {{if .Issue.Design}}
  <h3>Design</h3>
  <pre>{{.Issue.Design}}</pre>
  {{end}}

  {{if .Issue.AcceptanceCriteria}}
  <h3>Acceptance Criteria</h3>
  <pre>{{.Issue.AcceptanceCriteria}}</pre>
  {{end}}

  {{if .Issue.Notes}}
  <h3>Notes</h3>
  <pre>{{.Issue.Notes}}</pre>
  {{end}}

  {{if .Issue.CloseReason}}
  <h3>Close Reason</h3>
  <pre>{{.Issue.CloseReason}}</pre>
  {{end}}
</div>

{{if .Dependencies}}
<h2>Dependencies</h2>
<table>
  <thead><tr><th>Direction</th><th>Type</th><th>Issue</th></tr></thead>
  <tbody>
  {{range .Dependencies}}
  <tr>
    {{if eq .IssueID $.Issue.ID}}
    <td>Depends on</td>
    <td><code>{{.Type}}</code></td>
    <td><a href="/ui/issues/{{.DependsOnID}}"><code>{{.DependsOnID}}</code></a></td>
    {{else}}
    <td>Depended on by</td>
    <td><code>{{.Type}}</code></td>
    <td><a href="/ui/issues/{{.IssueID}}"><code>{{.IssueID}}</code></a></td>
    {{end}}
  </tr>
  {{end}}
  </tbody>
</table>
{{end}}

{{if .Comments}}
<h2>Comments</h2>
{{range .Comments}}
<div class="comment">
  <div class="comment-meta"><strong>{{.Author}}</strong> &middot; {{.CreatedAt.Format "Jan 2 15:04"}}</div>
  <div class="comment-text">{{.Text}}</div>
</div>
{{end}}
{{end}}
{{end}}`

const readyPage = `{{define "page"}}
<h1>Ready for Work</h1>
<p style="color:#64748b;margin-bottom:1.5rem">Issues that are open, unblocked, and not deferred.</p>

{{if .Issues}}
<table>
  <thead><tr><th>ID</th><th>Title</th><th>Type</th><th>Priority</th><th>Assignee</th><th>Created</th></tr></thead>
  <tbody>
  {{range .Issues}}
  <tr>
    <td><a href="/ui/issues/{{.ID}}"><code>{{.ID}}</code></a></td>
    <td><a href="/ui/issues/{{.ID}}">{{truncate .Title 50}}</a></td>
    <td><span class="badge badge-{{typeClass .IssueType}}">{{.IssueType}}</span></td>
    <td><span class="badge badge-p{{.Priority}}">{{priorityLabel .Priority}}</span></td>
    <td>{{if .Assignee}}{{.Assignee}}{{else}}<span style="color:#94a3b8">—</span>{{end}}</td>
    <td style="color:#64748b;font-size:0.85rem">{{.CreatedAt.Format "Jan 2"}}</td>
  </tr>
  {{end}}
  </tbody>
</table>
{{else}}
<div class="empty">No issues are ready for work right now.</div>
{{end}}
{{end}}`

const errorPage = `{{define "page"}}
<div class="error-box">
  <div class="code">{{.Code}}</div>
  <div class="message">{{.Message}}</div>
  <p style="margin-top:1.5rem"><a href="/ui/">Back to Dashboard</a></p>
</div>
{{end}}`
