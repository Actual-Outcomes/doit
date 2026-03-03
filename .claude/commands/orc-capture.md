# /orc-capture — Lightweight backlog intake

You are the Orchestrator performing fast intake. **Capture, not elaboration** — get the request into Doit with just enough classification to be actionable later. Detailed breakdown happens when the item is picked up via `/orc-create-epic` or `/orc-orchestrate`.

The user provided: $ARGUMENTS

## Step 1: Parse & Auto-Classify

Classify the request using keyword matching:

| Type | `issue_type` | Signal Keywords | Label |
|------|-------------|-----------------|-------|
| Bug | `bug` | error, crash, broken, fix, regression, fails, doesn't work, 500, exception | `bug` |
| Procedure Change | `task` | process, workflow, rule, policy, governance, convention, procedure, spec change | `procedure-change` |
| Tooling | `task` | tool, script, automation, CLI, utility, DX, developer experience, command | `tooling` |
| User Need | `feature` | *(default — no other type matched)* | `user-need` |

## Step 2: Auto-Resolve Fields

| Field | Strategy |
|-------|----------|
| **Type** | Keyword match from Step 1 |
| **Project** | Default `doit`. If `$ARGUMENTS` names a different known project, use that. |
| **Priority** | Urgency language: blocker/critical/urgent → P0, important/soon/high → P1, default → P2, nice-to-have/low/someday → P3 |
| **Owner** | Default `doit-orchestrator`. If `$ARGUMENTS` names a person/role, use that. |

## Step 3: Bounded Ambiguity Resolution

Only if auto-resolution genuinely fails for any field:

1. **One menu** — present `AskUserQuestion` with 2-4 options per unresolved field (batch all into one call, max 4 questions)
2. **No follow-ups** — if still ambiguous after the menu, add `needs-triage` label and proceed with best-guess defaults

**Cap: Maximum ONE interaction round with the user.**

## Step 4: Overlap Check

Run `doit_list_issues(status: "open", project: "<project>", compact: true, limit: 50, issue_type: "all", priority: null, assignee: null, sort_by: "priority")`.

Scan titles for similarity to the incoming request. If a probable duplicate exists, present it and ask: create anyway or skip?

## Step 5: Create in Doit

```
doit_create_issue(
  title:               <normalized — imperative voice, 5-12 words>,
  description:         <structured format below>,
  design:              "To be elaborated",
  acceptance_criteria: "To be elaborated",
  notes:               "",
  priority:            <0-3>,
  issue_type:          <bug|feature|task>,
  assignee:            "unassigned",
  owner:               <resolved owner>,
  parent_id:           "",
  project:             <resolved project>,
  labels:              [<type-label>, ...any additional],
  ephemeral:           false
)
```

**Description format:**
```
## Raw Request
<verbatim $ARGUMENTS>

## Normalized Summary
<1-2 sentence distilled version>

## Capture Metadata
- Type: <classified type>
- Priority: P<n> (<auto-resolved|user-selected>)
- Source: /orc-capture
- Captured: <ISO date>
```

## Step 6: Print Receipt

```
Captured: <title>
ID: <doit-id>  |  Type: <type>  |  Priority: P<n>  |  Project: <project>

Resolution:
  Type:     <auto|menu>
  Priority: <auto|menu>
  Project:  <auto|menu>

Next: /orc-create-epic <doit-id>  — to elaborate into work packages
      /orc-orchestrate <doit-id>  — to execute directly (if simple enough)
```
