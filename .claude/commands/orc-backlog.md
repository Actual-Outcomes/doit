# /orc-backlog — Project health dashboard

## Procedure
1. doit_list_issues(status: "open", project: "doit")
2. doit_list_issues(status: "in_progress", project: "doit")
3. doit_list_issues(project: "doit") for total counts

For each blocked item, call doit_get_issue(<blocker-id>) to identify root blockers.

## Output Format
Project Health: Doit (AI Work Tracker)

Summary:
  Open: <count> | In Progress: <count> | Closed: <count> | Blocked: <count>

By Type:
  Epics: <count> | Tasks: <count> | Bugs: <count>

Top Blockers:
  1. <blocker-title> — blocking <N> items

Ready to Start (dependencies resolved):
  1. <issue-title> (priority: <P0/P1/P2>)

Recent Activity (7 days):
  Created: <count> | Closed: <count> | Lessons recorded: <count>
