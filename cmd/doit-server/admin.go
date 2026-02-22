package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/jackc/pgx/v5/pgxpool"
)

const recoveryPhrase = "I have no memory of this place! Even Gimly is lost down here!"

func runAdmin(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: doit-server <command> [flags]")
		fmt.Fprintln(os.Stderr, "commands: serve, create-tenant, create-key, revoke-key, list-tenants, list-keys, reset-admin-key")
		os.Exit(1)
	}

	switch args[0] {
	case "create-tenant":
		adminCreateTenant(args[1:])
	case "create-key":
		adminCreateKey(args[1:])
	case "revoke-key":
		adminRevokeKey(args[1:])
	case "list-tenants":
		adminListTenants()
	case "list-keys":
		adminListKeys(args[1:])
	case "reset-admin-key":
		adminResetKey(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		os.Exit(1)
	}
}

func mustPool() *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	return pool
}

func parseFlags(args []string) map[string]string {
	flags := map[string]string{}
	for i := 0; i < len(args)-1; i += 2 {
		key := args[i]
		if len(key) > 2 && key[:2] == "--" {
			flags[key[2:]] = args[i+1]
		}
	}
	return flags
}

func adminCreateTenant(args []string) {
	flags := parseFlags(args)
	name := flags["name"]
	slug := flags["slug"]
	if name == "" || slug == "" {
		fmt.Fprintln(os.Stderr, "usage: create-tenant --name <name> --slug <slug>")
		os.Exit(1)
	}

	pool := mustPool()
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var id string
	err := pool.QueryRow(ctx,
		`INSERT INTO tenant (name, slug) VALUES ($1, $2) RETURNING id`, name, slug).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create tenant: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("tenant created: id=%s slug=%s\n", id, slug)
}

func adminCreateKey(args []string) {
	flags := parseFlags(args)
	tenantSlug := flags["tenant"]
	label := flags["label"]
	if tenantSlug == "" {
		fmt.Fprintln(os.Stderr, "usage: create-key --tenant <slug> [--label <label>]")
		os.Exit(1)
	}

	pool := mustPool()
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var tenantID string
	err := pool.QueryRow(ctx, `SELECT id FROM tenant WHERE slug = $1`, tenantSlug).Scan(&tenantID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tenant %q not found: %v\n", tenantSlug, err)
		os.Exit(1)
	}

	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate key: %v\n", err)
		os.Exit(1)
	}
	rawKey := hex.EncodeToString(rawBytes)
	prefix := rawKey[:8]
	keyHash := auth.HashKey(rawKey)

	var id string
	err = pool.QueryRow(ctx,
		`INSERT INTO api_key (tenant_id, key_hash, prefix, label) VALUES ($1, $2, $3, $4) RETURNING id`,
		tenantID, keyHash, prefix, label).Scan(&id)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create API key: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("API key created for tenant %q:\n", tenantSlug)
	fmt.Printf("  id:     %s\n", id)
	fmt.Printf("  prefix: %s\n", prefix)
	fmt.Printf("  key:    %s\n", rawKey)
	fmt.Println("\nSave this key now — it cannot be retrieved again.")
}

func adminRevokeKey(args []string) {
	flags := parseFlags(args)
	prefix := flags["prefix"]
	if prefix == "" {
		fmt.Fprintln(os.Stderr, "usage: revoke-key --prefix <8-char-prefix>")
		os.Exit(1)
	}

	pool := mustPool()
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := pool.Exec(ctx,
		`UPDATE api_key SET revoked_at = now() WHERE prefix = $1 AND revoked_at IS NULL`, prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to revoke key: %v\n", err)
		os.Exit(1)
	}
	if tag.RowsAffected() == 0 {
		fmt.Fprintln(os.Stderr, "no active key found with that prefix")
		os.Exit(1)
	}
	fmt.Printf("key with prefix %s revoked\n", prefix)
}

func adminListTenants() {
	pool := mustPool()
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx, `SELECT id, slug, name, created_at FROM tenant ORDER BY created_at`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list tenants: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Printf("%-36s  %-20s  %-30s  %s\n", "ID", "SLUG", "NAME", "CREATED")
	for rows.Next() {
		var id, slug, name string
		var createdAt time.Time
		if err := rows.Scan(&id, &slug, &name, &createdAt); err != nil {
			fmt.Fprintf(os.Stderr, "failed to scan row: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%-36s  %-20s  %-30s  %s\n", id, slug, name, createdAt.Format(time.RFC3339))
	}
}

func adminListKeys(args []string) {
	flags := parseFlags(args)
	tenantSlug := flags["tenant"]
	if tenantSlug == "" {
		fmt.Fprintln(os.Stderr, "usage: list-keys --tenant <slug>")
		os.Exit(1)
	}

	pool := mustPool()
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := pool.Query(ctx,
		`SELECT ak.id, ak.prefix, ak.label, ak.created_at, ak.revoked_at
		 FROM api_key ak
		 JOIN tenant t ON t.id = ak.tenant_id
		 WHERE t.slug = $1
		 ORDER BY ak.created_at`, tenantSlug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list keys: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Printf("%-36s  %-8s  %-20s  %-25s  %s\n", "ID", "PREFIX", "LABEL", "CREATED", "REVOKED")
	for rows.Next() {
		var id, prefix, label string
		var createdAt time.Time
		var revokedAt *time.Time
		if err := rows.Scan(&id, &prefix, &label, &createdAt, &revokedAt); err != nil {
			fmt.Fprintf(os.Stderr, "failed to scan row: %v\n", err)
			os.Exit(1)
		}
		revoked := "-"
		if revokedAt != nil {
			revoked = revokedAt.Format(time.RFC3339)
		}
		fmt.Printf("%-36s  %-8s  %-20s  %-25s  %s\n", id, prefix, label, createdAt.Format(time.RFC3339), revoked)
	}
}

func adminResetKey(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, `usage: reset-admin-key "<recovery phrase>" <new-admin-key>`)
		os.Exit(1)
	}

	phrase := args[0]
	newKey := args[1]

	if phrase != recoveryPhrase {
		fmt.Fprintln(os.Stderr, "You shall not pass!")
		os.Exit(1)
	}

	if len(newKey) < 32 {
		fmt.Fprintln(os.Stderr, "admin key must be at least 32 characters")
		os.Exit(1)
	}

	envPath := ".env"
	var lines []string
	replaced := false

	if f, err := os.Open(envPath); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "API_KEY=") {
				lines = append(lines, "API_KEY="+newKey)
				replaced = true
			} else {
				lines = append(lines, line)
			}
		}
		f.Close()
	}

	if !replaced {
		lines = append(lines, "API_KEY="+newKey)
	}

	if err := os.WriteFile(envPath, []byte(strings.Join(lines, "\n")+"\n"), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write .env: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Admin key updated in .env")
	fmt.Println("Restart the server for the change to take effect.")
}
