package auth

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey int

const (
	ctxTenantID ctxKey = iota
	ctxAdmin
	ctxAllowedProjects
)

// WithTenant stores the tenant ID in the context.
func WithTenant(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, ctxTenantID, id)
}

// WithAdmin marks the context as admin (no tenant scoping).
func WithAdmin(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxAdmin, true)
}

// TenantFromContext returns the tenant ID if set.
func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxTenantID).(uuid.UUID)
	return id, ok
}

// IsAdmin returns true if the context is marked as admin.
func IsAdmin(ctx context.Context) bool {
	v, _ := ctx.Value(ctxAdmin).(bool)
	return v
}

// WithAllowedProjects stores the allowed project IDs in the context.
func WithAllowedProjects(ctx context.Context, projectIDs []string) context.Context {
	return context.WithValue(ctx, ctxAllowedProjects, projectIDs)
}

// AllowedProjectsFromContext returns the allowed project IDs if set.
func AllowedProjectsFromContext(ctx context.Context) []string {
	ids, _ := ctx.Value(ctxAllowedProjects).([]string)
	return ids
}
