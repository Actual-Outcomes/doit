# /orc-init-feature — Scaffold a feature into the orchestration system

The user provided: $ARGUMENTS

## Step 1: Parse Request
Extract feature name, description, user need. Identify target PBS components.

## Step 2: Complexity Assessment
| Level | Lines | Components | Risk |
|-------|-------|------------|------|
| Trivial | <10 | 1 | LOW |
| Simple | <100 | 1-2 | LOW-MED |
| Moderate | <500 | 2-4 | MEDIUM |
| Complex | <2000 | 4+ | HIGH |
| Epic | 2000+ | System-wide | CRITICAL |

## Step 3: Feature Overlap Check
Search Feature Registry for similar features. If overlap: STOP, present conflict.

## Step 4: Acceptance Criteria
Draft criteria from user need. Each must be: Observable, Testable, Specific. Get human approval.

## Step 5: Scaffold Issues
Create epic/task issues in Doit with dependencies. Update Feature Registry.

## Step 6: Output
feature_key, complexity, risk, issues_created, acceptance_criteria, ready_for: /orc-orchestrate
