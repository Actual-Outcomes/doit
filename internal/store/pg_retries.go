package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/jackc/pgx/v5"
)

// GenerateRetryID creates a hash-based retry ID with adaptive length.
func (s *PgStore) GenerateRetryID(ctx context.Context) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	var count int
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM retries").Scan(&count)
	if err != nil {
		return "", fmt.Errorf("counting retries: %w", err)
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
		id := fmt.Sprintf("rty-%s", hexHash[:hashLen])

		var exists bool
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM retries WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("checking retry ID uniqueness: %w", err)
		}
		if !exists {
			return id, nil
		}

		if attempt%10 == 9 && hashLen < 8 {
			hashLen++
		}
	}

	return "", fmt.Errorf("failed to generate unique retry ID after 30 attempts")
}

// RecordRetry inserts a new retry attempt for an issue.
func (s *PgStore) RecordRetry(ctx context.Context, input RecordRetryInput) (*model.Retry, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	id, err := s.GenerateRetryID(ctx)
	if err != nil {
		return nil, err
	}

	// Auto-compute attempt number for this issue
	var attempt int
	err = s.pool.QueryRow(ctx,
		"SELECT COALESCE(MAX(attempt), 0) + 1 FROM retries WHERE issue_id = $1 AND tenant_id = $2",
		input.IssueID, tenantID).Scan(&attempt)
	if err != nil {
		return nil, fmt.Errorf("computing attempt number: %w", err)
	}

	status := input.Status
	if status == "" {
		status = string(model.RetryFailed)
	}

	r := &model.Retry{}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO retries (id, tenant_id, project_id, issue_id, attempt, status, error, agent, started_at, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), $9)
		 RETURNING id, tenant_id, project_id, issue_id, attempt, status, error, agent, started_at, ended_at, created_by`,
		id, tenantID, nullEmpty(input.ProjectID), input.IssueID,
		attempt, status, input.Error, nullEmpty(input.Agent), nullEmpty(input.CreatedBy)).
		Scan(&r.ID, &ns{&r.TenantID}, &ns{&r.ProjectID}, &r.IssueID,
			&r.Attempt, &r.Status, &r.Error, &ns{&r.Agent},
			&r.StartedAt, &r.EndedAt, &ns{&r.CreatedBy})
	if err != nil {
		return nil, fmt.Errorf("recording retry: %w", err)
	}

	return r, nil
}

// ListRetries returns retry attempts for an issue.
func (s *PgStore) ListRetries(ctx context.Context, issueID string, filter model.RetryFilter) ([]model.Retry, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	query := `SELECT id, tenant_id, project_id, issue_id, attempt, status, error, agent, started_at, ended_at, created_by
		FROM retries WHERE tenant_id = $1 AND issue_id = $2`
	args := []any{tenantID, issueID}
	argN := 2

	if filter.Status != nil {
		argN++
		query += fmt.Sprintf(" AND status = $%d", argN)
		args = append(args, string(*filter.Status))
	}

	query += " ORDER BY attempt ASC"

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	argN++
	query += fmt.Sprintf(" LIMIT $%d", argN)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing retries: %w", err)
	}
	defer rows.Close()

	retries := []model.Retry{}
	for rows.Next() {
		r, err := scanRetry(rows)
		if err != nil {
			return nil, err
		}
		retries = append(retries, *r)
	}
	return retries, rows.Err()
}

func scanRetry(rows pgx.Rows) (*model.Retry, error) {
	var r model.Retry
	err := rows.Scan(&r.ID, &ns{&r.TenantID}, &ns{&r.ProjectID}, &r.IssueID,
		&r.Attempt, &r.Status, &r.Error, &ns{&r.Agent},
		&r.StartedAt, &r.EndedAt, &ns{&r.CreatedBy})
	if err != nil {
		return nil, fmt.Errorf("scanning retry: %w", err)
	}
	return &r, nil
}
