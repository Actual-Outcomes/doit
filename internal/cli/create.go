package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var (
		priority    int
		issueType   string
		assignee    string
		owner       string
		parent      string
		description string
		design      string
		acceptance  string
		notes       string
		labels      []string
		ephemeral   bool
	)

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new issue",
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

			// Generate ID
			var id string
			if parent != "" {
				id, err = pg.NextChildID(ctx, parent)
			} else {
				id, err = pg.GenerateID(ctx, "")
			}
			if err != nil {
				return fmt.Errorf("generating ID: %w", err)
			}

			issue, err := pg.CreateIssue(ctx, store.CreateIssueInput{
				ID:                 id,
				Title:              args[0],
				Description:        description,
				Design:             design,
				AcceptanceCriteria: acceptance,
				Notes:              notes,
				Status:             model.StatusOpen,
				Priority:           priority,
				IssueType:          model.IssueType(issueType),
				Assignee:           assignee,
				Owner:              owner,
				ParentID:           parent,
				Labels:             labels,
				Ephemeral:          ephemeral,
			})
			if err != nil {
				return fmt.Errorf("creating issue: %w", err)
			}

			if jsonOutput {
				outputJSON(issue)
			} else {
				printSuccess("Created issue: %s", issue.ID)
				fmt.Printf("  Title: %s\n", issue.Title)
				fmt.Printf("  Priority: P%d\n", issue.Priority)
				fmt.Printf("  Status: %s\n", issue.Status)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&priority, "priority", "p", 2, "Priority (0-4, 0=highest)")
	cmd.Flags().StringVarP(&issueType, "type", "t", "task", "Issue type (bug|feature|task|epic|chore|decision|message)")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Assignee")
	cmd.Flags().StringVar(&owner, "owner", "", "Owner")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent issue ID (creates hierarchical child)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Description")
	cmd.Flags().StringVar(&design, "design", "", "Design notes")
	cmd.Flags().StringVar(&acceptance, "acceptance", "", "Acceptance criteria")
	cmd.Flags().StringVar(&notes, "notes", "", "Additional notes")
	cmd.Flags().StringSliceVar(&labels, "label", nil, "Labels (repeatable)")
	cmd.Flags().BoolVar(&ephemeral, "ephemeral", false, "Ephemeral (excluded from export)")

	return cmd
}
