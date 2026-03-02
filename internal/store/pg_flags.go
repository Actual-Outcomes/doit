package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/jackc/pgx/v5"
)

// GenerateFlagID creates a hash-based flag ID with adaptive length.
func (s *PgStore) GenerateFlagID(ctx context.Context) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	var count int
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM flags").Scan(&count)
	if err != nil {
		return "", fmt.Errorf("counting flags: %w", err)
	}

	hashLen := 3
	switch {
	case count > 1500:
		hashLen = 6
	case count > 500:
		hashLen = 5
	case count > 100:
		hashLen = 4
	}

	for attempt := 0; attempt < 30; attempt++ {
		seed := fmt.Sprintf("%d-%d-%d", time.Now().UnixNano(), rand.Int63(), attempt)
		hash := sha256.Sum256([]byte(seed))
		hexHash := hex.EncodeToString(hash[:])
		id := fmt.Sprintf("flg-%s", hexHash[:hashLen])

		var exists bool
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM flags WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("checking flag ID uniqueness: %w", err)
		}
		if !exists {
			return id, nil
		}

		if attempt%10 == 9 && hashLen < 8 {
			hashLen++
		}
	}

	return "", fmt.Errorf("failed to generate unique flag ID after 30 attempts")
}

// RaiseFlag inserts a new flag tied to an issue.
func (s *PgStore) RaiseFlag(ctx context.Context, input RaiseFlagInput) (*model.Flag, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	id, err := s.GenerateFlagID(ctx)
	if err != nil {
		return nil, err
	}

	severity := input.Severity
	if severity == 0 {
		severity = 2
	}

	ctxJSON := input.Context
	if ctxJSON == nil {
		ctxJSON = json.RawMessage(`{}`)
	}

	f := &model.Flag{}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO flags (id, tenant_id, project_id, issue_id, type, severity, summary,
		 context, status, created_at, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'open', NOW(), $9)
		 RETURNING id, tenant_id, project_id, issue_id, type, severity, summary,
		 context, status, resolution, resolved_by, resolved_at, created_at, created_by`,
		id, tenantID, nullEmpty(input.ProjectID), nullEmpty(input.IssueID),
		input.Type, severity, input.Summary,
		ctxJSON, nullEmpty(input.CreatedBy)).
		Scan(&f.ID, &f.TenantID, &ns{&f.ProjectID}, &ns{&f.IssueID},
			&f.Type, &f.Severity, &f.Summary,
			&f.Context, &f.Status, &ns{&f.Resolution}, &ns{&f.ResolvedBy},
			&f.ResolvedAt, &f.CreatedAt, &ns{&f.CreatedBy})
	if err != nil {
		return nil, fmt.Errorf("raising flag: %w", err)
	}

	return f, nil
}

// ListFlags returns flags matching the filter.
func (s *PgStore) ListFlags(ctx context.Context, filter model.FlagFilter) ([]model.Flag, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	query := `SELECT id, tenant_id, project_id, issue_id, type, severity, summary,
		context, status, resolution, resolved_by, resolved_at, created_at, created_by
		FROM flags WHERE tenant_id = $1`
	args := []any{tenantID}
	argN := 1

	if filter.ProjectID != nil {
		argN++
		query += fmt.Sprintf(" AND project_id = $%d::uuid", argN)
		args = append(args, *filter.ProjectID)
	}
	if filter.Status != nil {
		argN++
		query += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, string(*filter.Status))
	}
	if filter.Severity != nil {
		argN++
		query += fmt.Sprintf(" AND severity = $%d", argN)
		args = append(args, *filter.Severity)
	}
	if filter.IssueID != nil {
		argN++
		query += fmt.Sprintf(" AND issue_id = $%d", argN)
		args = append(args, *filter.IssueID)
	}

	// Allowed projects filter
	projectIDs := auth.AllowedProjectsFromContext(ctx)
	if len(projectIDs) > 0 {
		argN++
		query += fmt.Sprintf(" AND project_id = ANY($%d::uuid[])", argN)
		args = append(args, projectIDs)
	}

	query += " ORDER BY severity ASC, created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	argN++
	query += fmt.Sprintf(" LIMIT $%d", argN)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing flags: %w", err)
	}
	defer rows.Close()

	flags := []model.Flag{}
	for rows.Next() {
		f, err := scanFlag(rows)
		if err != nil {
			return nil, err
		}
		flags = append(flags, *f)
	}
	return flags, rows.Err()
}

// ResolveFlag marks a flag as resolved with a resolution message.
func (s *PgStore) ResolveFlag(ctx context.Context, id string, resolution, resolvedBy string) (*model.Flag, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	f := &model.Flag{}
	err := s.pool.QueryRow(ctx,
		`UPDATE flags SET status = 'resolved', resolution = $1, resolved_at = NOW(), resolved_by = $2
		 WHERE id = $3 AND tenant_id = $4
		 RETURNING id, tenant_id, project_id, issue_id, type, severity, summary,
		 context, status, resolution, resolved_by, resolved_at, created_at, created_by`,
		resolution, nullEmpty(resolvedBy), id, tenantID).
		Scan(&f.ID, &f.TenantID, &ns{&f.ProjectID}, &ns{&f.IssueID},
			&f.Type, &f.Severity, &f.Summary,
			&f.Context, &f.Status, &ns{&f.Resolution}, &ns{&f.ResolvedBy},
			&f.ResolvedAt, &f.CreatedAt, &ns{&f.CreatedBy})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("flag %s not found", id)
		}
		return nil, fmt.Errorf("resolving flag: %w", err)
	}

	return f, nil
}

func scanFlag(rows pgx.Rows) (*model.Flag, error) {
	var f model.Flag
	err := rows.Scan(&f.ID, &f.TenantID, &ns{&f.ProjectID}, &ns{&f.IssueID},
		&f.Type, &f.Severity, &f.Summary,
		&f.Context, &f.Status, &ns{&f.Resolution}, &ns{&f.ResolvedBy},
		&f.ResolvedAt, &f.CreatedAt, &ns{&f.CreatedBy})
	if err != nil {
		return nil, fmt.Errorf("scanning flag: %w", err)
	}
	return &f, nil
}
