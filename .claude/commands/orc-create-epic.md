# /orc-create-epic — Turn a user need into backlog work packages

You are the Orchestrator creating Doit backlog items from a user need or story.

**Key guardrail: Define WHAT needs to be done, not HOW to build it.**

The user provided: $ARGUMENTS

## Step 1: Parse User Need
1. Extract the core user need, story, or goal
2. Identify scope boundaries (what's in, what's out)
3. If ambiguous or too broad: STOP and ask clarifying questions

## Step 2: Check for Overlaps
1. doit_list_issues(status: "open", issue_type: "epic", project: "doit", limit: 50)
2. Scan existing epics for overlap
3. If overlap found: present conflict and ask whether to extend or create new

## Step 3: Break Down into Epic + Tasks
For each task:
- Title: Clear, actionable (imperative form)
- Description: What needs to happen (not how)
- Acceptance criteria: Observable, testable outcomes
- Priority: P0 (critical) / P1 (important) / P2 (normal) / P3 (nice-to-have)

Rules: Tasks describe outcomes, not implementation steps. Each should be completable in a single /orc-orchestrate session.

## Step 4: Wire Dependencies
Data model tasks before API tasks. Core functionality before extensions.

## Step 5: Create in Doit
1. doit_create_issue(issue_type: "epic", ...)
2. doit_create_issue(parent_id: "<epic-id>", ...)
3. doit_add_dependency(issue_id, depends_on_id, type: "blocks")

## Step 6: Output Summary
Epic Created: <title>
ID: <epic-id> | Priority: <P-level> | Tasks: <count>
Dependencies: <count> blocking relationships wired
Ready to Start: <list of unblocked task IDs>
