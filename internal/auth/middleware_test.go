package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// mockResolver implements KeyResolver for testing.
type mockResolver struct {
	keys map[string]uuid.UUID // keyHash -> tenantID
}

func (m *mockResolver) ResolveAPIKey(_ context.Context, keyHash string) (uuid.UUID, error) {
	if id, ok := m.keys[keyHash]; ok {
		return id, nil
	}
	return uuid.Nil, fmt.Errorf("key not found")
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestAPIKeyMiddleware_HealthBypass(t *testing.T) {
	mw := APIKeyMiddleware(MiddlewareConfig{})
	handler := mw(okHandler())

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAPIKeyMiddleware_DocumentationBypass(t *testing.T) {
	mw := APIKeyMiddleware(MiddlewareConfig{})
	handler := mw(okHandler())

	req := httptest.NewRequest("GET", "/documentation", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAPIKeyMiddleware_UIBypass(t *testing.T) {
	mw := APIKeyMiddleware(MiddlewareConfig{})
	handler := mw(okHandler())

	req := httptest.NewRequest("GET", "/ui/graph", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAPIKeyMiddleware_MissingHeader(t *testing.T) {
	mw := APIKeyMiddleware(MiddlewareConfig{})
	handler := mw(okHandler())

	req := httptest.NewRequest("POST", "/mcp", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAPIKeyMiddleware_BadFormat(t *testing.T) {
	mw := APIKeyMiddleware(MiddlewareConfig{})
	handler := mw(okHandler())

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAPIKeyMiddleware_AdminKey(t *testing.T) {
	adminTenantID := uuid.New()
	mw := APIKeyMiddleware(MiddlewareConfig{
		AdminKey:      "secret-admin-key",
		AdminTenantID: &adminTenantID,
	})

	var gotAdmin bool
	var gotTenant uuid.UUID
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotAdmin = IsAdmin(r.Context())
		gotTenant, _ = TenantFromContext(r.Context())
	}))

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer secret-admin-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !gotAdmin {
		t.Error("expected admin context to be set")
	}
	if gotTenant != adminTenantID {
		t.Errorf("tenant = %s, want %s", gotTenant, adminTenantID)
	}
}

func TestAPIKeyMiddleware_TenantKey(t *testing.T) {
	tenantID := uuid.New()
	tenantKey := "tenant-key-123"
	keyHash := HashKey(tenantKey)

	mw := APIKeyMiddleware(MiddlewareConfig{
		AdminKey: "admin-key",
		Resolver: &mockResolver{keys: map[string]uuid.UUID{keyHash: tenantID}},
	})

	var gotAdmin bool
	var gotTenant uuid.UUID
	handler := mw(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotAdmin = IsAdmin(r.Context())
		gotTenant, _ = TenantFromContext(r.Context())
	}))

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+tenantKey)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if gotAdmin {
		t.Error("expected admin to be false for tenant key")
	}
	if gotTenant != tenantID {
		t.Errorf("tenant = %s, want %s", gotTenant, tenantID)
	}
}

func TestAPIKeyMiddleware_InvalidKey(t *testing.T) {
	mw := APIKeyMiddleware(MiddlewareConfig{
		AdminKey: "admin-key",
		Resolver: &mockResolver{keys: map[string]uuid.UUID{}},
	})
	handler := mw(okHandler())

	req := httptest.NewRequest("POST", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer bad-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAdminOnlyMiddleware_Admin(t *testing.T) {
	mw := AdminOnlyMiddleware()
	handler := mw(okHandler())

	req := httptest.NewRequest("POST", "/admin/mcp", nil)
	ctx := WithAdmin(req.Context())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAdminOnlyMiddleware_NonAdmin(t *testing.T) {
	mw := AdminOnlyMiddleware()
	handler := mw(okHandler())

	req := httptest.NewRequest("POST", "/admin/mcp", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

func TestHashKey(t *testing.T) {
	hash := HashKey("test-key")
	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}
	// Same input should produce same hash
	if hash != HashKey("test-key") {
		t.Error("HashKey not deterministic")
	}
	// Different input should produce different hash
	if hash == HashKey("other-key") {
		t.Error("different keys produced same hash")
	}
}
