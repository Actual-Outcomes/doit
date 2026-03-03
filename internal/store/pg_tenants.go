package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/google/uuid"
)

// ResolveAPIKey looks up an active API key by its SHA-256 hash and returns the tenant ID.
func (s *PgStore) ResolveAPIKey(ctx context.Context, keyHash string) (uuid.UUID, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	var tenantID uuid.UUID
	err := s.pool.QueryRow(ctx,
		`SELECT ak.tenant_id FROM api_key ak
		 JOIN tenant t ON t.id = ak.tenant_id
		 WHERE ak.key_hash = $1 AND ak.revoked_at IS NULL`, keyHash).Scan(&tenantID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolving API key: %w", err)
	}
	return tenantID, nil
}

// CreateTenant creates a new tenant.
func (s *PgStore) CreateTenant(ctx context.Context, name, slug string) (*model.Tenant, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	t := &model.Tenant{}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tenant (name, slug) VALUES ($1, $2)
		 RETURNING id, name, slug, created_at`, name, slug).
		Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating tenant: %w", err)
	}
	return t, nil
}

// UpdateTenant updates a tenant's name and/or slug by tenant ID.
func (s *PgStore) UpdateTenant(ctx context.Context, tenantID string, name, slug *string) (*model.Tenant, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	if name == nil && slug == nil {
		return nil, fmt.Errorf("nothing to update: provide name or slug")
	}

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

	query := fmt.Sprintf(
		"UPDATE tenant SET %s WHERE id = $%d RETURNING id, name, slug, created_at",
		strings.Join(sets, ", "), argN,
	)

	t := &model.Tenant{}
	err := s.pool.QueryRow(ctx, query, args...).
		Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("updating tenant: %w", err)
	}
	return t, nil
}

// ListTenants returns all tenants.
func (s *PgStore) ListTenants(ctx context.Context) ([]model.Tenant, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	rows, err := s.pool.Query(ctx,
		"SELECT id, name, slug, created_at FROM tenant ORDER BY created_at")
	if err != nil {
		return nil, fmt.Errorf("listing tenants: %w", err)
	}
	defer rows.Close()

	var tenants []model.Tenant
	for rows.Next() {
		var t model.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning tenant: %w", err)
		}
		tenants = append(tenants, t)
	}
	return tenants, rows.Err()
}

// DeleteTenant deletes a tenant by ID. Rejects if any projects still exist.
// Cascades to API keys.
func (s *PgStore) DeleteTenant(ctx context.Context, tenantID string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	// Check for existing projects
	var count int
	err := s.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM project WHERE tenant_id = $1", tenantID).Scan(&count)
	if err != nil {
		return fmt.Errorf("checking tenant projects: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot delete tenant: %d projects still exist (delete them first)", count)
	}

	// Delete API keys first (FK constraint)
	_, err = s.pool.Exec(ctx, "DELETE FROM api_key WHERE tenant_id = $1", tenantID)
	if err != nil {
		return fmt.Errorf("deleting tenant API keys: %w", err)
	}

	tag, err := s.pool.Exec(ctx, "DELETE FROM tenant WHERE id = $1", tenantID)
	if err != nil {
		return fmt.Errorf("deleting tenant: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found")
	}
	return nil
}

// CreateAPIKey creates a new API key for a tenant. Returns the key info (not the raw key).
func (s *PgStore) CreateAPIKey(ctx context.Context, tenantSlug, label, keyHash, prefix string) (*model.APIKeyInfo, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	// Resolve tenant slug to ID
	var tenantID uuid.UUID
	err := s.pool.QueryRow(ctx, "SELECT id FROM tenant WHERE slug = $1", tenantSlug).Scan(&tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant %q not found: %w", tenantSlug, err)
	}

	k := &model.APIKeyInfo{}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO api_key (tenant_id, key_hash, prefix, label)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, tenant_id, prefix, label, created_at`,
		tenantID, keyHash, prefix, label).
		Scan(&k.ID, &k.TenantID, &k.Prefix, &k.Label, &k.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating API key: %w", err)
	}
	return k, nil
}

// RevokeAPIKey revokes an API key by prefix.
func (s *PgStore) RevokeAPIKey(ctx context.Context, prefix string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tag, err := s.pool.Exec(ctx,
		"UPDATE api_key SET revoked_at = NOW() WHERE prefix = $1 AND revoked_at IS NULL", prefix)
	if err != nil {
		return fmt.Errorf("revoking API key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("API key with prefix %q not found or already revoked", prefix)
	}
	return nil
}

// ListAPIKeys lists API keys for a tenant.
func (s *PgStore) ListAPIKeys(ctx context.Context, tenantSlug string) ([]model.APIKeyInfo, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	rows, err := s.pool.Query(ctx,
		`SELECT ak.id, ak.tenant_id, ak.prefix, ak.label, ak.created_at, ak.revoked_at
		 FROM api_key ak
		 JOIN tenant t ON t.id = ak.tenant_id
		 WHERE t.slug = $1
		 ORDER BY ak.created_at`, tenantSlug)
	if err != nil {
		return nil, fmt.Errorf("listing API keys: %w", err)
	}
	defer rows.Close()

	var keys []model.APIKeyInfo
	for rows.Next() {
		var k model.APIKeyInfo
		if err := rows.Scan(&k.ID, &k.TenantID, &k.Prefix, &k.Label, &k.CreatedAt, &k.RevokedAt); err != nil {
			return nil, fmt.Errorf("scanning API key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}
