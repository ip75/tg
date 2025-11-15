package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/bvgm/tg/internal/database"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list configured topics",
	Long:  `list configured topics`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := viper.GetViper()

		d, err := database.New(cfg.GetString("database.dsn"))
		if err != nil {
			fmt.Printf("connect to database: %s\n", err)
			return
		}
		defer d.Close()

		topics, err := d.ListAllTopics(ctx)
		if err != nil {
			fmt.Printf("load database: %s\n", err)
			return
		}

		t := table.NewWriter()
		t.SetStyle(table.StyleColoredDark)
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"ID", "Topic ID", "Topic Name", "Emoji ID", "CreatedAt", "TagID", "Tag Name"})
		for _, topic := range topics {
			t.AppendRow(table.Row{
				topic.ID, topic.MessageThreadID, topic.Name, topic.IconCustomEmojiID, topic.CreatedAt, topic.TagID, topic.Tag,
			})
		}
		t.Render()
	},
}

func init() {
	topicsCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
