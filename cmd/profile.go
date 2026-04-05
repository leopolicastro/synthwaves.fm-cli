package cmd

import (
	"fmt"

	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Show your profile",
	RunE:  runProfileShow,
}

var profileShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show your profile",
	RunE:  runProfileShow,
}

var profileUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update your profile",
	RunE: func(cmd *cobra.Command, args []string) error {
		fields := map[string]any{}
		if v, _ := cmd.Flags().GetString("name"); v != "" {
			fields["name"] = v
		}
		if v, _ := cmd.Flags().GetString("theme"); v != "" {
			fields["theme"] = v
		}

		if len(fields) == 0 {
			return fmt.Errorf("no fields to update (use --name, --theme)")
		}

		client := mustClient()
		svc := api.NewProfileService(client)
		profile, err := svc.Update(cmdContext(), fields)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "Name", Value: profile.Name},
			{Key: "Theme", Value: profile.Theme},
		}))
		return nil
	},
}

func runProfileShow(cmd *cobra.Command, args []string) error {
	client := mustClient()
	svc := api.NewProfileService(client)
	profile, err := svc.Get(cmdContext())
	if err != nil {
		return err
	}

	fmt.Print(formatter.FormatItem([]output.Field{
		{Key: "ID", Value: fmt.Sprintf("%d", profile.ID)},
		{Key: "Name", Value: profile.Name},
		{Key: "Email", Value: profile.EmailAddress},
		{Key: "Theme", Value: profile.Theme},
		{Key: "Artists", Value: fmt.Sprintf("%d", profile.Stats.ArtistsCount)},
		{Key: "Albums", Value: fmt.Sprintf("%d", profile.Stats.AlbumsCount)},
		{Key: "Tracks", Value: fmt.Sprintf("%d", profile.Stats.TracksCount)},
		{Key: "Playlists", Value: fmt.Sprintf("%d", profile.Stats.PlaylistsCount)},
		{Key: "Created", Value: profile.CreatedAt},
	}))
	return nil
}

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileShowCmd, profileUpdateCmd)

	profileUpdateCmd.Flags().String("name", "", "new display name")
	profileUpdateCmd.Flags().String("theme", "", "new theme")
}
