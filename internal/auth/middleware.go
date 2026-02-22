package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// KeyResolver looks up a tenant ID by API key hash.
type KeyResolver interface {
	ResolveAPIKey(ctx context.Context, keyHash string) (uuid.UUID, error)
}

// MiddlewareConfig configures the API key authentication middleware.
type MiddlewareConfig struct {
	AdminKey      string
	AdminTenantID *uuid.UUID
	Resolver      KeyResolver
}

// APIKeyMiddleware authenticates requests via Bearer token.
func APIKeyMiddleware(cfg MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/documentation" || strings.HasPrefix(r.URL.Path, "/ui/") {
				next.ServeHTTP(w, r)
				return
			}

			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, "missing Authorization header", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			if token == header {
				http.Error(w, "invalid Authorization format, expected Bearer token", http.StatusUnauthorized)
				return
			}

			// Check admin key
			if cfg.AdminKey != "" && subtle.ConstantTimeCompare([]byte(token), []byte(cfg.AdminKey)) == 1 {
				ctx := WithAdmin(r.Context())
				if cfg.AdminTenantID != nil {
					ctx = WithTenant(ctx, *cfg.AdminTenantID)
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Hash and resolve tenant key
			hash := HashKey(token)
			tenantID, err := cfg.Resolver.ResolveAPIKey(r.Context(), hash)
			if err != nil {
				http.Error(w, "invalid API key", http.StatusUnauthorized)
				return
			}

			ctx := WithTenant(r.Context(), tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// HashKey returns the SHA-256 hex digest of a raw API key.
func HashKey(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
