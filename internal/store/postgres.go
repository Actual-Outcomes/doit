package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgStore implements Store against PostgreSQL.
type PgStore struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
	idPrefix     string // e.g. "doit" — configurable per instance
}

func NewPgStore(ctx context.Context, databaseURL string, queryTimeout time.Duration, idPrefix string) (*PgStore, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	if idPrefix == "" {
		idPrefix = "doit"
	}

	return &PgStore{pool: pool, queryTimeout: queryTimeout, idPrefix: idPrefix}, nil
}

func (s *PgStore) Close() { s.pool.Close() }

func (s *PgStore) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, s.queryTimeout)
}

// GenerateID creates a hash-based issue ID with adaptive length.
// Uses SHA-256 of timestamp+random, truncated to avoid collisions.
func (s *PgStore) GenerateID(ctx context.Context, prefix string) (string, error) {
	if prefix == "" {
		prefix = s.idPrefix
	}
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	// Count existing issues to determine hash length
	var count int
	err := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM issues").Scan(&count)
	if err != nil {
		return "", fmt.Errorf("counting issues: %w", err)
	}

	hashLen := 3 // minimum
	switch {
	case count > 1500:
		hashLen = 6
	case count > 500:
		hashLen = 5
	case count > 100:
		hashLen = 4
	}

	// Try to generate a unique ID
	for attempt := 0; attempt < 30; attempt++ {
		seed := fmt.Sprintf("%d-%d-%d", time.Now().UnixNano(), rand.Int63(), attempt)
		hash := sha256.Sum256([]byte(seed))
		hexHash := hex.EncodeToString(hash[:])
		id := fmt.Sprintf("%s-%s", prefix, hexHash[:hashLen])

		// Check uniqueness
		var exists bool
		err := s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM issues WHERE id = $1)", id).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("checking ID uniqueness: %w", err)
		}
		if !exists {
			return id, nil
		}

		// Increase length on collision
		if attempt%10 == 9 && hashLen < 8 {
			hashLen++
		}
	}

	return "", fmt.Errorf("failed to generate unique ID after 30 attempts")
}

// NextChildID returns the next hierarchical child ID for a parent.
// e.g. parent "doit-abc" → "doit-abc.1", "doit-abc.2", etc.
func (s *PgStore) NextChildID(ctx context.Context, parentID string) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var lastChild int
	err = tx.QueryRow(ctx,
		`INSERT INTO child_counters (parent_id, last_child)
		 VALUES ($1, 1)
		 ON CONFLICT (parent_id) DO UPDATE SET last_child = child_counters.last_child + 1
		 RETURNING last_child`, parentID).Scan(&lastChild)
	if err != nil {
		return "", fmt.Errorf("incrementing child counter: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("committing: %w", err)
	}

	return fmt.Sprintf("%s.%d", parentID, lastChild), nil
}

// CreateIssue inserts a new issue and optionally creates a parent-child dependency.
func (s *PgStore) CreateIssue(ctx context.Context, input CreateIssueInput) (*model.Issue, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	now := time.Now().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	issue := &model.Issue{
		ID:        input.ID,
		Title:     input.Title,
		Description: input.Description,
		Design:    input.Design,
		AcceptanceCriteria: input.AcceptanceCriteria,
		Notes:     input.Notes,
		Status:    input.Status,
		Priority:  input.Priority,
		IssueType: input.IssueType,
		Assignee:  input.Assignee,
		Owner:     input.Owner,
		CreatedAt: now,
		CreatedBy: input.CreatedBy,
		UpdatedAt: now,
		Ephemeral: input.Ephemeral,
		MolType:   input.MolType,
		WorkType:  input.WorkType,
		WispType:  input.WispType,
	}

	// Compute content hash
	issue.ContentHash = contentHash(issue)

	_, err = tx.Exec(ctx,
		`INSERT INTO issues (id, content_hash, title, description, design, acceptance_criteria,
		 notes, status, priority, issue_type, assignee, owner, created_at, created_by,
		 updated_at, ephemeral, mol_type, work_type, wisp_type, project_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		issue.ID, issue.ContentHash, issue.Title, issue.Description, issue.Design,
		issue.AcceptanceCriteria, issue.Notes, issue.Status, issue.Priority,
		issue.IssueType, nullEmpty(issue.Assignee), nullEmpty(issue.Owner),
		issue.CreatedAt, nullEmpty(issue.CreatedBy), issue.UpdatedAt,
		issue.Ephemeral, nullEmpty(string(issue.MolType)),
		nullEmpty(string(issue.WorkType)), nullEmpty(string(issue.WispType)),
		nullEmpty(input.ProjectID))
	if err != nil {
		return nil, fmt.Errorf("inserting issue: %w", err)
	}

	// Create parent-child dependency if parent specified
	if input.ParentID != "" {
		_, err = tx.Exec(ctx,
			`INSERT INTO dependencies (issue_id, depends_on_id, type, created_at, created_by)
			 VALUES ($1, $2, 'parent-child', $3, $4)`,
			input.ID, input.ParentID, now, nullEmpty(input.CreatedBy))
		if err != nil {
			return nil, fmt.Errorf("creating parent-child dependency: %w", err)
		}
		issue.ParentID = input.ParentID
	}

	// Add labels
	for _, label := range input.Labels {
		_, err = tx.Exec(ctx,
			`INSERT INTO labels (issue_id, label) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			input.ID, label)
		if err != nil {
			return nil, fmt.Errorf("adding label: %w", err)
		}
	}
	issue.Labels = input.Labels

	// Record creation event
	_, err = tx.Exec(ctx,
		`INSERT INTO events (issue_id, event_type, actor, new_value, created_at)
		 VALUES ($1, 'created', $2, $3, $4)`,
		issue.ID, nullEmpty(input.CreatedBy), issue.Title, now)
	if err != nil {
		return nil, fmt.Errorf("recording creation event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return issue, nil
}

// GetIssue retrieves an issue by ID with labels, dependencies, and parent.
func (s *PgStore) GetIssue(ctx context.Context, id string) (*model.Issue, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	issue, err := s.scanIssue(ctx, s.pool,
		`SELECT `+issueColumns+` FROM issues WHERE id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("getting issue %s: %w", id, err)
	}

	// Load labels
	issue.Labels, err = s.queryLabels(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load parent ID
	var parentID *string
	err = s.pool.QueryRow(ctx,
		`SELECT depends_on_id FROM dependencies WHERE issue_id = $1 AND type = 'parent-child'`, id).
		Scan(&parentID)
	if err == nil && parentID != nil {
		issue.ParentID = *parentID
	}

	return issue, nil
}

// UpdateIssue applies partial updates to an issue.
func (s *PgStore) UpdateIssue(ctx context.Context, id string, input UpdateIssueInput) (*model.Issue, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Build dynamic SET clause
	sets := []string{"updated_at = NOW()"}
	args := []any{}
	argN := 0

	addSet := func(col string, val any) {
		argN++
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, val)
	}

	if input.Title != nil {
		addSet("title", *input.Title)
	}
	if input.Description != nil {
		addSet("description", *input.Description)
	}
	if input.Design != nil {
		addSet("design", *input.Design)
	}
	if input.AcceptanceCriteria != nil {
		addSet("acceptance_criteria", *input.AcceptanceCriteria)
	}
	if input.Notes != nil {
		addSet("notes", *input.Notes)
	}
	if input.Status != nil {
		addSet("status", string(*input.Status))
		if *input.Status == model.StatusClosed {
			addSet("closed_at", time.Now().UTC())
		}
	}
	if input.Priority != nil {
		addSet("priority", *input.Priority)
	}
	if input.Assignee != nil {
		addSet("assignee", nullEmpty(*input.Assignee))
	}
	if input.Owner != nil {
		addSet("owner", nullEmpty(*input.Owner))
	}
	if input.Pinned != nil {
		addSet("pinned", *input.Pinned)
	}
	if input.ExternalRef != nil {
		addSet("external_ref", nullEmpty(*input.ExternalRef))
	}
	if input.CloseReason != nil {
		addSet("close_reason", *input.CloseReason)
	}

	argN++
	args = append(args, id)
	query := fmt.Sprintf("UPDATE issues SET %s WHERE id = $%d RETURNING %s",
		strings.Join(sets, ", "), argN, issueColumns)

	issue, err := s.scanIssue(ctx, tx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("updating issue %s: %w", id, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing: %w", err)
	}

	return issue, nil
}

// ListIssues returns issues matching the filter.
func (s *PgStore) ListIssues(ctx context.Context, filter model.IssueFilter) ([]model.Issue, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	query := "SELECT " + issueColumns + " FROM issues WHERE 1=1"
	args := []any{}
	argN := 0

	addWhere := func(clause string, val any) {
		argN++
		query += fmt.Sprintf(" AND %s", fmt.Sprintf(clause, argN))
		args = append(args, val)
	}

	if filter.Status != nil {
		addWhere("status = $%d", string(*filter.Status))
	}
	if len(filter.StatusNot) > 0 {
		strs := make([]string, len(filter.StatusNot))
		for i, s := range filter.StatusNot {
			strs[i] = string(s)
		}
		addWhere("status != ALL($%d)", strs)
	}
	if filter.Priority != nil {
		addWhere("priority = $%d", *filter.Priority)
	}
	if filter.IssueType != nil {
		addWhere("issue_type = $%d", string(*filter.IssueType))
	}
	if filter.Assignee != nil {
		addWhere("assignee = $%d", *filter.Assignee)
	}
	if filter.Owner != nil {
		addWhere("owner = $%d", *filter.Owner)
	}
	if filter.Ephemeral != nil {
		addWhere("ephemeral = $%d", *filter.Ephemeral)
	}
	if filter.Pinned != nil {
		addWhere("pinned = $%d", *filter.Pinned)
	}
	if filter.Search != nil {
		addWhere("(title ILIKE '%%' || $%d || '%%' OR description ILIKE '%%' || $%d || '%%')", *filter.Search)
		// Note: search uses same arg twice, need to fix arg counting
	}
	if filter.ParentID != nil {
		argN++
		query += fmt.Sprintf(` AND id IN (SELECT issue_id FROM dependencies WHERE depends_on_id = $%d AND type = 'parent-child')`, argN)
		args = append(args, *filter.ParentID)
	}

	// Project filter
	query, args, argN = addProjectFilter(ctx, query, args, argN, "project_id")

	// Sort
	switch filter.SortBy {
	case "priority":
		query += " ORDER BY priority ASC, created_at ASC"
	case "oldest":
		query += " ORDER BY created_at ASC"
	case "updated":
		query += " ORDER BY updated_at DESC"
	default: // "hybrid"
		query += " ORDER BY priority ASC, updated_at DESC"
	}

	// Limit/offset
	if filter.Limit > 0 {
		argN++
		query += fmt.Sprintf(" LIMIT $%d", argN)
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		argN++
		query += fmt.Sprintf(" OFFSET $%d", argN)
		args = append(args, filter.Offset)
	}

	return s.scanIssues(ctx, s.pool, query, args...)
}

// ListReady returns issues from the ready_issues view.
func (s *PgStore) ListReady(ctx context.Context, filter model.IssueFilter) ([]model.Issue, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	query := "SELECT " + issueColumns + " FROM ready_issues"
	args := []any{}
	argN := 0

	where := []string{}
	if filter.IssueType != nil {
		argN++
		where = append(where, fmt.Sprintf("issue_type = $%d", argN))
		args = append(args, string(*filter.IssueType))
	}
	if filter.Priority != nil {
		argN++
		where = append(where, fmt.Sprintf("priority = $%d", argN))
		args = append(args, *filter.Priority)
	}
	if filter.Assignee != nil {
		argN++
		where = append(where, fmt.Sprintf("assignee = $%d", argN))
		args = append(args, *filter.Assignee)
	}

	// Project filter
	projectIDs := auth.AllowedProjectsFromContext(ctx)
	if len(projectIDs) > 0 {
		argN++
		where = append(where, fmt.Sprintf("project_id = ANY($%d::uuid[])", argN))
		args = append(args, projectIDs)
	}

	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}

	query += " ORDER BY priority ASC, created_at ASC"

	if filter.Limit > 0 {
		argN++
		query += fmt.Sprintf(" LIMIT $%d", argN)
		args = append(args, filter.Limit)
	}

	return s.scanIssues(ctx, s.pool, query, args...)
}

// DeleteIssue removes an issue and cascades to dependencies, labels, etc.
func (s *PgStore) DeleteIssue(ctx context.Context, id string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	tag, err := s.pool.Exec(ctx, "DELETE FROM issues WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("deleting issue: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("issue %s not found", id)
	}
	return nil
}

// --- Dependencies ---

func (s *PgStore) AddDependency(ctx context.Context, input AddDependencyInput) (*model.Dependency, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	now := time.Now().UTC()
	dep := &model.Dependency{
		IssueID:     input.IssueID,
		DependsOnID: input.DependsOnID,
		Type:        input.Type,
		CreatedAt:   now,
		CreatedBy:   input.CreatedBy,
		ThreadID:    input.ThreadID,
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO dependencies (issue_id, depends_on_id, type, created_at, created_by, thread_id)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (issue_id, depends_on_id) DO UPDATE SET type = $3`,
		dep.IssueID, dep.DependsOnID, dep.Type, dep.CreatedAt,
		nullEmpty(dep.CreatedBy), nullEmpty(dep.ThreadID))
	if err != nil {
		return nil, fmt.Errorf("adding dependency: %w", err)
	}

	return dep, nil
}

func (s *PgStore) RemoveDependency(ctx context.Context, issueID, dependsOnID string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	_, err := s.pool.Exec(ctx,
		"DELETE FROM dependencies WHERE issue_id = $1 AND depends_on_id = $2",
		issueID, dependsOnID)
	return err
}

func (s *PgStore) ListDependencies(ctx context.Context, issueID string, direction string) ([]model.Dependency, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	var query string
	switch direction {
	case "upstream":
		query = `SELECT issue_id, depends_on_id, type, created_at, created_by, metadata, thread_id
				 FROM dependencies WHERE issue_id = $1`
	case "downstream":
		query = `SELECT issue_id, depends_on_id, type, created_at, created_by, metadata, thread_id
				 FROM dependencies WHERE depends_on_id = $1`
	default: // both
		query = `SELECT issue_id, depends_on_id, type, created_at, created_by, metadata, thread_id
				 FROM dependencies WHERE issue_id = $1 OR depends_on_id = $1`
	}

	rows, err := s.pool.Query(ctx, query, issueID)
	if err != nil {
		return nil, fmt.Errorf("listing dependencies: %w", err)
	}
	defer rows.Close()

	var deps []model.Dependency
	for rows.Next() {
		var d model.Dependency
		var metadata []byte
		err := rows.Scan(&d.IssueID, &d.DependsOnID, &d.Type, &d.CreatedAt, &d.CreatedBy, &metadata, &d.ThreadID)
		if err != nil {
			return nil, fmt.Errorf("scanning dependency: %w", err)
		}
		d.Metadata = metadata
		deps = append(deps, d)
	}

	return deps, rows.Err()
}

func (s *PgStore) GetDependencyTree(ctx context.Context, rootID string, maxDepth int) ([]model.TreeNode, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	// Recursive CTE to walk parent-child tree
	rows, err := s.pool.Query(ctx, `
		WITH RECURSIVE tree AS (
			SELECT id, 0 as depth
			FROM issues WHERE id = $1
			UNION ALL
			SELECT i.id, t.depth + 1
			FROM issues i
			JOIN dependencies d ON d.issue_id = i.id AND d.type = 'parent-child'
			JOIN tree t ON d.depends_on_id = t.id
			WHERE t.depth < $2
		)
		SELECT t.depth, `+issueColumns+`
		FROM tree t
		JOIN issues i ON i.id = t.id
		ORDER BY t.depth, i.priority, i.created_at`, rootID, maxDepth)
	if err != nil {
		return nil, fmt.Errorf("walking dependency tree: %w", err)
	}
	defer rows.Close()

	var nodes []model.TreeNode
	for rows.Next() {
		var depth int
		issue, err := scanIssueFromRow(rows, &depth)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, model.TreeNode{
			Issue: *issue,
			Depth: depth,
		})
	}

	return nodes, rows.Err()
}

// --- Labels ---

func (s *PgStore) AddLabel(ctx context.Context, issueID, label string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	_, err := s.pool.Exec(ctx,
		"INSERT INTO labels (issue_id, label) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		issueID, label)
	return err
}

func (s *PgStore) RemoveLabel(ctx context.Context, issueID, label string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	_, err := s.pool.Exec(ctx,
		"DELETE FROM labels WHERE issue_id = $1 AND label = $2", issueID, label)
	return err
}

func (s *PgStore) ListLabels(ctx context.Context, issueID string) ([]string, error) {
	return s.queryLabels(ctx, issueID)
}

// --- Comments ---

func (s *PgStore) AddComment(ctx context.Context, issueID, author, text string) (*model.Comment, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	c := &model.Comment{}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO comments (issue_id, author, text) VALUES ($1, $2, $3)
		 RETURNING id, issue_id, author, text, created_at`,
		issueID, author, text).
		Scan(&c.ID, &c.IssueID, &c.Author, &c.Text, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("adding comment: %w", err)
	}
	return c, nil
}

func (s *PgStore) ListComments(ctx context.Context, issueID string) ([]model.Comment, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	rows, err := s.pool.Query(ctx,
		"SELECT id, issue_id, author, text, created_at FROM comments WHERE issue_id = $1 ORDER BY created_at",
		issueID)
	if err != nil {
		return nil, fmt.Errorf("listing comments: %w", err)
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.Author, &c.Text, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// --- Events ---

func (s *PgStore) AddEvent(ctx context.Context, input AddEventInput) (*model.Event, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	e := &model.Event{}
	err := s.pool.QueryRow(ctx,
		`INSERT INTO events (issue_id, event_type, actor, old_value, new_value, comment)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, issue_id, event_type, actor, old_value, new_value, comment, created_at`,
		input.IssueID, input.EventType, input.Actor,
		nullEmpty(input.OldValue), nullEmpty(input.NewValue), nullEmpty(input.Comment)).
		Scan(&e.ID, &e.IssueID, &e.EventType, &e.Actor, &e.OldValue, &e.NewValue, &e.Comment, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("adding event: %w", err)
	}
	return e, nil
}

func (s *PgStore) ListEvents(ctx context.Context, issueID string, limit int) ([]model.Event, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	query := "SELECT id, issue_id, event_type, actor, old_value, new_value, comment, created_at FROM events WHERE issue_id = $1 ORDER BY created_at DESC"
	args := []any{issueID}
	if limit > 0 {
		query += " LIMIT $2"
		args = append(args, limit)
	}

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing events: %w", err)
	}
	defer rows.Close()

	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(&e.ID, &e.IssueID, &e.EventType, &e.Actor, &e.OldValue, &e.NewValue, &e.Comment, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// --- Compaction ---

func (s *PgStore) SaveCompactionSnapshot(ctx context.Context, issueID string, level int, summary, original string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	_, err := s.pool.Exec(ctx,
		`INSERT INTO compaction_snapshots (issue_id, level, summary, original) VALUES ($1, $2, $3, $4)`,
		issueID, level, summary, original)
	return err
}

func (s *PgStore) GetCompactionSnapshots(ctx context.Context, issueID string) ([]model.CompactionSnapshot, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	rows, err := s.pool.Query(ctx,
		"SELECT id, issue_id, level, summary, original, created_at FROM compaction_snapshots WHERE issue_id = $1 ORDER BY level",
		issueID)
	if err != nil {
		return nil, fmt.Errorf("listing snapshots: %w", err)
	}
	defer rows.Close()

	var snaps []model.CompactionSnapshot
	for rows.Next() {
		var snap model.CompactionSnapshot
		if err := rows.Scan(&snap.ID, &snap.IssueID, &snap.Level, &snap.Summary, &snap.Original, &snap.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning snapshot: %w", err)
		}
		snaps = append(snaps, snap)
	}
	return snaps, rows.Err()
}

// --- Aggregation ---

func (s *PgStore) CountIssuesByStatus(ctx context.Context) (map[string]int, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()

	query := "SELECT status, COUNT(*) FROM issues WHERE 1=1"
	args := []any{}
	argN := 0
	query, args, _ = addProjectFilter(ctx, query, args, argN, "project_id")
	query += " GROUP BY status"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("counting issues by status: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scanning count: %w", err)
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

// --- Helpers ---

const issueColumns = `id, content_hash, title, description, design, acceptance_criteria, notes,
	spec_id, status, priority, issue_type, assignee, owner, estimated_minutes,
	created_at, created_by, updated_at, closed_at, due_at, defer_until,
	close_reason, closed_by_session, external_ref, source_system, source_repo,
	metadata, compaction_level, compacted_at, compacted_at_commit, original_size,
	sender, ephemeral, mol_type, work_type, crystallizes, wisp_type,
	pinned, is_template, quality_score, event_kind, actor, target, payload,
	await_type, await_id, timeout_ns, agent_state, last_activity, role_type, rig,
	hook_bead, role_bead`

type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (s *PgStore) scanIssue(ctx context.Context, q querier, query string, args ...any) (*model.Issue, error) {
	row := q.QueryRow(ctx, query, args...)
	return scanIssueFromSingleRow(row)
}

func scanIssueFromSingleRow(row pgx.Row) (*model.Issue, error) {
	var i model.Issue
	var metadata []byte
	err := row.Scan(
		&i.ID, &i.ContentHash, &i.Title, &i.Description, &i.Design,
		&i.AcceptanceCriteria, &i.Notes, &i.SpecID, &i.Status, &i.Priority,
		&i.IssueType, &i.Assignee, &i.Owner, &i.EstimatedMinutes,
		&i.CreatedAt, &i.CreatedBy, &i.UpdatedAt, &i.ClosedAt, &i.DueAt, &i.DeferUntil,
		&i.CloseReason, &i.ClosedBySession, &i.ExternalRef, &i.SourceSystem, &i.SourceRepo,
		&metadata, &i.CompactionLevel, &i.CompactedAt, &i.CompactedAtCommit, &i.OriginalSize,
		&i.Sender, &i.Ephemeral, &i.MolType, &i.WorkType, &i.Crystallizes, &i.WispType,
		&i.Pinned, &i.IsTemplate, &i.QualityScore, &i.EventKind, &i.Actor, &i.Target, &i.Payload,
		&i.AwaitType, &i.AwaitID, &i.Timeout, &i.AgentState, &i.LastActivity, &i.RoleType, &i.Rig,
		&i.HookBead, &i.RoleBead,
	)
	if err != nil {
		return nil, err
	}
	i.Metadata = metadata
	return &i, nil
}

func scanIssueFromRow(rows pgx.Rows, extraFields ...any) (*model.Issue, error) {
	var i model.Issue
	var metadata []byte

	scanArgs := make([]any, 0, len(extraFields)+52)
	scanArgs = append(scanArgs, extraFields...)
	scanArgs = append(scanArgs,
		&i.ID, &i.ContentHash, &i.Title, &i.Description, &i.Design,
		&i.AcceptanceCriteria, &i.Notes, &i.SpecID, &i.Status, &i.Priority,
		&i.IssueType, &i.Assignee, &i.Owner, &i.EstimatedMinutes,
		&i.CreatedAt, &i.CreatedBy, &i.UpdatedAt, &i.ClosedAt, &i.DueAt, &i.DeferUntil,
		&i.CloseReason, &i.ClosedBySession, &i.ExternalRef, &i.SourceSystem, &i.SourceRepo,
		&metadata, &i.CompactionLevel, &i.CompactedAt, &i.CompactedAtCommit, &i.OriginalSize,
		&i.Sender, &i.Ephemeral, &i.MolType, &i.WorkType, &i.Crystallizes, &i.WispType,
		&i.Pinned, &i.IsTemplate, &i.QualityScore, &i.EventKind, &i.Actor, &i.Target, &i.Payload,
		&i.AwaitType, &i.AwaitID, &i.Timeout, &i.AgentState, &i.LastActivity, &i.RoleType, &i.Rig,
		&i.HookBead, &i.RoleBead,
	)

	if err := rows.Scan(scanArgs...); err != nil {
		return nil, fmt.Errorf("scanning issue: %w", err)
	}
	i.Metadata = metadata
	return &i, nil
}

func (s *PgStore) scanIssues(ctx context.Context, q querier, query string, args ...any) ([]model.Issue, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying issues: %w", err)
	}
	defer rows.Close()

	var issues []model.Issue
	for rows.Next() {
		issue, err := scanIssueFromRow(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, *issue)
	}
	return issues, rows.Err()
}

func (s *PgStore) queryLabels(ctx context.Context, issueID string) ([]string, error) {
	rows, err := s.pool.Query(ctx,
		"SELECT label FROM labels WHERE issue_id = $1 ORDER BY label", issueID)
	if err != nil {
		return nil, fmt.Errorf("querying labels: %w", err)
	}
	defer rows.Close()

	var labels []string
	for rows.Next() {
		var l string
		if err := rows.Scan(&l); err != nil {
			return nil, err
		}
		labels = append(labels, l)
	}
	return labels, rows.Err()
}

func contentHash(i *model.Issue) string {
	h := sha256.New()
	h.Write([]byte(i.Title))
	h.Write([]byte(i.Description))
	h.Write([]byte(i.Design))
	h.Write([]byte(i.AcceptanceCriteria))
	h.Write([]byte(i.Notes))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func nullEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
