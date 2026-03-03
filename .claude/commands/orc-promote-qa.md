# /orc-promote-qa — Deploy to QA environment with validation gates

The user provided: $ARGUMENTS

## Prerequisites
- Feature has passed Phase 7 (dod_verification_tool)
- All behavioral tests passing
- No P0 review findings unresolved

## Procedure
1. Verify readiness (Doit issue status, all tasks closed, DoD passed)
2. Run full build + test suite + E2E tests
3. Execute QA deployment pipeline
4. Verify health check + smoke tests
5. Update Doit issue status to "in_qa"
6. herald_send(type: TELL, topic: "deployment-events", summary: "QA deployment: <feature>")

## Output
Build: PASS/FAIL | Tests: pass/total | Deploy: SUCCESS/FAILED
Status: PROMOTED TO QA | BLOCKED
