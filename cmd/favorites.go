package cmd

import (
	"fmt"
	"strconv"

	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var favoritesCmd = &cobra.Command{
	Use:   "favorites",
	Short: "Manage favorites",
}

var favoritesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List favorites",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewFavoriteService(client)

		favType, _ := cmd.Flags().GetString("type")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		resp, err := svc.List(cmdContext(), api.FavoriteListParams{
			Type: favType, Page: page, PerPage: perPage,
		})
		if err != nil {
			return err
		}

		headers := []string{"ID", "Type", "Resource ID", "Created"}
		rows := make([][]string, len(resp.Items))
		for i, f := range resp.Items {
			rows[i] = []string{
				fmt.Sprintf("%d", f.ID), f.FavorableType,
				fmt.Sprintf("%d", f.FavorableID), f.CreatedAt,
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, &resp.Pagination))
		return nil
	},
}

var favoritesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a favorite",
	RunE: func(cmd *cobra.Command, args []string) error {
		favType, _ := cmd.Flags().GetString("type")
		favID, _ := cmd.Flags().GetInt64("id")

		if favType == "" || favID == 0 {
			return fmt.Errorf("--type (Track|Album|Artist) and --id are required")
		}

		client := mustClient()
		svc := api.NewFavoriteService(client)
		fav, err := svc.Add(cmdContext(), favType, favID)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", fav.ID)},
			{Key: "Type", Value: fav.FavorableType},
			{Key: "Resource ID", Value: fmt.Sprintf("%d", fav.FavorableID)},
		}))
		return nil
	},
}

var favoritesRemoveCmd = &cobra.Command{
	Use:   "remove [id]",
	Short: "Remove a favorite",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewFavoriteService(client)

		if len(args) > 0 {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid favorite ID: %s", args[0])
			}
			return svc.Remove(cmdContext(), id)
		}

		favType, _ := cmd.Flags().GetString("type")
		favID, _ := cmd.Flags().GetInt64("id")
		if favType == "" || favID == 0 {
			return fmt.Errorf("provide a favorite ID or --type and --id")
		}
		return svc.RemoveByTarget(cmdContext(), favType, favID)
	},
}

func init() {
	rootCmd.AddCommand(favoritesCmd)

	favoritesCmd.AddCommand(favoritesListCmd)
	favoritesListCmd.Flags().String("type", "", "filter by type: Track, Album, Artist")
	var page, perPage int
	addPaginationFlags(favoritesListCmd, &page, &perPage)

	favoritesCmd.AddCommand(favoritesAddCmd)
	favoritesAddCmd.Flags().String("type", "", "resource type: Track, Album, Artist")
	favoritesAddCmd.Flags().Int64("id", 0, "resource ID")

	favoritesCmd.AddCommand(favoritesRemoveCmd)
	favoritesRemoveCmd.Flags().String("type", "", "resource type")
	favoritesRemoveCmd.Flags().Int64("id", 0, "resource ID")
}
