package api

import (
	"github.com/Actual-Outcomes/doit/internal/store"
)

// Handlers wraps the store for MCP tool implementations.
type Handlers struct {
	store store.Store
}

func NewHandlers(s store.Store) *Handlers {
	return &Handlers{store: s}
}
