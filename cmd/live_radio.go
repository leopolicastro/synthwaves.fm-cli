package cmd

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/sixel"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/leo/synthwaves-cli/internal/player"
	"github.com/spf13/cobra"
)

var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "Public live radio stations",
}

var liveListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active live radio stations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewPublicRadioService(client)
		stations, err := svc.List(cmdContext())
		if err != nil {
			return err
		}

		headers := []string{"Slug", "Name", "Status", "Listeners", "Now Playing"}
		rows := make([][]string, len(stations))
		for i, s := range stations {
			nowPlaying := "-"
			if s.CurrentTrack != nil {
				nowPlaying = s.CurrentTrack.Title + " - " + s.CurrentTrack.Artist.Name
			}
			rows[i] = []string{
				s.Slug, s.Name, s.Status,
				fmt.Sprintf("%d", s.ListenerCount), nowPlaying,
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, nil))
		return nil
	},
}

var liveShowCmd = &cobra.Command{
	Use:   "show <slug>",
	Short: "Show live radio station details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewPublicRadioService(client)
		s, err := svc.Get(cmdContext(), args[0])
		if err != nil {
			return err
		}

		nowPlaying := "-"
		if s.CurrentTrack != nil {
			nowPlaying = s.CurrentTrack.Title + " - " + s.CurrentTrack.Artist.Name
		}

		fields := []output.Field{
			{Key: "Name", Value: s.Name},
			{Key: "Status", Value: s.Status},
			{Key: "Slug", Value: s.Slug},
			{Key: "Listeners", Value: fmt.Sprintf("%d", s.ListenerCount)},
			{Key: "Now Playing", Value: nowPlaying},
			{Key: "Stream URL", Value: s.ListenURL},
		}
		fmt.Print(formatter.FormatItem(fields))

		// Render cover art as sixel if image URL available
		if s.ImageURL != "" {
			renderSixelImage(s.ImageURL)
		}

		return nil
	},
}

var livePlayCmd = &cobra.Command{
	Use:   "play <slug>",
	Short: "Stream a live radio station",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewPublicRadioService(client)
		s, err := svc.Get(cmdContext(), args[0])
		if err != nil {
			return err
		}

		if s.ListenURL == "" {
			return fmt.Errorf("station has no stream URL")
		}

		nowPlaying := s.Name
		nowArtist := ""
		if s.CurrentTrack != nil {
			nowPlaying = s.CurrentTrack.Title
			nowArtist = s.CurrentTrack.Artist.Name
		}

		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff2d95")).Bold(true)
		artistStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00fff2"))
		mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c8a"))

		fmt.Println(titleStyle.Render("LIVE") + mutedStyle.Render(" streaming ") + titleStyle.Render(s.Name))
		if nowArtist != "" {
			fmt.Printf("  %s %s %s\n",
				titleStyle.Render(">>"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#e8e8f0")).Bold(true).Render(nowPlaying),
				artistStyle.Render("- "+nowArtist),
			)
		}
		fmt.Println(mutedStyle.Render("  Press Ctrl+C to stop"))
		fmt.Println()

		p := player.New()
		if err := p.Play(s.ListenURL, player.Track{
			Title: nowPlaying, Artist: nowArtist, Album: s.Name,
		}); err != nil {
			return err
		}

		// Wait for Ctrl+C
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig

		p.Cleanup()
		fmt.Println()
		fmt.Println(mutedStyle.Render("Stopped."))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(liveCmd)
	liveCmd.AddCommand(liveListCmd, liveShowCmd, livePlayCmd)
}

// renderSixelImage fetches an image URL and renders it as sixel graphics.
func renderSixelImage(url string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return
	}

	var buf bytes.Buffer
	var enc sixel.Encoder
	if err := enc.Encode(&buf, img); err != nil {
		return
	}

	fmt.Println()
	fmt.Print(ansi.SixelGraphics(0, 1, 0, buf.Bytes()))
	fmt.Println()
}
