package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search across artists, albums, and tracks",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		types, _ := cmd.Flags().GetString("types")
		genre, _ := cmd.Flags().GetString("genre")
		yearFrom, _ := cmd.Flags().GetInt("year-from")
		yearTo, _ := cmd.Flags().GetInt("year-to")
		limit, _ := cmd.Flags().GetInt("limit")
		favsOnly, _ := cmd.Flags().GetBool("favorites-only")
		category, _ := cmd.Flags().GetString("category")
		tags, _ := cmd.Flags().GetString("tags")

		client := mustClient()
		svc := api.NewSearchService(client)
		result, err := svc.Search(cmdContext(), api.SearchParams{
			Query: query, Types: types, Genre: genre,
			YearFrom: yearFrom, YearTo: yearTo, Limit: limit,
			FavoritesOnly: favsOnly, Category: category, Tags: tags,
		})
		if err != nil {
			return err
		}

		sectionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff2d95")).
			Bold(true).
			MarginTop(1)

		if len(result.Artists) > 0 {
			fmt.Println(sectionStyle.Render(fmt.Sprintf("Artists (%d)", len(result.Artists))))
			headers := []string{"ID", "Name"}
			rows := make([][]string, len(result.Artists))
			for i, a := range result.Artists {
				rows[i] = []string{fmt.Sprintf("%d", a.ID), a.Name}
			}
			fmt.Print(formatter.FormatList(headers, rows, nil))
		}

		if len(result.Albums) > 0 {
			fmt.Println(sectionStyle.Render(fmt.Sprintf("Albums (%d)", len(result.Albums))))
			headers := []string{"ID", "Title", "Artist", "Year", "Genre"}
			rows := make([][]string, len(result.Albums))
			for i, a := range result.Albums {
				rows[i] = []string{
					fmt.Sprintf("%d", a.ID), a.Title, a.Artist.Name,
					fmt.Sprintf("%d", a.Year), a.Genre,
				}
			}
			fmt.Print(formatter.FormatList(headers, rows, nil))
		}

		if len(result.Tracks) > 0 {
			fmt.Println(sectionStyle.Render(fmt.Sprintf("Tracks (%d)", len(result.Tracks))))
			headers := []string{"ID", "Title", "Artist", "Album", "Duration"}
			rows := make([][]string, len(result.Tracks))
			for i, t := range result.Tracks {
				rows[i] = []string{
					fmt.Sprintf("%d", t.ID), t.Title,
					t.Artist.Name, t.Album.Title,
					formatDuration(t.Duration),
				}
			}
			fmt.Print(formatter.FormatList(headers, rows, nil))
		}

		total := len(result.Artists) + len(result.Albums) + len(result.Tracks)
		if total == 0 {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c8a")).Render("No results found."))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().String("types", "", "comma-separated: artist,album,track")
	searchCmd.Flags().String("genre", "", "filter by genre")
	searchCmd.Flags().Int("year-from", 0, "minimum year")
	searchCmd.Flags().Int("year-to", 0, "maximum year")
	searchCmd.Flags().Int("limit", 0, "max results per type (max 50)")
	searchCmd.Flags().Bool("favorites-only", false, "only search favorites")
	searchCmd.Flags().String("category", "", "filter by category")
	searchCmd.Flags().String("tags", "", "comma-separated tag names")
}
