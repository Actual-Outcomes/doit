package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	// Register message commands on root in root.go
}

func newMessageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "msg",
		Aliases: []string{"message"},
		Short:   "Agent-to-agent messaging",
	}

	cmd.AddCommand(newMsgSendCmd())
	cmd.AddCommand(newMsgListCmd())
	cmd.AddCommand(newMsgReadCmd())

	return cmd
}

func newMsgSendCmd() *cobra.Command {
	var (
		to      string
		thread  string
		ephemeral bool
	)

	cmd := &cobra.Command{
		Use:   "send <body>",
		Short: "Send a message to another agent",
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

			id, err := pg.GenerateID(ctx, "msg")
			if err != nil {
				return fmt.Errorf("generating ID: %w", err)
			}

			issue, err := pg.CreateIssue(ctx, store.CreateIssueInput{
				ID:          id,
				Title:       truncate(args[0], 80),
				Description: args[0],
				Status:      model.StatusOpen,
				Priority:    2,
				IssueType:   model.TypeMessage,
				Assignee:    to,
				Ephemeral:   ephemeral,
			})
			if err != nil {
				return fmt.Errorf("sending message: %w", err)
			}

			// Thread via replies-to dependency
			if thread != "" {
				_, err = pg.AddDependency(ctx, store.AddDependencyInput{
					IssueID:     issue.ID,
					DependsOnID: thread,
					Type:        model.DepRepliesTo,
					ThreadID:    thread,
				})
				if err != nil {
					return fmt.Errorf("threading message: %w", err)
				}
			}

			if jsonOutput {
				outputJSON(issue)
			} else {
				printSuccess("Sent message: %s → %s", issue.ID, to)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Recipient (assignee)")
	cmd.Flags().StringVar(&thread, "thread", "", "Reply to this message ID (creates thread)")
	cmd.Flags().BoolVar(&ephemeral, "ephemeral", true, "Ephemeral message (default true)")

	return cmd
}

func newMsgListCmd() *cobra.Command {
	var (
		to     string
		unread bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List messages",
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

			msgType := model.TypeMessage
			filter := model.IssueFilter{
				IssueType: &msgType,
				Limit:     50,
				SortBy:    "updated",
			}
			if to != "" {
				filter.Assignee = &to
			}
			if unread {
				s := model.StatusOpen
				filter.Status = &s
			}

			issues, err := pg.ListIssues(ctx, filter)
			if err != nil {
				return fmt.Errorf("listing messages: %w", err)
			}

			if jsonOutput {
				outputJSON(issues)
				return nil
			}

			if len(issues) == 0 {
				fmt.Println("No messages.")
				return nil
			}

			for _, i := range issues {
				status := "●"
				if i.Status == model.StatusClosed {
					status = "○"
				}
				from := i.Sender
				if from == "" {
					from = i.CreatedBy
				}
				fmt.Printf("%s %s [%s→%s] %s\n", status, i.ID, from, i.Assignee, truncate(i.Title, 60))
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Filter by recipient")
	cmd.Flags().BoolVar(&unread, "unread", false, "Only show unread (open) messages")

	return cmd
}

func newMsgReadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "read <id>",
		Short: "Mark a message as read (closed)",
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

			s := model.StatusClosed
			_, err = pg.UpdateIssue(ctx, args[0], store.UpdateIssueInput{
				Status: &s,
			})
			if err != nil {
				return fmt.Errorf("marking as read: %w", err)
			}

			printSuccess("Marked %s as read", args[0])
			return nil
		},
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
