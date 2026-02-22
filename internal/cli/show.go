package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show issue details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbURL := getDBURL()
			if dbURL == "" {
				return fmt.Errorf("DATABASE_URL not set")
			}

			ctx := context.Background()
			pg, err := store.NewPgStore(ctx, dbURL, 10*time.Second, "")
			if err != nil {
				return fmt.Errorf("connecting to database: %w", err)
			}
			defer pg.Close()

			issue, err := pg.GetIssue(ctx, args[0])
			if err != nil {
				return fmt.Errorf("getting issue: %w", err)
			}

			if jsonOutput {
				outputJSON(issue)
				return nil
			}

			// Pretty print
			fmt.Printf("%s: %s\n", issue.ID, issue.Title)
			fmt.Printf("  Status:   %s\n", issue.Status)
			fmt.Printf("  Priority: P%d\n", issue.Priority)
			fmt.Printf("  Type:     %s\n", issue.IssueType)

			if issue.Assignee != "" {
				fmt.Printf("  Assignee: %s\n", issue.Assignee)
			}
			if issue.Owner != "" {
				fmt.Printf("  Owner:    %s\n", issue.Owner)
			}
			if issue.ParentID != "" {
				fmt.Printf("  Parent:   %s\n", issue.ParentID)
			}
			if len(issue.Labels) > 0 {
				fmt.Printf("  Labels:   %s\n", strings.Join(issue.Labels, ", "))
			}
			if issue.Description != "" {
				fmt.Printf("\n  Description:\n    %s\n", issue.Description)
			}
			if issue.Design != "" {
				fmt.Printf("\n  Design:\n    %s\n", issue.Design)
			}
			if issue.AcceptanceCriteria != "" {
				fmt.Printf("\n  Acceptance Criteria:\n    %s\n", issue.AcceptanceCriteria)
			}

			fmt.Printf("\n  Created: %s\n", issue.CreatedAt.Format(time.RFC3339))
			fmt.Printf("  Updated: %s\n", issue.UpdatedAt.Format(time.RFC3339))

			// Show dependencies
			deps, err := pg.ListDependencies(ctx, args[0], "both")
			if err == nil && len(deps) > 0 {
				fmt.Println("\n  Dependencies:")
				for _, d := range deps {
					if d.IssueID == args[0] {
						fmt.Printf("    → %s (%s) %s\n", d.DependsOnID, d.Type, d.DependsOnID)
					} else {
						fmt.Printf("    ← %s (%s) %s\n", d.IssueID, d.Type, d.IssueID)
					}
				}
			}

			return nil
		},
	}

	return cmd
}
