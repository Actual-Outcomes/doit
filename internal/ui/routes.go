package ui

import (
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RegisterUIRoutes mounts the web UI sub-router under /ui/.
func RegisterUIRoutes(r chi.Router, s store.Store, adminKey string, adminTenantID *uuid.UUID) {
	h := NewUIHandlers(s, adminKey, adminKey, adminTenantID)

	r.Route("/ui", func(ui chi.Router) {
		// Login pages — no session required
		ui.Get("/login", h.LoginPage)
		ui.Post("/login", h.LoginSubmit)

		// All other routes require a valid session
		ui.Group(func(protected chi.Router) {
			protected.Use(SessionMiddleware(adminKey))

			protected.Post("/logout", h.Logout)
			protected.Post("/project", h.ProjectSwitch)
			protected.Get("/", h.Dashboard)
			protected.Get("/issues", h.IssueList)
			protected.Get("/issues/{id}", h.IssueDetail)
			protected.Get("/ready", h.ReadyWork)
		})

		// Admin routes — require admin session
		ui.Route("/admin", func(admin chi.Router) {
			admin.Use(AdminSessionMiddleware(adminKey))

			admin.Get("/", h.AdminDashboard)
			admin.Get("/tenants", h.AdminTenants)
			admin.Post("/tenants", h.AdminCreateTenant)
			admin.Get("/tenants/{slug}/keys", h.AdminAPIKeys)
			admin.Post("/tenants/{slug}/keys", h.AdminCreateAPIKey)
			admin.Post("/tenants/{slug}/keys/revoke", h.AdminRevokeAPIKey)
			admin.Get("/projects", h.AdminProjects)
			admin.Post("/projects", h.AdminUpdateProject)
		})
	})
}
