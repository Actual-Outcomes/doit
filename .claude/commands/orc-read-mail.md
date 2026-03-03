# /orc-read-mail — Check Herald inbox for the project lead

You are the Orchestrator checking your mailbox. Read, triage, and respond to messages from other agents.

## Step 1: Register & Fetch

1. Register: `herald_register(key: "doit-orchestrator", name: "Doit Orchestrator", agent_type: "orchestrator", project: "doit")`
2. Fetch unread: `herald_inbox(agent_key: "doit-orchestrator", unread: true)`
3. If no unread messages: report "Inbox clear — no unread messages." and stop.

## Step 2: Summarize

Present a table of all unread messages:

```
| # | From | Type | Project | Summary | Received |
|---|------|------|---------|---------|----------|
```

## Step 3: Triage & Respond

Process each message by type, in priority order:

| Type | Priority | Action |
|------|----------|--------|
| **DO** | 1 — Highest | CLAIM immediately. Evaluate scope: if it's a doit-project task, create a Doit issue and begin work. If out of scope, REJECT with explanation. |
| **ASK** | 2 | Read full message with `herald_get`. Formulate ANSWER. If you need human input to answer, present the question to the user. |
| **HELP** | 3 | Read full message. Evaluate if you can assist. If yes, respond with TELL. If not, present to user for guidance. |
| **TELL** | 4 | ACK and note any action items. If it references a doit issue, update the issue notes. |
| **DONE** | 5 | ACK. If it closes something you requested, update the corresponding doit issue. |
| **BLOCK** | 5 | Read context. If you can unblock, respond. Otherwise, raise a flag: `doit_raise_flag(type: "human_decision", severity: 2)` on the related issue. |
| **REJECT** | 5 | ACK. Note the rejection reason. Present to user if it was something they requested. |
| **ACK/CLAIM** | 6 — Lowest | Note and mark read. No response needed. |

For each message processed, mark as read.

## Step 4: Report

After processing all messages, present:

```
Mail Report
━━━━━━━━━━━
Received:  <count> unread messages
Processed: <count>
  DO:      <count claimed>
  ASK:     <count answered>
  TELL:    <count acknowledged>
  Other:   <count>

Actions Taken:
- <list of issues created, updated, or flags raised>

Pending (needs human input):
- <list of items requiring user decision, if any>
```
