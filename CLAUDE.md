# Doit - AI Agent Work Planner & Tracker

## Project Overview
Doit is a CLI and MCP server for AI agents to plan and track work. It is a replacement for [beads](https://github.com/steveyegge/beads), built on the same Go + PostgreSQL + MCP patterns as [getsit](https://github.com/Actual-Outcomes/getsit).

## Architecture
- **Language:** Go
- **Storage:** PostgreSQL (shared with getsit pattern)
- **Interfaces:** CLI (`doit`) + MCP server (`doit-server`)
- **Module:** `github.com/Actual-Outcomes/doit`

## Project Structure
```
cmd/doit/          - CLI binary entry point
cmd/doit-server/   - MCP server entry point
internal/api/      - MCP tool handlers and registration
internal/cli/      - CLI command implementations
internal/config/   - Configuration loading
internal/model/    - Data model types
internal/store/    - PostgreSQL store layer
internal/version/  - Version info
migrations/        - SQL migrations (goose)
scripts/           - Utility scripts
tests/             - Integration and smoke tests
```

## Key Design Decisions
- Follow getsit's proven patterns (chi router, goose migrations, go-sdk/mcp)
- CLI and MCP server share the same store layer
- Hash-based task IDs (like beads) to prevent merge conflicts
- Hierarchical tasks: epic > task > subtask
- Dependency graph with automatic "ready" detection

## Development Commands
```bash
go build ./cmd/doit          # Build CLI
go build ./cmd/doit-server   # Build MCP server
go test ./...                # Run all tests
```

## Feature Scope (beads parity)
- Task CRUD with hash-based IDs
- Hierarchical tasks (epic.task.subtask)
- Dependency tracking and "ready" detection
- Semantic compaction (memory decay)
- Agent-to-agent messaging with threading
- Multi-tenant with API keys
- JSON output for agent consumption

## AKL (Architecture Knowledge Layer)
This project is tracked in AKL at `ama.aoendpoint.com`. The MCP config is in `.mcp.json`.

**At the start of every session:**
1. Call `akl_overview` to orient yourself
2. Call `akl_recent_changes` with `since: 24h` to see what changed

**When making changes:**
- Search AKL before creating new components (`akl_search`)
- Update AKL when creating/modifying components, relationships, or decisions
- Record architectural decisions as ADRs via `akl_save_decision`

**AKL project slug:** `doit`

## Work Tracking (Doit)
This project uses its own doit MCP server at `din.aoendpoint.com` for tracking development work.
- `doit_list_projects` — list available projects
- `doit_ready` — see what's unblocked and ready for work
- `doit_get_issue` — view issue details
- `doit_update_issue` with claim=true — claim and start work
- `doit_create_issue` — create new work items (set `project` slug)
