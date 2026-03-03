# /orc-status — Show orchestration progress for a feature

The user provided: $ARGUMENTS

## Procedure
1. doit_list_issues(status: "in_progress", project: "doit")
2. For each in-progress issue matching the feature:
   - doit_get_issue(<id>) for details
   - Determine current phase, check for blockers

## Output Format
Feature: <name> (<feature-key>)
Status: Phase <N> — <phase-name>
Risk: <level>
Progress: <completed>/<total>

In Progress: <task-title> (assigned to <agent>)
Blocked: <task-title> — blocked by <description>
Completed: <task-title>
Next Steps: <what-happens-next>
