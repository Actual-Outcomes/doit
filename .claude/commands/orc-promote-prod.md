# /orc-promote-prod — Deploy to production with human approval gates

The user provided: $ARGUMENTS

## Prerequisites
- Feature has been in QA with no regressions
- Human approval obtained (MANDATORY — do NOT proceed without it)

## Procedure
1. Human approval gate — present summary, wait for explicit approval
2. Pre-deployment: full test suite, no new commits since QA, check for STOP messages
3. Deploy: execute pipeline, monitor health checks, run smoke tests
4. Post-deployment: verify health, monitor error rates 5 minutes
5. Update Doit, record in journal, herald_send(topic: "deployment-events")

## Rollback Trigger
If any post-deployment check fails:
1. Execute rollback immediately
2. doit_record_lesson(severity: 4)
3. herald_send(type: HELP, summary: "Production rollback: <feature>")

## Output
Human Approval: OBTAINED/PENDING | Deploy: SUCCESS/FAILED | Status: DEPLOYED | ROLLED BACK | BLOCKED
