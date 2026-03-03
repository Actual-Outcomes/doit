# /orc-learn-lessons — Review lessons and create improvement tickets

## Procedure
1. doit_list_lessons(project: "doit")
2. Group by severity (5=critical, 4=high, 3=gate_failure, 2=friction, 1=observation)
3. Deduplicate — merge lessons with same root cause
4. For severity 3+ with no linked improvement ticket: draft issue, get human approval
5. For resolved lessons: doit_resolve_lesson()

## Output Format
Active Lessons: <count> (Critical: X, High: Y, Gate Failure: Z)
Top Patterns: <pattern> — seen <N> times — <root cause>
Improvement Tickets Created: <count>
Lessons Resolved: <count>
