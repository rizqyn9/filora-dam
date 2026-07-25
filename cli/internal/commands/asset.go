package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func uploadCmd() *cobra.Command {
	var galleryID int64
	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload a file to a gallery",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if galleryID == 0 {
				return fmt.Errorf("--gallery <id> is required")
			}
			c, _, err := authedClient()
			if err != nil {
				return err
			}
			var asset map[string]any
			path := "/api/v1/galleries/" + strconv.FormatInt(galleryID, 10) + "/assets"
			if err := c.Upload(path, args[0], &asset); err != nil {
				return err
			}
			printJSON(asset)
			return nil
		},
	}
	cmd.Flags().Int64Var(&galleryID, "gallery", 0, "target gallery id (required)")
	return cmd
}

func assetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Manage assets",
	}
	cmd.AddCommand(assetsListCmd())
	return cmd
}

func assetsListCmd() *cobra.Command {
	var galleryID int64
	var limit, offset int
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List assets in a gallery",
		RunE: func(cmd *cobra.Command, args []string) error {
			if galleryID == 0 {
				return fmt.Errorf("--gallery <id> is required")
			}
			c, _, err := authedClient()
			if err != nil {
				return err
			}
			path := fmt.Sprintf("/api/v1/galleries/%d/assets?limit=%d&offset=%d", galleryID, limit, offset)
			var res struct {
				Assets []struct {
					ID       string `json:"id"`
					Name     string `json:"name"`
					Type     string `json:"type"`
					Size     int64  `json:"size"`
					MimeType string `json:"mime_type"`
				} `json:"assets"`
				Total int64 `json:"total"`
			}
			if err := c.Do("GET", path, nil, &res); err != nil {
				return err
			}
			for _, a := range res.Assets {
				fmt.Printf("%s  %-8s %-9s %s\n", a.ID, a.Type, humanBytes(a.Size), a.Name)
			}
			fmt.Printf("\n%d of %d\n", len(res.Assets), res.Total)
			return nil
		},
	}
	cmd.Flags().Int64Var(&galleryID, "gallery", 0, "gallery id (required)")
	cmd.Flags().IntVar(&limit, "limit", 20, "page size")
	cmd.Flags().IntVar(&offset, "offset", 0, "page offset")
	return cmd
}

func downloadCmd() *cobra.Command {
	var out string
	cmd := &cobra.Command{
		Use:   "download <asset-id>",
		Short: "Download an asset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, _, err := authedClient()
			if err != nil {
				return err
			}
			if out == "" {
				out = args[0]
			}
			f, err := os.Create(out)
			if err != nil {
				return err
			}
			defer f.Close()
			if err := c.Download("/api/v1/assets/"+args[0]+"/download", f); err != nil {
				return err
			}
			fmt.Printf("Saved to %s\n", out)
			return nil
		},
	}
	cmd.Flags().StringVarP(&out, "output", "o", "", "output file (default: asset id)")
	return cmd
}
