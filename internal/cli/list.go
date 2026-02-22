package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		status    string
		issueType string
		priority  int
		assignee  string
		limit     int
		sortBy    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
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

			filter := model.IssueFilter{
				Limit:  limit,
				SortBy: sortBy,
			}
			if status != "" {
				s := model.Status(status)
				filter.Status = &s
			}
			if issueType != "" {
				t := model.IssueType(issueType)
				filter.IssueType = &t
			}
			if cmd.Flags().Changed("priority") {
				filter.Priority = &priority
			}
			if assignee != "" {
				filter.Assignee = &assignee
			}

			issues, err := pg.ListIssues(ctx, filter)
			if err != nil {
				return fmt.Errorf("listing issues: %w", err)
			}

			if jsonOutput {
				outputJSON(issues)
				return nil
			}

			if len(issues) == 0 {
				fmt.Println("No issues found.")
				return nil
			}

			for _, i := range issues {
				statusIcon := "○"
				switch i.Status {
				case model.StatusInProgress:
					statusIcon = "◐"
				case model.StatusClosed:
					statusIcon = "●"
				case model.StatusBlocked:
					statusIcon = "✕"
				}
				fmt.Printf("%s %s [P%d] [%s] %s\n", statusIcon, i.ID, i.Priority, i.IssueType, i.Title)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "Filter by status")
	cmd.Flags().StringVarP(&issueType, "type", "t", "", "Filter by type")
	cmd.Flags().IntVarP(&priority, "priority", "p", -1, "Filter by priority")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Filter by assignee")
	cmd.Flags().IntVarP(&limit, "limit", "l", 50, "Max results")
	cmd.Flags().StringVar(&sortBy, "sort", "hybrid", "Sort by: priority, oldest, updated, hybrid")

	return cmd
}
