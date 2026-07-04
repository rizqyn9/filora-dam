// Package commands implements the Filora CLI (Cobra).
package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rizqynugroho9/filora-dam/cli/internal/client"
	"github.com/rizqynugroho9/filora-dam/cli/internal/config"
)

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "filora",
		Short: "Filora DAM command-line client",
		Long:  "Filora is a multi-cloud Digital Asset Management CLI (thin client over the Filora API).",
	}
	root.AddCommand(
		loginCmd(),
		logoutCmd(),
		whoamiCmd(),
		sessionsCmd(),
		galleriesCmd(),
		uploadCmd(),
		assetsCmd(),
		downloadCmd(),
	)
	return root
}

// authedClient loads config and returns an authenticated API client.
func authedClient() (*client.Client, *config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}
	if cfg.Token == "" {
		return nil, nil, fmt.Errorf("not logged in; run 'filora login --token <token>'")
	}
	return client.New(cfg.APIURL, cfg.Token), cfg, nil
}

func printJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}
