package cmd

import (
	"fmt"

	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List tags",
}

var tagsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := mustClient()
		svc := api.NewTagService(client)

		tagType, _ := cmd.Flags().GetString("type")
		q, _ := cmd.Flags().GetString("q")

		tags, err := svc.List(cmdContext(), tagType, q)
		if err != nil {
			return err
		}

		headers := []string{"ID", "Name", "Type"}
		rows := make([][]string, len(tags))
		for i, t := range tags {
			rows[i] = []string{
				fmt.Sprintf("%d", t.ID), t.Name, t.TagType,
			}
		}

		fmt.Print(formatter.FormatList(headers, rows, nil))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
	tagsCmd.AddCommand(tagsListCmd)
	tagsListCmd.Flags().String("type", "", "filter by tag type")
	tagsListCmd.Flags().String("q", "", "search by name")
}
