# Governance Gates — 7 reusable validation procedures

All gates return: status (PASS|FAIL|WARN), details, failures[], requires_human_approval, herald_escalation
On failure: doit_record_lesson(severity: 3). On critical: herald_send(type: HELP).

## Gate 1: orientation_tool
Phase: 2 | Blocks: Planning
Load: akl_overview, akl_get_component (each key), akl_impact_analysis, akl_list_decisions, akl_list_constraints, akl_recent_changes(7d), doit_list_issues(in_progress), doit_list_lessons(severity 3+), herald_inbox

## Gate 2: check_feature_uniqueness
Phase: 2 (new features) | Blocks: Planning
Search Feature Registry for duplicates. FAIL if exact duplicate found.

## Gate 3: check_surface_ownership
Phase: 2 (UI changes) | Blocks: Planning
Check UX Surface Map for route/component conflicts. FAIL if exclusive surface claimed.

## Gate 4: check_structural_boundaries
Phase: 5 | Blocks: Delivery
akl_resolve_file for each changed file. Check cross-component import violations. FAIL unconditionally.

## Gate 5: check_feature_preservation
Phase: 5 | Blocks: Delivery
Run behavioral tests for affected features. FAIL cannot be bypassed.

## Gate 6: validate_product_model
Phase: 6 | Blocks: Delivery
Cross-registry referential integrity: Feature Registry <-> AKL <-> Surface Map <-> tests.

## Gate 7: dod_verification_tool
Phase: 7 | Blocks: Delivery | Depends on: gates 4, 5, 6
9-point checklist: behavioral tests present + passing, annotations, traceability, registry, no unauthorized deprecations, AKL updated, features.md, structural compliance.
