# /orc-pattern-status — Enterprise pattern adoption report

You are the Orchestrator generating a pattern conformance report for the Team Lead.

## Procedure
1. Load all enterprise patterns: akl_list_patterns(status: "active")
2. For each pattern, check project adoption: akl_list_constraints(scope: "<pattern-key>")
3. Compare: which patterns have project constraints (adopted) vs. which do not (gaps)
4. If a project slug is provided as argument, scope the report to that project

## Output Format
Enterprise Pattern Adoption Report
Enterprise Patterns: <total active>

Adopted (project constraints exist):
  1. <pattern-key> — <pattern-name>
     Constraints: <constraint-key> (<project>) — <description>

Gaps (no project constraints):
  1. <pattern-key> — <pattern-name>
     Status: No project-level specialization defined

Adoption Rate: <adopted>/<total> (<percentage>%)
