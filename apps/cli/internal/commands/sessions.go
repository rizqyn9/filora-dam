package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func sessionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "Manage CLI sessions",
	}
	cmd.AddCommand(sessionsListCmd(), sessionsRevokeCmd())
	return cmd
}

func sessionsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List active CLI sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, cfg, err := authedClient()
			if err != nil {
				return err
			}
			var sessions []struct {
				ID         string  `json:"id"`
				Label      *string `json:"label"`
				LastUsedAt *string `json:"last_used_at"`
				CreatedAt  string  `json:"created_at"`
			}
			if err := c.Do("GET", "/api/v1/cli/sessions", nil, &sessions); err != nil {
				return err
			}
			for _, s := range sessions {
				label := "-"
				if s.Label != nil {
					label = *s.Label
				}
				marker := " "
				if s.ID == cfg.SessionID {
					marker = "*"
				}
				fmt.Printf("%s %s  %-20s  created %s\n", marker, s.ID, label, s.CreatedAt)
			}
			return nil
		},
	}
}

func sessionsRevokeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "revoke <session-id>",
		Short: "Revoke a CLI session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := authedClient()
			if err != nil {
				return err
			}
			if err := c.Do("DELETE", "/api/v1/cli/sessions/"+args[0], nil, nil); err != nil {
				return err
			}
			fmt.Println("Session revoked.")
			return nil
		},
	}
}
