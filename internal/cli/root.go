package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Actual-Outcomes/doit/internal/version"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	quiet      bool
	dbURL      string
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "doit",
		Short: "AI agent work planner & tracker",
		Long:  "doit is a CLI for AI agents to plan and track work. It manages issues with dependencies, hierarchical tasks, and ready detection.",
		SilenceUsage: true,
	}

	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	root.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	root.PersistentFlags().StringVar(&dbURL, "db", "", "Database URL (default: $DATABASE_URL)")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newCreateCmd())
	root.AddCommand(newShowCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newReadyCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newDepCmd())
	root.AddCommand(newMessageCmd())
	root.AddCommand(newCompactCmd())

	return root
}

func getDBURL() string {
	if dbURL != "" {
		return dbURL
	}
	if v := os.Getenv("DATABASE_URL"); v != "" {
		return v
	}
	return ""
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("doit version %s\n", version.Number)
		},
	}
}

// outputJSON prints v as indented JSON.
func outputJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

// printSuccess prints a success message unless quiet mode is on.
func printSuccess(format string, args ...any) {
	if !quiet {
		fmt.Printf("✓ "+format+"\n", args...)
	}
}

// printError prints an error message to stderr.
func printError(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}
