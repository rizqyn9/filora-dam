package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rizqynugroho9/filora-dam/cli/internal/client"
	"github.com/rizqynugroho9/filora-dam/cli/internal/config"
)

func loginCmd() *cobra.Command {
	var apiURL, bootstrap, label string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate and store a CLI token",
		Long: "Exchanges a bootstrap token (a Clerk web session token or an existing\n" +
			"CLI token) for a dedicated, revocable CLI token stored under ~/.filora.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if apiURL != "" {
				cfg.APIURL = apiURL
			}
			if bootstrap == "" {
				return fmt.Errorf("--token is required (a Clerk session token or existing CLI token)")
			}
			if label == "" {
				label, _ = os.Hostname()
			}

			// Mint a CLI token using the bootstrap token as the bearer.
			var res struct {
				Token   string `json:"token"`
				Session struct {
					ID string `json:"id"`
				} `json:"session"`
			}
			boot := client.New(cfg.APIURL, bootstrap)
			if err := boot.Do("POST", "/api/v1/cli/sessions", map[string]any{"label": label}, &res); err != nil {
				return err
			}
			cfg.Token = res.Token
			cfg.SessionID = res.Session.ID

			// Best-effort: fetch the profile for display.
			var user struct {
				Email string `json:"email"`
			}
			_ = client.New(cfg.APIURL, cfg.Token).Do("GET", "/api/v1/me", nil, &user)
			cfg.Email = user.Email

			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Printf("Logged in as %s (session %s)\n", cfg.Email, cfg.SessionID)
			return nil
		},
	}
	cmd.Flags().StringVar(&apiURL, "api-url", "", "API base URL (default: config or http://localhost:3000)")
	cmd.Flags().StringVar(&bootstrap, "token", "", "bootstrap token (Clerk session or existing CLI token)")
	cmd.Flags().StringVar(&label, "label", "", "session label (default: hostname)")
	return cmd
}

func logoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Revoke this CLI session and clear local credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.Token != "" && cfg.SessionID != "" {
				c := client.New(cfg.APIURL, cfg.Token)
				_ = c.Do("DELETE", "/api/v1/cli/sessions/"+cfg.SessionID, nil, nil) // best effort
			}
			if err := config.Clear(); err != nil {
				return err
			}
			fmt.Println("Logged out.")
			return nil
		},
	}
}

func whoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := authedClient()
			if err != nil {
				return err
			}
			var user map[string]any
			if err := c.Do("GET", "/api/v1/me", nil, &user); err != nil {
				return err
			}
			printJSON(user)
			return nil
		},
	}
}
