package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Set up API credentials",
	RunE:  runAuthLogin,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(authLoginCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ff2d95")).
		Bold(true)

	fmt.Println(titleStyle.Render("SYNTHWAVES.FM") + " -- connect your account")
	fmt.Println()

	var baseURL, clientID, secretKey string

	// Load existing config for defaults
	existing, _ := config.Load()
	if existing != nil {
		baseURL = existing.BaseURL
		clientID = existing.ClientID
	}

	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	theme := huh.ThemeCharm()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Base URL").
				Description("Your synthwaves.fm server URL").
				Value(&baseURL),

			huh.NewInput().
				Title("Client ID").
				Description("API key client_id (starts with bc_)").
				Value(&clientID),

			huh.NewInput().
				Title("Secret Key").
				Description("API key secret").
				EchoMode(huh.EchoModePassword).
				Value(&secretKey),
		),
	).WithTheme(theme)

	if err := form.Run(); err != nil {
		return err
	}

	cfg := &config.Config{
		BaseURL:   baseURL,
		ClientID:  clientID,
		SecretKey: secretKey,
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	// Test the credentials
	fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c8a")).Render("Testing credentials... "))
	client := api.NewClient(cfg)
	profile := api.NewProfileService(client)
	p, err := profile.Get(cmdContext())
	if err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444")).Render("failed"))
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#44ff44")).Render("ok"))
	fmt.Println()
	fmt.Printf("  Logged in as %s (%s)\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00fff2")).Bold(true).Render(p.Name),
		p.EmailAddress,
	)
	fmt.Printf("  Library: %d artists, %d albums, %d tracks\n",
		p.Stats.ArtistsCount, p.Stats.AlbumsCount, p.Stats.TracksCount,
	)

	// Save config
	if err := cfg.Save(); err != nil {
		fmt.Fprintln(os.Stderr, "Warning: could not save config:", err)
	} else {
		fmt.Println()
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c8a")).Render(
			"  Config saved to " + config.ConfigPath(),
		))
	}

	return nil
}
