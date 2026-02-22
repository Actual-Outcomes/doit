package cli

import (
	"context"
	"fmt"
	"os/user"
	"time"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var (
		title       string
		description string
		status      string
		priority    int
		assignee    string
		owner       string
		claim       bool
		pinned      bool
		notes       string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an issue",
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

			input := store.UpdateIssueInput{}

			if claim {
				// Atomic claim: set assignee to current user, status to in_progress
				u, _ := user.Current()
				username := "agent"
				if u != nil {
					username = u.Username
				}
				input.Assignee = &username
				s := model.StatusInProgress
				input.Status = &s
			}

			if cmd.Flags().Changed("title") {
				input.Title = &title
			}
			if cmd.Flags().Changed("description") {
				input.Description = &description
			}
			if cmd.Flags().Changed("status") {
				s := model.Status(status)
				input.Status = &s
			}
			if cmd.Flags().Changed("priority") {
				input.Priority = &priority
			}
			if cmd.Flags().Changed("assignee") {
				input.Assignee = &assignee
			}
			if cmd.Flags().Changed("owner") {
				input.Owner = &owner
			}
			if cmd.Flags().Changed("pinned") {
				input.Pinned = &pinned
			}
			if cmd.Flags().Changed("notes") {
				input.Notes = &notes
			}

			issue, err := pg.UpdateIssue(ctx, args[0], input)
			if err != nil {
				return fmt.Errorf("updating issue: %w", err)
			}

			if jsonOutput {
				outputJSON(issue)
			} else {
				printSuccess("Updated issue: %s", issue.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVarP(&description, "description", "d", "", "New description")
	cmd.Flags().StringVarP(&status, "status", "s", "", "New status")
	cmd.Flags().IntVarP(&priority, "priority", "p", -1, "New priority (0-4)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "New assignee")
	cmd.Flags().StringVar(&owner, "owner", "", "New owner")
	cmd.Flags().BoolVar(&claim, "claim", false, "Atomically claim (sets assignee + in_progress)")
	cmd.Flags().BoolVar(&pinned, "pinned", false, "Pin/unpin issue")
	cmd.Flags().StringVar(&notes, "notes", "", "Additional notes")

	return cmd
}
