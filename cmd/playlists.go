package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var playlistsCmd = &cobra.Command{
	Use:   "playlists",
	Short: "Manage playlists",
}

var playlistsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List playlists",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewPlaylistService(client)

		q, _ := cmd.Flags().GetString("q")
		sort, _ := cmd.Flags().GetString("sort")
		direction, _ := cmd.Flags().GetString("direction")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		resp, err := svc.List(cmdContext(), api.PlaylistListParams{
			Query: q, Sort: sort, Direction: direction, Page: page, PerPage: perPage,
		})
		if err != nil {
			return err
		}

		headers := []string{"ID", "Name", "Tracks", "Updated"}
		rows := make([][]string, len(resp.Items))
		for i, p := range resp.Items {
			rows[i] = []string{
				fmt.Sprintf("%d", p.ID), p.Name,
				fmt.Sprintf("%d", p.TracksCount), p.UpdatedAt,
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, &resp.Pagination))
		return nil
	},
}

var playlistsShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show playlist with tracks",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid playlist ID: %s", args[0])
		}

		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		client := mustClient()
		svc := api.NewPlaylistService(client)
		pl, err := svc.Get(cmdContext(), id, "", page, perPage)
		if err != nil {
			return err
		}

		fields := []output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", pl.ID)},
			{Key: "Name", Value: pl.Name},
			{Key: "Tracks", Value: fmt.Sprintf("%d", pl.TracksCount)},
			{Key: "Duration", Value: formatDuration(pl.TotalDuration)},
			{Key: "Created", Value: pl.CreatedAt},
		}
		fmt.Print(formatter.FormatItem(fields))

		if len(pl.Tracks) > 0 {
			fmt.Println()
			headers := []string{"#", "Track ID", "Title", "Artist", "Duration"}
			rows := make([][]string, len(pl.Tracks))
			for i, pt := range pl.Tracks {
				rows[i] = []string{
					fmt.Sprintf("%d", pt.Position),
					fmt.Sprintf("%d", pt.Track.ID),
					pt.Track.Title,
					pt.Track.Artist.Name,
					formatDuration(pt.Track.Duration),
				}
			}
			fmt.Print(formatter.FormatList(headers, rows, &pl.Pagination))
		}

		return nil
	},
}

var playlistsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a playlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		if name == "" {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Playlist name").Value(&name),
				),
			).WithTheme(huh.ThemeCharm())
			if err := form.Run(); err != nil {
				return err
			}
		}

		var trackIDs []int64
		if ids, _ := cmd.Flags().GetString("track-ids"); ids != "" {
			for _, s := range strings.Split(ids, ",") {
				id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
				if err != nil {
					return fmt.Errorf("invalid track ID: %s", s)
				}
				trackIDs = append(trackIDs, id)
			}
		}

		client := mustClient()
		svc := api.NewPlaylistService(client)
		pl, err := svc.Create(cmdContext(), name, trackIDs)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", pl.ID)},
			{Key: "Name", Value: pl.Name},
			{Key: "Tracks", Value: fmt.Sprintf("%d", pl.TracksCount)},
		}))
		return nil
	},
}

var playlistsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a playlist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid playlist ID: %s", args[0])
		}

		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return fmt.Errorf("--name is required")
		}

		client := mustClient()
		svc := api.NewPlaylistService(client)
		pl, err := svc.Update(cmdContext(), id, name)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", pl.ID)},
			{Key: "Name", Value: pl.Name},
		}))
		return nil
	},
}

var playlistsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a playlist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid playlist ID: %s", args[0])
		}

		var confirm bool
		huh.NewConfirm().Title(fmt.Sprintf("Delete playlist %d?", id)).Value(&confirm).Run()
		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}

		client := mustClient()
		svc := api.NewPlaylistService(client)
		if err := svc.Delete(cmdContext(), id); err != nil {
			return err
		}
		fmt.Println("Playlist deleted.")
		return nil
	},
}

var playlistsAddTrackCmd = &cobra.Command{
	Use:   "add-track <playlist-id>",
	Short: "Add track(s) to a playlist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid playlist ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewPlaylistService(client)

		if albumID, _ := cmd.Flags().GetInt64("album-id"); albumID > 0 {
			result, err := svc.AddAlbum(cmdContext(), playlistID, albumID)
			if err != nil {
				return err
			}
			fmt.Printf("Added %d tracks (total: %d)\n", result.Added, result.TracksCount)
			return nil
		}

		if ids, _ := cmd.Flags().GetString("track-ids"); ids != "" {
			var trackIDs []int64
			for _, s := range strings.Split(ids, ",") {
				id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
				if err != nil {
					return fmt.Errorf("invalid track ID: %s", s)
				}
				trackIDs = append(trackIDs, id)
			}
			result, err := svc.AddTracks(cmdContext(), playlistID, trackIDs)
			if err != nil {
				return err
			}
			fmt.Printf("Added %d tracks (total: %d)\n", result.Added, result.TracksCount)
			return nil
		}

		trackID, _ := cmd.Flags().GetInt64("track-id")
		if trackID == 0 {
			return fmt.Errorf("provide --track-id, --track-ids, or --album-id")
		}
		result, err := svc.AddTrack(cmdContext(), playlistID, trackID)
		if err != nil {
			return err
		}
		fmt.Printf("Added %d track(s) (total: %d)\n", result.Added, result.TracksCount)
		return nil
	},
}

var playlistsRemoveTrackCmd = &cobra.Command{
	Use:   "remove-track <playlist-id> <playlist-track-id>",
	Short: "Remove a track from a playlist",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid playlist ID: %s", args[0])
		}
		ptID, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid playlist track ID: %s", args[1])
		}

		client := mustClient()
		svc := api.NewPlaylistService(client)
		if err := svc.RemoveTrack(cmdContext(), playlistID, ptID); err != nil {
			return err
		}
		fmt.Println("Track removed from playlist.")
		return nil
	},
}

var playlistsReorderCmd = &cobra.Command{
	Use:   "reorder <playlist-id>",
	Short: "Reorder tracks in a playlist",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid playlist ID: %s", args[0])
		}

		idsStr, _ := cmd.Flags().GetString("ids")
		if idsStr == "" {
			return fmt.Errorf("--ids is required (comma-separated playlist track IDs)")
		}

		var ids []int64
		for _, s := range strings.Split(idsStr, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err != nil {
				return fmt.Errorf("invalid ID: %s", s)
			}
			ids = append(ids, id)
		}

		client := mustClient()
		svc := api.NewPlaylistService(client)
		if err := svc.Reorder(cmdContext(), playlistID, ids); err != nil {
			return err
		}
		fmt.Println("Tracks reordered.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(playlistsCmd)

	playlistsCmd.AddCommand(playlistsListCmd)
	playlistsListCmd.Flags().String("q", "", "search query")
	var page, perPage int
	var sort, direction string
	addPaginationFlags(playlistsListCmd, &page, &perPage)
	addSortFlags(playlistsListCmd, &sort, &direction)

	playlistsCmd.AddCommand(playlistsShowCmd)
	var showPage, showPerPage int
	addPaginationFlags(playlistsShowCmd, &showPage, &showPerPage)

	playlistsCmd.AddCommand(playlistsCreateCmd)
	playlistsCreateCmd.Flags().String("name", "", "playlist name")
	playlistsCreateCmd.Flags().String("track-ids", "", "comma-separated track IDs")

	playlistsCmd.AddCommand(playlistsUpdateCmd)
	playlistsUpdateCmd.Flags().String("name", "", "new name")

	playlistsCmd.AddCommand(playlistsDeleteCmd)

	playlistsCmd.AddCommand(playlistsAddTrackCmd)
	playlistsAddTrackCmd.Flags().Int64("track-id", 0, "single track ID")
	playlistsAddTrackCmd.Flags().String("track-ids", "", "comma-separated track IDs")
	playlistsAddTrackCmd.Flags().Int64("album-id", 0, "add all tracks from album")

	playlistsCmd.AddCommand(playlistsRemoveTrackCmd)

	playlistsCmd.AddCommand(playlistsReorderCmd)
	playlistsReorderCmd.Flags().String("ids", "", "comma-separated playlist track IDs in desired order")
}
