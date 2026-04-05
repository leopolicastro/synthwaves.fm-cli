package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/config"
	"github.com/leo/synthwaves-cli/internal/logging"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var (
	cfgPath   string
	outFormat string
	verbose   bool

	cfg       *config.Config
	apiClient *api.Client
	formatter output.Formatter
)

var rootCmd = &cobra.Command{
	Use:   "synthwaves",
	Short: "synthwaves.fm CLI & TUI",
	Long: lipgloss.NewStyle().Foreground(lipgloss.Color("#ff2d95")).Bold(true).Render("SYNTHWAVES.FM") +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c8a")).Render(" -- manage your music library from the terminal"),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logging.SetVerbose(verbose)
		formatter = output.New(outFormat)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// No args -> launch TUI
		return runTUI(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "config file path (default: ~/.config/synthwaves/config.toml)")
	rootCmd.PersistentFlags().StringVarP(&outFormat, "format", "f", "table", "output format: table, json, text")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

// mustLoadConfig loads the config or exits with an error.
func mustLoadConfig() *config.Config {
	if cfg != nil {
		return cfg
	}
	var err error
	if cfgPath != "" {
		cfg, err = config.LoadFrom(cfgPath)
	} else {
		cfg, err = config.Load()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444")).Render(
			"Error loading config: "+err.Error(),
		))
		fmt.Fprintln(os.Stderr, lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c8a")).Render(
			"Run 'synthwaves auth login' to set up your credentials.",
		))
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444")).Render(
			"Invalid config: "+err.Error(),
		))
		os.Exit(1)
	}
	return cfg
}

// mustClient returns an authenticated API client or exits.
func mustClient() *api.Client {
	if apiClient != nil {
		return apiClient
	}
	c := mustLoadConfig()
	apiClient = api.NewClient(c)
	return apiClient
}

// cmdContext returns a context for CLI commands.
func cmdContext() context.Context {
	return context.Background()
}

// addPaginationFlags adds --page and --per-page flags to a command.
func addPaginationFlags(cmd *cobra.Command, page, perPage *int) {
	cmd.Flags().IntVar(page, "page", 1, "page number")
	cmd.Flags().IntVar(perPage, "per-page", 24, "items per page (max 100)")
}

// addSortFlags adds --sort and --direction flags to a command.
func addSortFlags(cmd *cobra.Command, sort, direction *string) {
	cmd.Flags().StringVar(sort, "sort", "", "sort column")
	cmd.Flags().StringVar(direction, "direction", "", "sort direction: asc, desc")
}

// addQueryFlag adds a --q flag for search/filter.
func addQueryFlag(cmd *cobra.Command, q *string) {
	cmd.Flags().StringVarP(q, "q", "q", "", "search query")
}
