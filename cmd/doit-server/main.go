package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Actual-Outcomes/doit/internal/api"
	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/Actual-Outcomes/doit/internal/config"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/Actual-Outcomes/doit/internal/ui"
	"github.com/Actual-Outcomes/doit/internal/version"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Subcommand routing: default to "serve" if no args or first arg is a flag
	cmd := "serve"
	if len(os.Args) > 1 && os.Args[1] != "" && os.Args[1][0] != '-' {
		cmd = os.Args[1]
	}

	if cmd != "serve" {
		runAdmin(os.Args[1:])
		return
	}

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	setupLogging(cfg.LogLevel)

	// Run database migrations before anything else
	if err := store.RunMigrations(cfg.DatabaseURL); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	pgStore, err := store.NewPgStore(ctx, cfg.DatabaseURL, cfg.DBQueryTimeout, cfg.IDPrefix)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pgStore.Close()

	// MCP server
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "doit-mcp",
		Version: version.Number,
	}, nil)

	handlers := api.NewHandlers(pgStore)
	api.RegisterTools(mcpServer, handlers)

	mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return mcpServer
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	// Auth config
	authCfg := auth.MiddlewareConfig{
		AdminKey: cfg.AdminAPIKey,
		Resolver: pgStore,
	}
	if cfg.AdminTenantSlug != "" {
		tenants, err := pgStore.ListTenants(ctx)
		if err != nil {
			slog.Error("failed to list tenants", "error", err)
			os.Exit(1)
		}
		for _, t := range tenants {
			if t.Slug == cfg.AdminTenantSlug {
				id := t.ID
				authCfg.AdminTenantID = &id
				slog.Info("admin key bound to tenant", "tenant", t.Slug, "id", t.ID)
				break
			}
		}
	}

	// HTTP router
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.HTTPTimeout))
	r.Use(auth.APIKeyMiddleware(authCfg))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	r.Handle("/mcp", mcpHandler)

	r.Get("/documentation", api.DocumentationHandler())

	ui.RegisterUIRoutes(r, pgStore, cfg.AdminAPIKey, authCfg.AdminTenantID)

	// Start server
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("starting doit-mcp server", "port", cfg.Port, "version", version.Number)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("server stopped")
}

func setupLogging(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(handler))
}
