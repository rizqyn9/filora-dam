package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func galleriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "galleries",
		Short: "Manage galleries",
	}
	cmd.AddCommand(galleriesListCmd())
	return cmd
}

func galleriesListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List galleries you belong to",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := authedClient()
			if err != nil {
				return err
			}
			var galleries []struct {
				ID           int64  `json:"id"`
				Name         string `json:"name"`
				IsDefault    bool   `json:"is_default"`
				StorageUsed  int64  `json:"storage_used"`
				StorageQuota int64  `json:"storage_quota"`
			}
			if err := c.Do("GET", "/api/v1/galleries", nil, &galleries); err != nil {
				return err
			}
			for _, g := range galleries {
				def := ""
				if g.IsDefault {
					def = " (default)"
				}
				fmt.Printf("%-6d %s%s  %s / %s\n", g.ID, g.Name, def,
					humanBytes(g.StorageUsed), humanBytes(g.StorageQuota))
			}
			return nil
		},
	}
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(n)/float64(div), "KMGTPE"[exp])
}
