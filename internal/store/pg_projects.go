package store

import (
	"context"
	"fmt"
	"strings"

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

// ListAllProjects returns all projects across all tenants. Admin use only.
func (s *PgStore) ListAllProjects(ctx context.Context) ([]model.Project, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	rows, err := s.pool.Query(ctx,
		"SELECT id, tenant_id, name, slug, created_at FROM project ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("listing all projects: %w", err)
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

// GetProjectBySlug returns a single project by its slug within the tenant.
func (s *PgStore) GetProjectBySlug(ctx context.Context, slug string) (*model.Project, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	p := &model.Project{}
	err := s.pool.QueryRow(ctx,
		"SELECT id, tenant_id, name, slug, created_at FROM project WHERE tenant_id = $1 AND slug = $2",
		tenantID, slug).
		Scan(&p.ID, &p.TenantID, &p.Name, &p.Slug, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("project not found for slug %q: %w", slug, err)
	}
	return p, nil
}

// UpdateProject updates a project's name and/or slug by project ID within the tenant.
func (s *PgStore) UpdateProject(ctx context.Context, projectID string, name, slug *string) (*model.Project, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	if name == nil && slug == nil {
		return nil, fmt.Errorf("nothing to update: provide name or slug")
	}

	// Build dynamic SET clause
	sets := []string{}
	args := []any{}
	argN := 0

	if name != nil {
		argN++
		sets = append(sets, fmt.Sprintf("name = $%d", argN))
		args = append(args, *name)
	}
	if slug != nil {
		argN++
		sets = append(sets, fmt.Sprintf("slug = $%d", argN))
		args = append(args, *slug)
	}

	argN++
	args = append(args, tenantID)
	argN++
	args = append(args, projectID)

	query := fmt.Sprintf(
		"UPDATE project SET %s WHERE tenant_id = $%d AND id = $%d RETURNING id, tenant_id, name, slug, created_at",
		strings.Join(sets, ", "), argN-1, argN,
	)

	p := &model.Project{}
	err := s.pool.QueryRow(ctx, query, args...).
		Scan(&p.ID, &p.TenantID, &p.Name, &p.Slug, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating project: %w", err)
	}
	return p, nil
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
