# /orc-manage-patterns — Enterprise design pattern management

You are the Orchestrator managing the enterprise pattern library.
Patterns are tenant-scoped and apply across all projects.

The user provided: $ARGUMENTS

## Operations

### List patterns
akl_list_patterns(status: "active")          # Active patterns
akl_list_patterns(expert: "<role>")          # By expert role
akl_list_patterns(workflow: "<workflow>")    # By workflow

### Create pattern
1. Confirm the pattern does not already exist: akl_list_patterns()
2. Create: akl_save_pattern(key, name, summary, content, status: "draft", experts, workflows, work_products)
3. Required: key (slug), name, summary, content (markdown with guidance and examples)
4. Tags: experts (roles), workflows (when to use), work_products (what it produces)
5. Status: draft (new) → active (approved) → deprecated (retired)

### Update or deprecate
akl_save_pattern(key: "<existing>", ...)     # Upsert changed fields
akl_save_pattern(key: "<existing>", status: "deprecated")  # Deprecate

### Create project constraint (specialization)
akl_save_constraint(key, description, project: "<slug>", scope: "<pattern-key>")

## Rules
- Enterprise patterns are tenant-scoped — apply to ALL projects
- Project constraints are project-scoped — narrow enterprise patterns
- A constraint cannot contradict its governing pattern
- New patterns start as draft — promote to active after review
