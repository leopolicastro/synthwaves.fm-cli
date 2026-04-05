package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var albumsCmd = &cobra.Command{
	Use:   "albums",
	Short: "Manage albums",
}

var albumsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List albums",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewAlbumService(client)

		q, _ := cmd.Flags().GetString("q")
		artistID, _ := cmd.Flags().GetInt64("artist-id")
		sort, _ := cmd.Flags().GetString("sort")
		direction, _ := cmd.Flags().GetString("direction")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		resp, err := svc.List(cmdContext(), api.AlbumListParams{
			Query: q, ArtistID: artistID, Sort: sort, Direction: direction,
			Page: page, PerPage: perPage,
		})
		if err != nil {
			return err
		}

		headers := []string{"ID", "Title", "Artist", "Year", "Genre", "Tracks"}
		rows := make([][]string, len(resp.Items))
		for i, a := range resp.Items {
			rows[i] = []string{
				fmt.Sprintf("%d", a.ID), a.Title, a.Artist.Name,
				fmt.Sprintf("%d", a.Year), a.Genre,
				fmt.Sprintf("%d", a.TracksCount),
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, &resp.Pagination))
		return nil
	},
}

var albumsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show album details with tracks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid album ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewAlbumService(client)
		album, err := svc.Get(cmdContext(), id)
		if err != nil {
			return err
		}

		fields := []output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", album.ID)},
			{Key: "Title", Value: album.Title},
			{Key: "Artist", Value: album.Artist.Name},
			{Key: "Year", Value: fmt.Sprintf("%d", album.Year)},
			{Key: "Genre", Value: album.Genre},
			{Key: "Tracks", Value: fmt.Sprintf("%d", album.TracksCount)},
			{Key: "Duration", Value: formatDuration(album.TotalDuration)},
			{Key: "Created", Value: album.CreatedAt},
		}
		fmt.Print(formatter.FormatItem(fields))

		if len(album.Tracks) > 0 {
			fmt.Println()
			headers := []string{"#", "Title", "Duration", "Format"}
			rows := make([][]string, len(album.Tracks))
			for i, t := range album.Tracks {
				rows[i] = []string{
					fmt.Sprintf("%d", t.TrackNumber), t.Title,
					formatDuration(t.Duration), t.FileFormat,
				}
			}
			fmt.Print(formatter.FormatList(headers, rows, nil))
		}

		return nil
	},
}

var albumsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an album",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		artistID, _ := cmd.Flags().GetInt64("artist-id")
		year, _ := cmd.Flags().GetInt("year")
		genre, _ := cmd.Flags().GetString("genre")

		if title == "" || artistID == 0 {
			var yearStr string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Album title").Value(&title),
					huh.NewInput().Title("Artist ID").Value(ptrStr(fmt.Sprintf("%d", artistID))),
					huh.NewInput().Title("Year").Value(&yearStr),
					huh.NewInput().Title("Genre").Value(&genre),
				),
			).WithTheme(huh.ThemeCharm())
			if err := form.Run(); err != nil {
				return err
			}
			if yearStr != "" {
				year, _ = strconv.Atoi(yearStr)
			}
		}

		client := mustClient()
		svc := api.NewAlbumService(client)
		album, err := svc.Create(cmdContext(), title, artistID, year, genre)
		if err != nil {
			return err
		}

		fields := []output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", album.ID)},
			{Key: "Title", Value: album.Title},
			{Key: "Artist", Value: album.Artist.Name},
			{Key: "Year", Value: fmt.Sprintf("%d", album.Year)},
			{Key: "Genre", Value: album.Genre},
		}
		fmt.Print(formatter.FormatItem(fields))
		return nil
	},
}

var albumsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an album",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid album ID: %s", args[0])
		}

		fields := map[string]any{}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			fields["title"] = v
		}
		if v, _ := cmd.Flags().GetInt64("artist-id"); v > 0 {
			fields["artist_id"] = v
		}
		if v, _ := cmd.Flags().GetInt("year"); v > 0 {
			fields["year"] = v
		}
		if v, _ := cmd.Flags().GetString("genre"); v != "" {
			fields["genre"] = v
		}

		if len(fields) == 0 {
			return fmt.Errorf("no fields to update (use --title, --artist-id, --year, --genre)")
		}

		client := mustClient()
		svc := api.NewAlbumService(client)
		album, err := svc.Update(cmdContext(), id, fields)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", album.ID)},
			{Key: "Title", Value: album.Title},
			{Key: "Artist", Value: album.Artist.Name},
			{Key: "Year", Value: fmt.Sprintf("%d", album.Year)},
			{Key: "Genre", Value: album.Genre},
		}))
		return nil
	},
}

var albumsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an album",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid album ID: %s", args[0])
		}

		var confirm bool
		huh.NewConfirm().
			Title(fmt.Sprintf("Delete album %d?", id)).
			Description("This cannot be undone.").
			Value(&confirm).
			Run()

		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}

		client := mustClient()
		svc := api.NewAlbumService(client)
		if err := svc.Delete(cmdContext(), id); err != nil {
			return err
		}
		fmt.Println("Album deleted.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(albumsCmd)

	albumsCmd.AddCommand(albumsListCmd)
	albumsListCmd.Flags().String("q", "", "search query")
	albumsListCmd.Flags().Int64("artist-id", 0, "filter by artist ID")
	var page, perPage int
	var sort, direction string
	addPaginationFlags(albumsListCmd, &page, &perPage)
	addSortFlags(albumsListCmd, &sort, &direction)

	albumsCmd.AddCommand(albumsShowCmd)

	albumsCmd.AddCommand(albumsCreateCmd)
	albumsCreateCmd.Flags().String("title", "", "album title")
	albumsCreateCmd.Flags().Int64("artist-id", 0, "artist ID")
	albumsCreateCmd.Flags().Int("year", 0, "release year")
	albumsCreateCmd.Flags().String("genre", "", "genre")

	albumsCmd.AddCommand(albumsUpdateCmd)
	albumsUpdateCmd.Flags().String("title", "", "new title")
	albumsUpdateCmd.Flags().Int64("artist-id", 0, "new artist ID")
	albumsUpdateCmd.Flags().Int("year", 0, "new year")
	albumsUpdateCmd.Flags().String("genre", "", "new genre")

	albumsCmd.AddCommand(albumsDeleteCmd)
}

func formatDuration(seconds float64) string {
	m := int(seconds) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

func ptrStr(s string) *string { return &s }
