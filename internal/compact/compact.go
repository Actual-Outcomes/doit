// Package compact implements semantic compaction (memory decay) for issues.
// Closed issues are progressively summarized to save context window tokens.
//
// Compaction levels:
//   - 0: Full detail (original content preserved)
//   - 1: Summarized — description/design/notes replaced with a short summary
//   - 2: Minimal — only title, status, and key metadata retained
package compact

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
)

// Compactor manages the memory decay process.
type Compactor struct {
	store store.Store
}

func New(s store.Store) *Compactor {
	return &Compactor{store: s}
}

// CompactResult reports what was done.
type CompactResult struct {
	IssueID string `json:"issue_id"`
	OldLevel int   `json:"old_level"`
	NewLevel int   `json:"new_level"`
}

// CompactOld finds closed issues older than the threshold and bumps their compaction level.
// Level 0 → 1 after closedAge; Level 1 → 2 after 2*closedAge.
func (c *Compactor) CompactOld(ctx context.Context, closedAge time.Duration) ([]CompactResult, error) {
	closedStatus := model.StatusClosed
	issues, err := c.store.ListIssues(ctx, model.IssueFilter{
		Status: &closedStatus,
		Limit:  500,
		SortBy: "oldest",
	})
	if err != nil {
		return nil, fmt.Errorf("listing closed issues: %w", err)
	}

	now := time.Now().UTC()
	var results []CompactResult

	for _, issue := range issues {
		if issue.ClosedAt == nil {
			continue
		}

		age := now.Sub(*issue.ClosedAt)
		targetLevel := 0
		if age > 2*closedAge {
			targetLevel = 2
		} else if age > closedAge {
			targetLevel = 1
		}

		if targetLevel <= issue.CompactionLevel {
			continue
		}

		// Snapshot original before compacting
		original := formatOriginal(&issue)
		summary := generateSummary(&issue, targetLevel)

		if err := c.store.SaveCompactionSnapshot(ctx, issue.ID, targetLevel, summary, original); err != nil {
			return nil, fmt.Errorf("saving snapshot for %s: %w", issue.ID, err)
		}

		// Update the issue with compacted content
		input := store.UpdateIssueInput{}
		desc := summary
		input.Description = &desc
		empty := ""
		input.Design = &empty
		input.Notes = &empty
		if targetLevel == 2 {
			input.AcceptanceCriteria = &empty
		}

		if _, err := c.store.UpdateIssue(ctx, issue.ID, input); err != nil {
			return nil, fmt.Errorf("updating %s: %w", issue.ID, err)
		}

		results = append(results, CompactResult{
			IssueID:  issue.ID,
			OldLevel: issue.CompactionLevel,
			NewLevel: targetLevel,
		})
	}

	return results, nil
}

// generateSummary creates a compact summary of the issue.
func generateSummary(issue *model.Issue, level int) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("[%s] %s", issue.IssueType, issue.Title))

	if level == 1 {
		// Keep a sentence or two from description
		if issue.Description != "" {
			lines := strings.SplitN(issue.Description, "\n", 3)
			if len(lines) > 2 {
				parts = append(parts, strings.Join(lines[:2], " "))
			} else {
				parts = append(parts, issue.Description)
			}
		}
		if issue.CloseReason != "" {
			parts = append(parts, fmt.Sprintf("Closed: %s", issue.CloseReason))
		}
	} else {
		// Level 2: minimal
		if issue.CloseReason != "" {
			parts = append(parts, fmt.Sprintf("Closed: %s", issue.CloseReason))
		}
	}

	return strings.Join(parts, " | ")
}

// formatOriginal captures the full issue content for snapshot preservation.
func formatOriginal(issue *model.Issue) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Title: %s\n", issue.Title))
	if issue.Description != "" {
		b.WriteString(fmt.Sprintf("Description: %s\n", issue.Description))
	}
	if issue.Design != "" {
		b.WriteString(fmt.Sprintf("Design: %s\n", issue.Design))
	}
	if issue.AcceptanceCriteria != "" {
		b.WriteString(fmt.Sprintf("Acceptance Criteria: %s\n", issue.AcceptanceCriteria))
	}
	if issue.Notes != "" {
		b.WriteString(fmt.Sprintf("Notes: %s\n", issue.Notes))
	}
	return b.String()
}
