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

// GenerateLessonID creates a hash-based lesson ID with adaptive length.
func (s *PgStore) GenerateLessonID(ctx context.Context) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	var count int
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM lessons").Scan(&count)
	if err != nil {
		return "", fmt.Errorf("counting lessons: %w", err)
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
		id := fmt.Sprintf("lsn-%s", hexHash[:hashLen])

		var exists bool
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM lessons WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("checking lesson ID uniqueness: %w", err)
		}
		if !exists {
			return id, nil
		}

		if attempt%10 == 9 && hashLen < 8 {
			hashLen++
		}
	}

	return "", fmt.Errorf("failed to generate unique lesson ID after 30 attempts")
}

// RecordLesson inserts a new lesson learned.
func (s *PgStore) RecordLesson(ctx context.Context, input RecordLessonInput) (*model.Lesson, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	id, err := s.GenerateLessonID(ctx)
	if err != nil {
		return nil, err
	}

	severity := input.Severity
	if severity == 0 {
		severity = 2
	}

	components := input.Components
	if components == nil {
		components = []string{}
	}

	l := &model.Lesson{}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO lessons (id, tenant_id, project_id, issue_id, title, mistake, correction,
		 expert, components, severity, status, created_at, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'open', NOW(), $11)
		 RETURNING id, tenant_id, project_id, issue_id, title, mistake, correction,
		 expert, components, severity, status, created_at, created_by, resolved_at, resolved_by`,
		id, tenantID, nullEmpty(input.ProjectID), nullEmpty(input.IssueID),
		input.Title, input.Mistake, input.Correction,
		nullEmpty(input.Expert), components, severity, nullEmpty(input.CreatedBy)).
		Scan(&l.ID, &l.TenantID, &ns{&l.ProjectID}, &ns{&l.IssueID},
			&l.Title, &l.Mistake, &l.Correction,
			&ns{&l.Expert}, &l.Components, &l.Severity, &l.Status,
			&l.CreatedAt, &ns{&l.CreatedBy}, &l.ResolvedAt, &ns{&l.ResolvedBy})
	if err != nil {
		return nil, fmt.Errorf("recording lesson: %w", err)
	}

	return l, nil
}

// ListLessons returns lessons matching the filter.
func (s *PgStore) ListLessons(ctx context.Context, filter model.LessonFilter) ([]model.Lesson, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	query := `SELECT id, tenant_id, project_id, issue_id, title, mistake, correction,
		expert, components, severity, status, created_at, created_by, resolved_at, resolved_by
		FROM lessons WHERE tenant_id = $1`
	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}
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
	if filter.Expert != nil {
		argN++
		query += fmt.Sprintf(" AND expert = $%d", argN)
		args = append(args, *filter.Expert)
	}
	if filter.Component != nil {
		argN++
		query += fmt.Sprintf(" AND $%d = ANY(components)", argN)
		args = append(args, *filter.Component)
	}
	if filter.Severity != nil {
		argN++
		query += fmt.Sprintf(" AND severity = $%d", argN)
		args = append(args, *filter.Severity)
	}

	// Project filter (allowed projects)
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
		return nil, fmt.Errorf("listing lessons: %w", err)
	}
	defer rows.Close()

	lessons := []model.Lesson{}
	for rows.Next() {
		l, err := scanLesson(rows)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, *l)
	}
	return lessons, rows.Err()
}

// ResolveLesson marks a lesson as resolved.
func (s *PgStore) ResolveLesson(ctx context.Context, id string, resolvedBy string) (*model.Lesson, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tenantID, ok := auth.TenantFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no tenant in context")
	}

	l := &model.Lesson{}
	err := s.pool.QueryRow(ctx,
		`UPDATE lessons SET status = 'resolved', resolved_at = NOW(), resolved_by = $1
		 WHERE id = $2 AND tenant_id = $3
		 RETURNING id, tenant_id, project_id, issue_id, title, mistake, correction,
		 expert, components, severity, status, created_at, created_by, resolved_at, resolved_by`,
		nullEmpty(resolvedBy), id, tenantID).
		Scan(&l.ID, &l.TenantID, &ns{&l.ProjectID}, &ns{&l.IssueID},
			&l.Title, &l.Mistake, &l.Correction,
			&ns{&l.Expert}, &l.Components, &l.Severity, &l.Status,
			&l.CreatedAt, &ns{&l.CreatedBy}, &l.ResolvedAt, &ns{&l.ResolvedBy})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("lesson %s not found", id)
		}
		return nil, fmt.Errorf("resolving lesson: %w", err)
	}

	return l, nil
}

func scanLesson(rows pgx.Rows) (*model.Lesson, error) {
	var l model.Lesson
	err := rows.Scan(&l.ID, &l.TenantID, &ns{&l.ProjectID}, &ns{&l.IssueID},
		&l.Title, &l.Mistake, &l.Correction,
		&ns{&l.Expert}, &l.Components, &l.Severity, &l.Status,
		&l.CreatedAt, &ns{&l.CreatedBy}, &l.ResolvedAt, &ns{&l.ResolvedBy})
	if err != nil {
		return nil, fmt.Errorf("scanning lesson: %w", err)
	}
	return &l, nil
}
