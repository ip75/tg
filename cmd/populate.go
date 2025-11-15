/*
Copyright © 2025 <admin@goswami.ru>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/bvgm/tg/internal/database"
)

var (
	since               time.Time
	recent              time.Duration
	sinceSet, recentSet bool
	tagID               int
)

// populateCmd represents the populate command
var populateCmd = &cobra.Command{
	Use:   "populate",
	Short: "Populate audio to database queue for uploading to telegram DC.",
	Long: `Populate audio to database queue since specified date.
Telegram bot processor will upload media from queue to telegram DC.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config = viper.GetViper()

		if cmd.Flags().Changed("since") && cmd.Flags().Changed("recent") {
			log.Fatal().Msg("specify only one flag, --recent or --since")
		}

		if !cmd.Flags().Changed("since") && !cmd.Flags().Changed("recent") {
			log.Fatal().Msg("specify flag, --recent or --since")
		}

		d, err := database.New(config.GetString("database.dsn"))
		if err != nil {
			log.Fatal().Err(err).Msg("connect to database")
		}
		defer d.Close()

		if cmd.Flags().Changed("since") {
			fmt.Printf("populate audio since: %s\n", since)
			if err := d.PopulateMedia(ctx, since, tagID); err != nil {
				log.Fatal().Err(err).Msg("populate audio to queue")
			}
		}

		if cmd.Flags().Changed("recent") {
			fmt.Printf("populate recent audio: %s\n", recent)
			if err := d.PopulateMedia(ctx, time.Now().Add(-recent), tagID); err != nil {
				log.Fatal().Err(err).Msg("populate audio to queue")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(populateCmd)
	populateCmd.Flags().TimeVarP(&since, "since", "s", time.Now(), []string{time.DateOnly, time.RFC3339}, "Time since populate audio to queue.")
	populateCmd.Flags().DurationVarP(&recent, "recent", "r", 0, "Specify duration populate audio to queue.")
	populateCmd.Flags().IntVarP(&tagID, "tagid", "t", 0, "Tag ID to populate audio to queue.")

}
