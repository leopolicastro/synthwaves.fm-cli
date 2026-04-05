package cmd

import (
	"fmt"
	"strconv"

	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Manage play history",
}

var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List play history",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewPlayHistoryService(client)

		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		resp, err := svc.List(cmdContext(), "", page, perPage)
		if err != nil {
			return err
		}

		headers := []string{"ID", "Track", "Artist", "Played At"}
		rows := make([][]string, len(resp.Items))
		for i, ph := range resp.Items {
			rows[i] = []string{
				fmt.Sprintf("%d", ph.ID), ph.Track.Title,
				ph.Track.Artist.Name, ph.PlayedAt,
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, &resp.Pagination))
		return nil
	},
}

var historyRecordCmd = &cobra.Command{
	Use:   "record <track-id>",
	Short: "Record a play event",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		trackID, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid track ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewPlayHistoryService(client)
		ph, err := svc.Record(cmdContext(), trackID)
		if err != nil {
			return err
		}

		fmt.Printf("Recorded play: %s - %s\n", ph.Track.Title, ph.Track.Artist.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.AddCommand(historyListCmd)
	var page, perPage int
	addPaginationFlags(historyListCmd, &page, &perPage)

	historyCmd.AddCommand(historyRecordCmd)
}
