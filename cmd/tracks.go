package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var tracksCmd = &cobra.Command{
	Use:   "tracks",
	Short: "Manage tracks",
}

var tracksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracks",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewTrackService(client)

		q, _ := cmd.Flags().GetString("q")
		albumID, _ := cmd.Flags().GetInt64("album-id")
		artistID, _ := cmd.Flags().GetInt64("artist-id")
		genre, _ := cmd.Flags().GetString("genre")
		language, _ := cmd.Flags().GetString("language")
		decade, _ := cmd.Flags().GetString("decade")
		sort, _ := cmd.Flags().GetString("sort")
		direction, _ := cmd.Flags().GetString("direction")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		resp, err := svc.List(cmdContext(), api.TrackListParams{
			Query: q, AlbumID: albumID, ArtistID: artistID,
			Genre: genre, Language: language, Decade: decade,
			Sort: sort, Direction: direction, Page: page, PerPage: perPage,
		})
		if err != nil {
			return err
		}

		headers := []string{"ID", "Title", "Artist", "Album", "Duration", "Format"}
		rows := make([][]string, len(resp.Items))
		for i, t := range resp.Items {
			rows[i] = []string{
				fmt.Sprintf("%d", t.ID), t.Title,
				t.Artist.Name, t.Album.Title,
				formatDuration(t.Duration), t.FileFormat,
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, &resp.Pagination))
		return nil
	},
}

var tracksShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show track details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid track ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewTrackService(client)
		track, err := svc.Get(cmdContext(), id)
		if err != nil {
			return err
		}

		fields := []output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", track.ID)},
			{Key: "Title", Value: track.Title},
			{Key: "Artist", Value: track.Artist.Name},
			{Key: "Album", Value: track.Album.Title},
			{Key: "Track #", Value: fmt.Sprintf("%d", track.TrackNumber)},
			{Key: "Disc #", Value: fmt.Sprintf("%d", track.DiscNumber)},
			{Key: "Duration", Value: formatDuration(track.Duration)},
			{Key: "Bitrate", Value: fmt.Sprintf("%d kbps", track.Bitrate)},
			{Key: "Format", Value: track.FileFormat},
			{Key: "Size", Value: formatFileSize(track.FileSize)},
			{Key: "Has Audio", Value: fmt.Sprintf("%v", track.HasAudio)},
			{Key: "Created", Value: track.CreatedAt},
		}
		fmt.Print(formatter.FormatItem(fields))
		return nil
	},
}

var tracksCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a track",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		artistID, _ := cmd.Flags().GetInt64("artist-id")
		albumID, _ := cmd.Flags().GetInt64("album-id")
		trackNum, _ := cmd.Flags().GetInt("track-number")

		if title == "" || artistID == 0 {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Track title").Value(&title),
					huh.NewInput().Title("Artist ID").Value(ptrStr(fmt.Sprintf("%d", artistID))),
					huh.NewInput().Title("Album ID (optional)").Value(ptrStr(fmt.Sprintf("%d", albumID))),
				),
			).WithTheme(huh.ThemeCharm())
			if err := form.Run(); err != nil {
				return err
			}
		}

		fields := map[string]any{
			"title":     title,
			"artist_id": artistID,
		}
		if albumID > 0 {
			fields["album_id"] = albumID
		}
		if trackNum > 0 {
			fields["track_number"] = trackNum
		}

		client := mustClient()
		svc := api.NewTrackService(client)
		track, err := svc.Create(cmdContext(), fields)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", track.ID)},
			{Key: "Title", Value: track.Title},
			{Key: "Artist", Value: track.Artist.Name},
		}))
		return nil
	},
}

var tracksUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a track",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid track ID: %s", args[0])
		}

		fields := map[string]any{}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			fields["title"] = v
		}
		if v, _ := cmd.Flags().GetInt64("artist-id"); v > 0 {
			fields["artist_id"] = v
		}
		if v, _ := cmd.Flags().GetInt64("album-id"); v > 0 {
			fields["album_id"] = v
		}
		if v, _ := cmd.Flags().GetInt("track-number"); v > 0 {
			fields["track_number"] = v
		}

		if len(fields) == 0 {
			return fmt.Errorf("no fields to update")
		}

		client := mustClient()
		svc := api.NewTrackService(client)
		track, err := svc.Update(cmdContext(), id, fields)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", track.ID)},
			{Key: "Title", Value: track.Title},
			{Key: "Artist", Value: track.Artist.Name},
		}))
		return nil
	},
}

var tracksDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a track",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid track ID: %s", args[0])
		}

		var confirm bool
		huh.NewConfirm().
			Title(fmt.Sprintf("Delete track %d?", id)).
			Value(&confirm).
			Run()
		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}

		client := mustClient()
		svc := api.NewTrackService(client)
		if err := svc.Delete(cmdContext(), id); err != nil {
			return err
		}
		fmt.Println("Track deleted.")
		return nil
	},
}

var tracksStreamCmd = &cobra.Command{
	Use:   "stream <id>",
	Short: "Get stream URL for a track",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid track ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewTrackService(client)
		info, err := svc.Stream(cmdContext(), id)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "URL", Value: info.URL},
			{Key: "Type", Value: info.ContentType},
			{Key: "Size", Value: formatFileSize(info.FileSize)},
		}))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tracksCmd)

	tracksCmd.AddCommand(tracksListCmd)
	tracksListCmd.Flags().String("q", "", "search query")
	tracksListCmd.Flags().Int64("album-id", 0, "filter by album ID")
	tracksListCmd.Flags().Int64("artist-id", 0, "filter by artist ID")
	tracksListCmd.Flags().String("genre", "", "filter by genre")
	tracksListCmd.Flags().String("language", "", "filter by language")
	tracksListCmd.Flags().String("decade", "", "filter by decade")
	var page, perPage int
	var sort, direction string
	addPaginationFlags(tracksListCmd, &page, &perPage)
	addSortFlags(tracksListCmd, &sort, &direction)

	tracksCmd.AddCommand(tracksShowCmd)

	tracksCmd.AddCommand(tracksCreateCmd)
	tracksCreateCmd.Flags().String("title", "", "track title")
	tracksCreateCmd.Flags().Int64("artist-id", 0, "artist ID")
	tracksCreateCmd.Flags().Int64("album-id", 0, "album ID")
	tracksCreateCmd.Flags().Int("track-number", 0, "track number")

	tracksCmd.AddCommand(tracksUpdateCmd)
	tracksUpdateCmd.Flags().String("title", "", "new title")
	tracksUpdateCmd.Flags().Int64("artist-id", 0, "new artist ID")
	tracksUpdateCmd.Flags().Int64("album-id", 0, "new album ID")
	tracksUpdateCmd.Flags().Int("track-number", 0, "new track number")

	tracksCmd.AddCommand(tracksDeleteCmd)
	tracksCmd.AddCommand(tracksStreamCmd)
}

func formatFileSize(bytes int64) string {
	if bytes == 0 {
		return "-"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
