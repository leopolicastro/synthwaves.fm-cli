package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var artistsCmd = &cobra.Command{
	Use:   "artists",
	Short: "Manage artists",
}

var artistsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List artists",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewArtistService(client)

		q, _ := cmd.Flags().GetString("q")
		category, _ := cmd.Flags().GetString("category")
		sort, _ := cmd.Flags().GetString("sort")
		direction, _ := cmd.Flags().GetString("direction")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		resp, err := svc.List(cmdContext(), api.ArtistListParams{
			Query: q, Category: category, Sort: sort, Direction: direction,
			Page: page, PerPage: perPage,
		})
		if err != nil {
			return err
		}

		headers := []string{"ID", "Name", "Category", "Albums", "Tracks"}
		rows := make([][]string, len(resp.Items))
		for i, a := range resp.Items {
			rows[i] = []string{
				fmt.Sprintf("%d", a.ID), a.Name, a.Category,
				fmt.Sprintf("%d", a.AlbumsCount), fmt.Sprintf("%d", a.TracksCount),
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, &resp.Pagination))
		return nil
	},
}

var artistsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show artist details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid artist ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewArtistService(client)
		artist, err := svc.Get(cmdContext(), id)
		if err != nil {
			return err
		}

		fields := []output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", artist.ID)},
			{Key: "Name", Value: artist.Name},
			{Key: "Category", Value: artist.Category},
			{Key: "Albums", Value: fmt.Sprintf("%d", artist.AlbumsCount)},
			{Key: "Tracks", Value: fmt.Sprintf("%d", artist.TracksCount)},
			{Key: "Created", Value: artist.CreatedAt},
		}
		fmt.Print(formatter.FormatItem(fields))

		if len(artist.Albums) > 0 {
			fmt.Println()
			headers := []string{"ID", "Title", "Year", "Genre", "Tracks"}
			rows := make([][]string, len(artist.Albums))
			for i, a := range artist.Albums {
				rows[i] = []string{
					fmt.Sprintf("%d", a.ID), a.Title,
					fmt.Sprintf("%d", a.Year), a.Genre,
					fmt.Sprintf("%d", a.TracksCount),
				}
			}
			fmt.Print(formatter.FormatList(headers, rows, nil))
		}

		return nil
	},
}

var artistsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an artist",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		category, _ := cmd.Flags().GetString("category")

		// Interactive form if flags are missing
		if name == "" {
			if category == "" {
				category = "music"
			}
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Artist name").Value(&name),
					huh.NewSelect[string]().
						Title("Category").
						Options(
							huh.NewOption("Music", "music"),
							huh.NewOption("Podcast", "podcast"),
						).
						Value(&category),
				),
			).WithTheme(huh.ThemeCharm())
			if err := form.Run(); err != nil {
				return err
			}
		}

		if category == "" {
			category = "music"
		}

		client := mustClient()
		svc := api.NewArtistService(client)
		artist, err := svc.Create(cmdContext(), name, category)
		if err != nil {
			return err
		}

		fields := []output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", artist.ID)},
			{Key: "Name", Value: artist.Name},
			{Key: "Category", Value: artist.Category},
		}
		fmt.Print(formatter.FormatItem(fields))
		return nil
	},
}

var artistsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an artist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid artist ID: %s", args[0])
		}

		name, _ := cmd.Flags().GetString("name")
		category, _ := cmd.Flags().GetString("category")

		client := mustClient()
		svc := api.NewArtistService(client)
		artist, err := svc.Update(cmdContext(), id, name, category)
		if err != nil {
			return err
		}

		fields := []output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", artist.ID)},
			{Key: "Name", Value: artist.Name},
			{Key: "Category", Value: artist.Category},
		}
		fmt.Print(formatter.FormatItem(fields))
		return nil
	},
}

var artistsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an artist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid artist ID: %s", args[0])
		}

		var confirm bool
		huh.NewConfirm().
			Title(fmt.Sprintf("Delete artist %d?", id)).
			Description("This cannot be undone.").
			Value(&confirm).
			Run()

		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}

		client := mustClient()
		svc := api.NewArtistService(client)
		if err := svc.Delete(cmdContext(), id); err != nil {
			return err
		}
		fmt.Println("Artist deleted.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(artistsCmd)

	artistsCmd.AddCommand(artistsListCmd)
	artistsListCmd.Flags().String("q", "", "search query")
	artistsListCmd.Flags().String("category", "", "filter by category")
	var page, perPage int
	var sort, direction string
	addPaginationFlags(artistsListCmd, &page, &perPage)
	addSortFlags(artistsListCmd, &sort, &direction)

	artistsCmd.AddCommand(artistsShowCmd)

	artistsCmd.AddCommand(artistsCreateCmd)
	artistsCreateCmd.Flags().String("name", "", "artist name")
	artistsCreateCmd.Flags().String("category", "", "category: music, podcast")

	artistsCmd.AddCommand(artistsUpdateCmd)
	artistsUpdateCmd.Flags().String("name", "", "new name")
	artistsUpdateCmd.Flags().String("category", "", "new category")

	artistsCmd.AddCommand(artistsDeleteCmd)
}
