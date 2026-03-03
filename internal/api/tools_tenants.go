package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NOTE: Admin authorization is enforced by auth.AdminOnlyMiddleware on the
// /admin/mcp route. These handlers no longer check IsAdmin themselves.

type createTenantArgs struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handlers) CreateTenant(ctx context.Context, _ *mcp.CallToolRequest, args createTenantArgs) (*mcp.CallToolResult, any, error) {
	tenant, err := h.store.CreateTenant(ctx, args.Name, args.Slug)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(tenant)
}

type listTenantsArgs struct{}

func (h *Handlers) ListTenants(ctx context.Context, _ *mcp.CallToolRequest, _ listTenantsArgs) (*mcp.CallToolResult, any, error) {
	tenants, err := h.store.ListTenants(ctx)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(tenants)
}

type createAPIKeyArgs struct {
	Tenant string `json:"tenant"`
	Label  string `json:"label"`
}

func (h *Handlers) CreateAPIKey(ctx context.Context, _ *mcp.CallToolRequest, args createAPIKeyArgs) (*mcp.CallToolResult, any, error) {
	// Generate raw key
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return errResult(fmt.Errorf("generating key: %w", err))
	}
	rawKey := hex.EncodeToString(raw)
	prefix := rawKey[:8]
	keyHash := auth.HashKey(rawKey)

	info, err := h.store.CreateAPIKey(ctx, args.Tenant, args.Label, keyHash, prefix)
	if err != nil {
		return errResult(err)
	}

	// Return raw key + info (raw key only shown once)
	result := map[string]any{
		"raw_key": rawKey,
		"info":    info,
	}
	return jsonResult(result)
}

type revokeAPIKeyArgs struct {
	Prefix string `json:"prefix"`
}

func (h *Handlers) RevokeAPIKey(ctx context.Context, _ *mcp.CallToolRequest, args revokeAPIKeyArgs) (*mcp.CallToolResult, any, error) {
	if err := h.store.RevokeAPIKey(ctx, args.Prefix); err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]string{"revoked": args.Prefix})
}

type listAPIKeysArgs struct {
	Tenant string `json:"tenant"`
}

func (h *Handlers) ListAPIKeys(ctx context.Context, _ *mcp.CallToolRequest, args listAPIKeysArgs) (*mcp.CallToolResult, any, error) {
	keys, err := h.store.ListAPIKeys(ctx, args.Tenant)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(keys)
}

type updateTenantArgs struct {
	Tenant string  `json:"tenant"`
	Name   *string `json:"name,omitempty"`
	Slug   *string `json:"slug,omitempty"`
}

func (h *Handlers) UpdateTenant(ctx context.Context, _ *mcp.CallToolRequest, args updateTenantArgs) (*mcp.CallToolResult, any, error) {
	// Resolve tenant slug to ID
	tenants, err := h.store.ListTenants(ctx)
	if err != nil {
		return errResult(err)
	}
	var tenantID string
	for _, t := range tenants {
		if t.Slug == args.Tenant {
			tenantID = t.ID.String()
			break
		}
	}
	if tenantID == "" {
		return errResult(fmt.Errorf("tenant %q not found", args.Tenant))
	}

	tenant, err := h.store.UpdateTenant(ctx, tenantID, args.Name, args.Slug)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(tenant)
}

type rotateAdminKeyArgs struct{}

func (h *Handlers) RotateAdminKey(ctx context.Context, _ *mcp.CallToolRequest, _ rotateAdminKeyArgs) (*mcp.CallToolResult, any, error) {
	// Generate new raw key
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return errResult(fmt.Errorf("generating key: %w", err))
	}
	rawKey := hex.EncodeToString(raw)
	keyHash := auth.HashKey(rawKey)

	// Store hash in config table
	if err := h.store.SetConfig(ctx, "admin_key_hash", keyHash); err != nil {
		return errResult(err)
	}

	result := map[string]string{
		"raw_key": rawKey,
		"message": "Admin key rotated. The old DB-stored key is invalidated. Env var key still works as fallback.",
	}
	return jsonResult(result)
}

type deleteProjectArgs struct {
	Project string `json:"project"`
}

func (h *Handlers) AdminDeleteProject(ctx context.Context, _ *mcp.CallToolRequest, args deleteProjectArgs) (*mcp.CallToolResult, any, error) {
	projectID, err := resolveProjectSlug(ctx, h.store, args.Project)
	if err != nil {
		return errResult(err)
	}
	if err := h.store.DeleteProject(ctx, projectID); err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]string{"deleted": args.Project})
}

type deleteTenantArgs struct {
	Tenant string `json:"tenant"`
}

func (h *Handlers) DeleteTenant(ctx context.Context, _ *mcp.CallToolRequest, args deleteTenantArgs) (*mcp.CallToolResult, any, error) {
	// Resolve tenant slug to ID
	tenants, err := h.store.ListTenants(ctx)
	if err != nil {
		return errResult(err)
	}
	var tenantID string
	for _, t := range tenants {
		if t.Slug == args.Tenant {
			tenantID = t.ID.String()
			break
		}
	}
	if tenantID == "" {
		return errResult(fmt.Errorf("tenant %q not found", args.Tenant))
	}

	if err := h.store.DeleteTenant(ctx, tenantID); err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]string{"deleted": args.Tenant})
}

type updateProjectArgs struct {
	Project string  `json:"project"`
	Name    *string `json:"name,omitempty"`
	Slug    *string `json:"slug,omitempty"`
}

func (h *Handlers) AdminUpdateProject(ctx context.Context, _ *mcp.CallToolRequest, args updateProjectArgs) (*mcp.CallToolResult, any, error) {
	projectID, err := resolveProjectSlug(ctx, h.store, args.Project)
	if err != nil {
		return errResult(err)
	}
	project, err := h.store.UpdateProject(ctx, projectID, args.Name, args.Slug)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(project)
}
