# Agent Instructions

This project uses **doit** MCP server for issue tracking. The server is configured in `.mcp.json`.

## Quick Reference

Use doit MCP tools (not CLI commands) for work tracking:

- `doit_ready` — Find available work (unblocked issues). Use `project` (slug) to scope to a project.
- `doit_list_issues` — List/filter issues. Use `project` (slug) to scope to a project.
- `doit_get_issue` — View issue details
- `doit_update_issue` with claim=true — Claim work
- `doit_close_issue` — Complete work
- `doit_create_issue` — Create new issues (set `project` slug)
- `doit_list_projects` — List available projects

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues via `doit_create_issue`
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work via `doit_close_issue`, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
