package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createTenantArgs struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (h *Handlers) CreateTenant(ctx context.Context, _ *mcp.CallToolRequest, args createTenantArgs) (*mcp.CallToolResult, any, error) {
	if !auth.IsAdmin(ctx) {
		return errResult(fmt.Errorf("admin access required"))
	}
	tenant, err := h.store.CreateTenant(ctx, args.Name, args.Slug)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(tenant)
}

type listTenantsArgs struct{}

func (h *Handlers) ListTenants(ctx context.Context, _ *mcp.CallToolRequest, _ listTenantsArgs) (*mcp.CallToolResult, any, error) {
	if !auth.IsAdmin(ctx) {
		return errResult(fmt.Errorf("admin access required"))
	}
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
	if !auth.IsAdmin(ctx) {
		return errResult(fmt.Errorf("admin access required"))
	}

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
	if !auth.IsAdmin(ctx) {
		return errResult(fmt.Errorf("admin access required"))
	}
	if err := h.store.RevokeAPIKey(ctx, args.Prefix); err != nil {
		return errResult(err)
	}
	return jsonResult(map[string]string{"revoked": args.Prefix})
}

type listAPIKeysArgs struct {
	Tenant string `json:"tenant"`
}

func (h *Handlers) ListAPIKeys(ctx context.Context, _ *mcp.CallToolRequest, args listAPIKeysArgs) (*mcp.CallToolResult, any, error) {
	if !auth.IsAdmin(ctx) {
		return errResult(fmt.Errorf("admin access required"))
	}
	keys, err := h.store.ListAPIKeys(ctx, args.Tenant)
	if err != nil {
		return errResult(err)
	}
	return jsonResult(keys)
}
