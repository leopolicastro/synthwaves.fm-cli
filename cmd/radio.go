package cmd

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var radioCmd = &cobra.Command{
	Use:   "radio",
	Short: "Manage radio stations",
}

var radioListCmd = &cobra.Command{
	Use:   "list",
	Short: "List radio stations",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewRadioStationService(client)
		stations, err := svc.List(cmdContext())
		if err != nil {
			return err
		}

		headers := []string{"ID", "Name", "Status", "Mode", "Bitrate", "Playlist"}
		rows := make([][]string, len(stations))
		for i, s := range stations {
			status := s.Status
			switch status {
			case "active":
				status = lipgloss.NewStyle().Foreground(lipgloss.Color("#44ff44")).Render("LIVE")
			case "stopped":
				status = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444")).Render("OFF")
			}
			rows[i] = []string{
				fmt.Sprintf("%d", s.ID), s.Name, status,
				s.PlaybackMode, fmt.Sprintf("%d", s.Bitrate),
				s.Playlist.Name,
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, nil))
		return nil
	},
}

var radioShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show radio station details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid station ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewRadioStationService(client)
		s, err := svc.Get(cmdContext(), id)
		if err != nil {
			return err
		}

		fields := []output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", s.ID)},
			{Key: "Name", Value: s.Name},
			{Key: "Status", Value: s.Status},
			{Key: "Playlist", Value: s.Playlist.Name},
			{Key: "Mode", Value: s.PlaybackMode},
			{Key: "Bitrate", Value: fmt.Sprintf("%d kbps", s.Bitrate)},
			{Key: "Crossfade", Value: fmt.Sprintf("%.1fs", s.CrossfadeDuration)},
			{Key: "Listen URL", Value: s.ListenURL},
			{Key: "Mount", Value: s.MountPoint},
		}
		if s.CurrentTrack != nil {
			fields = append(fields, output.Field{
				Key:   "Now Playing",
				Value: fmt.Sprintf("%s - %s", s.CurrentTrack.Title, s.CurrentTrack.Artist.Name),
			})
		}

		fmt.Print(formatter.FormatItem(fields))
		return nil
	},
}

var radioCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a radio station",
	RunE: func(cmd *cobra.Command, args []string) error {
		playlistID, _ := cmd.Flags().GetInt64("playlist-id")
		if playlistID == 0 {
			return fmt.Errorf("--playlist-id is required")
		}

		client := mustClient()
		svc := api.NewRadioStationService(client)
		s, err := svc.Create(cmdContext(), playlistID)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", s.ID)},
			{Key: "Name", Value: s.Name},
			{Key: "Status", Value: s.Status},
			{Key: "Listen URL", Value: s.ListenURL},
		}))
		return nil
	},
}

var radioUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update radio station settings",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid station ID: %s", args[0])
		}

		fields := map[string]any{}
		if v, _ := cmd.Flags().GetString("playback-mode"); v != "" {
			fields["playback_mode"] = v
		}
		if v, _ := cmd.Flags().GetInt("bitrate"); v > 0 {
			fields["bitrate"] = v
		}
		if v, _ := cmd.Flags().GetFloat64("crossfade-duration"); v > 0 {
			fields["crossfade_duration"] = v
		}

		if len(fields) == 0 {
			return fmt.Errorf("no fields to update (use --playback-mode, --bitrate, --crossfade-duration)")
		}

		client := mustClient()
		svc := api.NewRadioStationService(client)
		s, err := svc.Update(cmdContext(), id, fields)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", s.ID)},
			{Key: "Name", Value: s.Name},
			{Key: "Mode", Value: s.PlaybackMode},
			{Key: "Bitrate", Value: fmt.Sprintf("%d", s.Bitrate)},
		}))
		return nil
	},
}

var radioDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a radio station",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid station ID: %s", args[0])
		}

		var confirm bool
		huh.NewConfirm().Title(fmt.Sprintf("Delete radio station %d?", id)).Value(&confirm).Run()
		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}

		client := mustClient()
		svc := api.NewRadioStationService(client)
		if err := svc.Delete(cmdContext(), id); err != nil {
			return err
		}
		fmt.Println("Radio station deleted.")
		return nil
	},
}

var radioStartCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "Start a radio station",
	Args:  cobra.ExactArgs(1),
	RunE:  radioControl("start"),
}

var radioStopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "Stop a radio station",
	Args:  cobra.ExactArgs(1),
	RunE:  radioControl("stop"),
}

var radioSkipCmd = &cobra.Command{
	Use:   "skip <id>",
	Short: "Skip current track",
	Args:  cobra.ExactArgs(1),
	RunE:  radioControl("skip"),
}

func radioControl(action string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid station ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewRadioStationService(client)
		result, err := svc.Control(cmdContext(), id, action)
		if err != nil {
			return err
		}

		fmt.Printf("%s: %s\n", result.Status, result.Message)
		return nil
	}
}

func init() {
	rootCmd.AddCommand(radioCmd)
	radioCmd.AddCommand(radioListCmd, radioShowCmd, radioCreateCmd, radioUpdateCmd, radioDeleteCmd)
	radioCmd.AddCommand(radioStartCmd, radioStopCmd, radioSkipCmd)

	radioCreateCmd.Flags().Int64("playlist-id", 0, "playlist ID for the station")
	radioUpdateCmd.Flags().String("playback-mode", "", "playback mode")
	radioUpdateCmd.Flags().Int("bitrate", 0, "bitrate in kbps")
	radioUpdateCmd.Flags().Float64("crossfade-duration", 0, "crossfade duration in seconds")
}
