package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/spf13/cobra"
)

func newDepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dep",
		Short: "Manage issue dependencies",
	}

	cmd.AddCommand(newDepAddCmd())
	cmd.AddCommand(newDepRemoveCmd())
	cmd.AddCommand(newDepListCmd())

	return cmd
}

func newDepAddCmd() *cobra.Command {
	var depType string

	cmd := &cobra.Command{
		Use:   "add <issue-id> <depends-on-id>",
		Short: "Add a dependency (issue depends on depends-on)",
		Args:  cobra.ExactArgs(2),
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

			dep, err := pg.AddDependency(ctx, store.AddDependencyInput{
				IssueID:     args[0],
				DependsOnID: args[1],
				Type:        model.DependencyType(depType),
			})
			if err != nil {
				return fmt.Errorf("adding dependency: %w", err)
			}

			if jsonOutput {
				outputJSON(dep)
			} else {
				printSuccess("Added dependency: %s depends on %s (%s)", args[0], args[1], depType)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&depType, "type", "t", "blocks", "Dependency type")

	return cmd
}

func newDepRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <issue-id> <depends-on-id>",
		Short: "Remove a dependency",
		Args:  cobra.ExactArgs(2),
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

			if err := pg.RemoveDependency(ctx, args[0], args[1]); err != nil {
				return fmt.Errorf("removing dependency: %w", err)
			}

			printSuccess("Removed dependency: %s → %s", args[0], args[1])
			return nil
		},
	}
}

func newDepListCmd() *cobra.Command {
	var direction string

	cmd := &cobra.Command{
		Use:   "list <issue-id>",
		Short: "List dependencies for an issue",
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

			deps, err := pg.ListDependencies(ctx, args[0], direction)
			if err != nil {
				return fmt.Errorf("listing dependencies: %w", err)
			}

			if jsonOutput {
				outputJSON(deps)
				return nil
			}

			if len(deps) == 0 {
				fmt.Println("No dependencies.")
				return nil
			}

			for _, d := range deps {
				if d.IssueID == args[0] {
					fmt.Printf("  → depends on %s (%s)\n", d.DependsOnID, d.Type)
				} else {
					fmt.Printf("  ← depended on by %s (%s)\n", d.IssueID, d.Type)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&direction, "direction", "both", "Filter: upstream, downstream, both")

	return cmd
}
