package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/spf13/cobra"
)

func newReadyCmd() *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "ready",
		Short: "List issues ready for work (no open blockers)",
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

			issues, err := pg.ListReady(ctx, model.IssueFilter{Limit: limit})
			if err != nil {
				return fmt.Errorf("listing ready issues: %w", err)
			}

			if jsonOutput {
				outputJSON(issues)
				return nil
			}

			if len(issues) == 0 {
				fmt.Println("No ready work items.")
				return nil
			}

			fmt.Printf("📋 Ready work (%d issues with no blockers):\n\n", len(issues))
			for idx, i := range issues {
				fmt.Printf("%d. [P%d] [%s] %s: %s\n", idx+1, i.Priority, i.IssueType, i.ID, i.Title)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Max results")

	return cmd
}
