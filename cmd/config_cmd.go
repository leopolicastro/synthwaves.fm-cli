package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := mustLoadConfig()

		keyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00fff2")).
			Bold(true).
			Width(12)
		valStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e8e8f0"))

		fmt.Println(keyStyle.Render("base_url") + valStyle.Render(cfg.BaseURL))
		fmt.Println(keyStyle.Render("client_id") + valStyle.Render(cfg.ClientID))
		fmt.Println(keyStyle.Render("secret_key") + valStyle.Render("********"))
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c8a")).Render(
			"Config file: " + config.ConfigPath(),
		))
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive configuration setup",
	RunE:  runAuthLogin, // Same flow as auth login
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := mustLoadConfig()
		key, value := args[0], args[1]

		switch key {
		case "base_url":
			cfg.BaseURL = value
		case "client_id":
			cfg.ClientID = value
		case "secret_key":
			cfg.SecretKey = value
		default:
			return fmt.Errorf("unknown config key: %s (valid: base_url, client_id, secret_key)", key)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("Set %s\n", key)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd, configInitCmd, configSetCmd)
}
