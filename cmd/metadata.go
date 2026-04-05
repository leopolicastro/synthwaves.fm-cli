package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/spf13/cobra"
)

var metadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Show available genres, languages, and decades",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewTrackMetadataService(client)
		meta, err := svc.Get(cmdContext())
		if err != nil {
			return err
		}

		sectionStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff2d95")).
			Bold(true)

		fmt.Println(sectionStyle.Render("Genres"))
		fmt.Println(strings.Join(meta.Genres, ", "))
		fmt.Println()

		fmt.Println(sectionStyle.Render("Languages"))
		fmt.Println(strings.Join(meta.Languages, ", "))
		fmt.Println()

		fmt.Println(sectionStyle.Render("Decades"))
		fmt.Println(strings.Join(meta.Decades, ", "))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(metadataCmd)
}
