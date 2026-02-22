package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Actual-Outcomes/doit/internal/compact"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/spf13/cobra"
)

func newCompactCmd() *cobra.Command {
	var age string

	cmd := &cobra.Command{
		Use:   "compact",
		Short: "Run semantic compaction on old closed issues",
		Long:  "Summarizes old closed issues to save context window tokens. Level 0→1 after threshold, Level 1→2 after 2x threshold.",
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

			threshold, err := time.ParseDuration(age)
			if err != nil {
				return fmt.Errorf("invalid age duration %q: %w", age, err)
			}

			compactor := compact.New(pg)
			results, err := compactor.CompactOld(ctx, threshold)
			if err != nil {
				return fmt.Errorf("compacting: %w", err)
			}

			if jsonOutput {
				outputJSON(results)
				return nil
			}

			if len(results) == 0 {
				fmt.Println("No issues needed compaction.")
				return nil
			}

			for _, r := range results {
				fmt.Printf("  %s: level %d → %d\n", r.IssueID, r.OldLevel, r.NewLevel)
			}
			printSuccess("Compacted %d issues", len(results))

			return nil
		},
	}

	cmd.Flags().StringVar(&age, "age", "168h", "Compaction threshold (e.g. 168h for 7 days)")

	return cmd
}
