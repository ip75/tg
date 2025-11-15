package cmd

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/bvgm/tg/internal/database"
	"gitlab.com/bvgm/tg/internal/tgapi"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update topics in telegram",
	Long:  `Update topics by tag to sync with database. To publish tg_topic.created must be null`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		cfg := viper.GetViper()

		d, err := database.New(cfg.GetString("database.dsn"))
		if err != nil {
			log.Error().Err(err).Msg("connect to database")
			return
		}
		defer d.Close()

		topList, err := d.ListAllTopics(ctx)
		if err != nil {
			log.Error().Err(err).Msg("list all topics")
			return
		}

		cntUpdated := 0
		tg := tgapi.New(topList)
		for _, topic := range topList {
			if topic.CreatedAt != nil {
				continue
			}

			resp, err := tg.CreateGroupTopic(topic)
			if err != nil {
				log.Error().Err(err).Msg("publish topic")
				return
			}

			if err := d.MakeTopicPublished(ctx, resp.MessageThreadID, topic.ID); err != nil {
				log.Error().Err(err).Msg("make topic published")
				return
			}
			cntUpdated++
			log.Info().
				Int("topicID", topic.MessageThreadID).
				Str("name", topic.Name).
				Msg("add topic")
		}
		log.Info().
			Int("count", cntUpdated).
			Msg("topics published")
	},
}

func init() {
	topicsCmd.AddCommand(updateCmd)
}
