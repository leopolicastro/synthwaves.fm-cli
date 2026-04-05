package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show listening statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeRange, _ := cmd.Flags().GetString("time-range")

		client := mustClient()
		svc := api.NewStatsService(client)
		stats, err := svc.Get(cmdContext(), timeRange)
		if err != nil {
			return err
		}

		sectionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff2d95")).
			Bold(true)

		fmt.Println(sectionStyle.Render("Library"))
		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "Artists", Value: fmt.Sprintf("%d", stats.Library.ArtistsCount)},
			{Key: "Albums", Value: fmt.Sprintf("%d", stats.Library.AlbumsCount)},
			{Key: "Tracks", Value: fmt.Sprintf("%d", stats.Library.TracksCount)},
			{Key: "Playlists", Value: fmt.Sprintf("%d", stats.Library.PlaylistsCount)},
			{Key: "Duration", Value: formatDuration(stats.Library.TotalDuration)},
			{Key: "Size", Value: formatFileSize(stats.Library.TotalFileSize)},
		}))
		fmt.Println()

		fmt.Println(sectionStyle.Render(fmt.Sprintf("Listening (%s)", stats.Listening.TimeRange)))
		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "Total Plays", Value: fmt.Sprintf("%d", stats.Listening.TotalPlays)},
			{Key: "Listen Time", Value: formatDuration(stats.Listening.TotalListeningTime)},
			{Key: "Streak", Value: fmt.Sprintf("%d days", stats.Listening.CurrentStreak)},
			{Key: "Best Streak", Value: fmt.Sprintf("%d days", stats.Listening.LongestStreak)},
		}))

		if len(stats.Listening.TopTracks) > 0 {
			fmt.Println()
			fmt.Println(sectionStyle.Render("Top Tracks"))
			headers := []string{"#", "Title", "Artist", "Plays"}
			rows := make([][]string, len(stats.Listening.TopTracks))
			for i, t := range stats.Listening.TopTracks {
				rows[i] = []string{
					fmt.Sprintf("%d", i+1), t.Title, t.ArtistName,
					fmt.Sprintf("%d", t.PlayCount),
				}
			}
			fmt.Print(formatter.FormatList(headers, rows, nil))
		}

		if len(stats.Listening.TopArtists) > 0 {
			fmt.Println()
			fmt.Println(sectionStyle.Render("Top Artists"))
			headers := []string{"#", "Artist", "Plays"}
			rows := make([][]string, len(stats.Listening.TopArtists))
			for i, a := range stats.Listening.TopArtists {
				rows[i] = []string{
					fmt.Sprintf("%d", i+1), a.Name,
					fmt.Sprintf("%d", a.PlayCount),
				}
			}
			fmt.Print(formatter.FormatList(headers, rows, nil))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().String("time-range", "month", "time range: week, month, year, all_time")
}
