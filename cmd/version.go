package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		logo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff2d95")).
			Bold(true).
			Render("SYNTHWAVES")

		dot := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#b026ff")).
			Render(".")

		fm := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00fff2")).
			Bold(true).
			Render("FM")

		ver := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6c6c8a")).
			Render(" v" + Version)

		fmt.Println(logo + dot + fm + ver)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
