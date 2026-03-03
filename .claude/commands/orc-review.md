# /orc-review — Code review checklist

Review local changes before pushing.

The user provided: $ARGUMENTS

## Procedure
1. Get the diff: git diff (unstaged) or git diff --cached (staged)
2. Load PBS context for affected components
3. Load Feature Registry for affected features
4. Dispatch Review Agent with diff + context

## Review Checklist
### Structural Compliance
- Files within PBS component boundaries
- No cross-component import violations
- New files in correct PBS locations

### Feature Compliance
- No existing features duplicated or degraded
- Behavioral tests adequate

### Code Quality
- No OWASP Top 10 / CWE Top 25 issues
- Performance anti-patterns identified
- Convention compliance

### Test Adequacy
- Unit tests for new logic
- Integration tests for new interfaces
- All tests passing

## Output
Blocking [P0], Suggestions [P1], Notes [P2]
Verdict: PASS | PASS_WITH_SUGGESTIONS | FAIL
