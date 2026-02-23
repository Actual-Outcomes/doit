package store

import (
	"context"
	"fmt"

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/Actual-Outcomes/doit/internal/model"
)

// CreateProject creates a new project within the tenant from context.
func (s *PgStore) CreateProject(ctx context.Context, name, slug string) (*model.Project, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	p := &model.Project{}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO project (tenant_id, name, slug) VALUES ($1, $2, $3)
		 RETURNING id, tenant_id, name, slug, created_at`,
		tenantID, name, slug).
		Scan(&p.ID, &p.TenantID, &p.Name, &p.Slug, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}
	return p, nil
}

// ListProjects returns projects for the tenant from context.
func (s *PgStore) ListProjects(ctx context.Context) ([]model.Project, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	rows, err := s.pool.Query(ctx,
		"SELECT id, tenant_id, name, slug, created_at FROM project WHERE tenant_id = $1 ORDER BY name",
		tenantID)
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Slug, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// addProjectFilter appends a project_id filter to a query when allowed projects are set in context.
// Returns the modified query, args, and argN.
func addProjectFilter(ctx context.Context, query string, args []any, argN int, column string) (string, []any, int) {
	projectIDs := auth.AllowedProjectsFromContext(ctx)
	if len(projectIDs) == 0 {
		return query, args, argN
	}
	argN++
	query += fmt.Sprintf(" AND %s = ANY($%d::uuid[])", column, argN)
	args = append(args, projectIDs)
	return query, args, argN
}
