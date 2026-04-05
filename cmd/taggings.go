package cmd

import (
	"fmt"
	"strconv"

	"github.com/leo/synthwaves-cli/internal/api"
	"github.com/leo/synthwaves-cli/internal/output"
	"github.com/spf13/cobra"
)

var taggingsCmd = &cobra.Command{
	Use:   "taggings",
	Short: "Manage tag associations",
}

var taggingsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Tag a resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		tagType, _ := cmd.Flags().GetString("tag-type")
		taggableType, _ := cmd.Flags().GetString("taggable-type")
		taggableID, _ := cmd.Flags().GetInt64("taggable-id")

		if name == "" || tagType == "" || taggableType == "" || taggableID == 0 {
			return fmt.Errorf("all flags required: --name, --tag-type, --taggable-type, --taggable-id")
		}

		client := mustClient()
		svc := api.NewTaggingService(client)
		tagging, err := svc.Create(cmdContext(), name, tagType, taggableType, taggableID)
		if err != nil {
			return err
		}

		fmt.Print(formatter.FormatItem([]output.Field{
			{Key: "ID", Value: fmt.Sprintf("%d", tagging.ID)},
			{Key: "Tag", Value: tagging.Tag.Name},
			{Key: "Type", Value: tagging.Tag.TagType},
			{Key: "Resource", Value: fmt.Sprintf("%s #%d", tagging.TaggableType, tagging.TaggableID)},
		}))
		return nil
	},
}

var taggingsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove a tag association",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid tagging ID: %s", args[0])
		}

		client := mustClient()
		svc := api.NewTaggingService(client)
		if err := svc.Delete(cmdContext(), id); err != nil {
			return err
		}
		fmt.Println("Tagging removed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(taggingsCmd)
	taggingsCmd.AddCommand(taggingsCreateCmd, taggingsDeleteCmd)

	taggingsCreateCmd.Flags().String("name", "", "tag name")
	taggingsCreateCmd.Flags().String("tag-type", "", "tag type (e.g., genre, mood)")
	taggingsCreateCmd.Flags().String("taggable-type", "", "resource type: Track, Album, Artist")
	taggingsCreateCmd.Flags().Int64("taggable-id", 0, "resource ID")
}
