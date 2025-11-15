/*
Copyright © 2025 <admin@goswami.ru>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/bvgm/tg/internal/database"
	"gitlab.com/bvgm/tg/internal/domain"
	"gitlab.com/bvgm/tg/internal/mtproto"
)

var (
	chunkSize      int
	updateInterval time.Duration
	d              database.Tgdb
	config         *viper.Viper
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start service to upload media to telegram from queue.",
	Long:  `Start service to upload media to telegram from queue.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		config = viper.GetViper()
		log.Info().Interface("settings", config.AllSettings()).Msg("config settings")
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
		defer close(sig)

		chunkSize = config.GetInt("server.chunkSize")
		if chunkSize <= 0 {
			log.Fatal().Msg("chunk size must be greater than 0")
		}

		var err error
		if d, err = database.New(config.GetString("database.dsn")); err != nil {
			log.Fatal().Err(err).Msg("connect to database")
		}
		defer d.Close()

		queue := make(chan domain.Audio, chunkSize)
		defer close(queue)

		updateInterval = config.GetDuration("server.update_interval") * time.Second

		ctx, cancel := context.WithCancelCause(ctx)
		defer cancel(nil)
		wg := sync.WaitGroup{}

		// queue updater
		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				cursor uint64
				data   []domain.Audio
				err    error
			)
			for {

				data, cursor, err = d.ListMediaQueue(ctx, int32(chunkSize), cursor)
				if err != nil {
					if err == database.ErrEmptyQueue {
						log.Debug().Dur("wait", updateInterval).Msg("queue is empty, wait for new data")
					} else {
						log.Error().Dur("wait", updateInterval).Err(err).Msg("fetch queue from database failed.")
					}
					select {
					case <-ctx.Done():
						log.Info().Err(context.Cause(ctx)).Msg("queue updater stopped")
						return
					case s := <-sig:
						cancel(fmt.Errorf("got a signal %s", s.String()))
						return
					case <-time.After(updateInterval): // wait until new request for data
						continue
					}
				}
				for _, a := range data {
					select {
					case <-ctx.Done():
						log.Info().Err(context.Cause(ctx)).Msg("queue updater stopped")
						return
					case queue <- a:
						log.Debug().Str("audio", a.Title).Msg("added to queue")
					case s := <-sig:
						cancel(fmt.Errorf("got a signal %s", s.String()))
						return
					}
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			ticker := time.NewTicker(time.Second * 2) // to avoid duplication, this period must be less than updateInterval
			defer ticker.Stop()

			for {
				select {
				case err := <-processor(ctx, queue):
					if err != nil {
						log.Debug().Err(err).Dur("restart after", updateInterval).Msg("session closed, restarting processor")
					} else {
						log.Debug().Dur("restart after", updateInterval).Msg("processor finished, restarting")
					}
				case <-ticker.C:
					log.Debug().Msg("let's check if there is something to do")
					continue
				case <-ctx.Done():
					log.Info().Err(context.Cause(ctx)).Msg("queue processor stopped")
					return
				case s := <-sig:
					cancel(fmt.Errorf("got a signal: %s", s.String()))
					return
				}
			}
		}()

		wg.Wait()
	},
}

func processor(ctx context.Context, queue chan domain.Audio) <-chan error {

	if len(queue) == 0 {
		log.Debug().Msg("Nothing to do: queue is empty")
		return nil
	}

	log.Info().Int("queue size", len(queue)).Msg("starting queue processor")

	errc := make(chan error, 1)
	defer close(errc)

	// jobs := viper.GetInt("server.jobs")
	audioBasePath := config.GetString("storage.audio")

	client, err := mtproto.New(ctx, mtproto.SesstionParams{
		TgAppID:        config.GetInt("telegram.app_id"),
		TgAppHash:      config.GetString("telegram.app_hash"),
		MtprotoGroupID: config.GetInt64("telegram.mtproto_group_id"),
		AccessHash:     config.GetInt64("telegram.access_hash"),
		TgBotToken:     config.GetString("telegram.bot_token"),
		Threads:        config.GetInt("telegram.upload_threads"), // number of threads that will upload media to telegram
		RateLimit:      config.GetDuration("telegram.rate_limit"),
	})
	if err != nil {
		log.Error().Err(err).Msg("create mtproto client")
		errc <- err
		return errc
	}
	defer client.Close()

	performer := "Бхакти Вигьяна Госвами"
	if config.InConfig("server.performer") {
		performer = config.GetString("server.performer")
	}

	err = client.StartSession(ctx, func(pub mtproto.PublishAudioFunc) error {
		for {
			select {
			case a := <-queue:
				log.Info().
					Str("tag", a.Tag).
					Str("title", a.Title).
					Str("path", a.Path).
					Int("queue size", len(queue)).
					Msg("sending media to telegram DC")

				if err = d.RemoveFromQueue(ctx, a.MediaID, a.TagID); err != nil {
					return fmt.Errorf("remove '%s' from queue: %w", a.Title, err)
				}

				sifToken, err := d.GetSingleInstanceAudio(ctx, a.MediaID)
				if err != nil {
					if err != database.ErrNoSingleInstance {
						return fmt.Errorf("get single instance audio: %w", err)
					}
				}

				msgID, err := pub(a.FullLocalPath(audioBasePath).SetPerformer(performer), sifToken)
				if err != nil {
					log.Error().Err(err).Str("title", a.Title).Msg("move to failed queue")

					if errdb := d.AddAudioToFailedQueue(ctx, a, err); errdb != nil {
						return fmt.Errorf("add '%s' to failed queue: %w", a.Title, errdb)
					}

					return fmt.Errorf("send media %s with tag: %s: %w", a.Path, a.Tag, err)
				}

				// save telegram message ID to use it for single instance
				err = d.LinkMediaToTelegram(ctx, a.MediaID, msgID)
				if err != nil {
					return fmt.Errorf("add telegram message ID '%s' to media data: %w", a.Title, err)
				}

				err = d.SetRecentUploadTime(ctx, domain.DefaultConfigSlug, time.Now())
				if err != nil {
					return fmt.Errorf("set recent upload time: %w", err)
				}
				log.Info().Str("tag", a.Tag).Str("title", a.Title).Str("file", filepath.Base(a.Path)).Msg("sent to telegram DC")

			case <-time.NewTimer(time.Minute * 15).C:
				if len(queue) == 0 {
					log.Info().Msg("Nothing to do: stop session")
					return nil
				}
			case <-ctx.Done():
				log.Info().Msg("processor canceled")
				return nil
			}
		}
	})

	if err != nil {
		log.Error().Err(err).Msg("queue processor: telegram bot session closed")
		errc <- err
	}
	return errc
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().Int("jobs", 2, "Number of upload goroutines to run (default 2).")
	if err := viper.BindPFlag("server.jobs", startCmd.Flags().Lookup("jobs")); err != nil {
		log.Fatal().Err(err).Msg("bind jobs flag")
	}
	startCmd.Flags().Int("chunkSize", 20, "Number of media to fetch from queue at once (default 20).")
	if err := viper.BindPFlag("server.chunkSize", startCmd.Flags().Lookup("chunkSize")); err != nil {
		log.Fatal().Err(err).Msg("bind chunkSize flag")
	}
	startCmd.Flags().DurationVarP(&updateInterval, "interval", "i", 15*time.Minute, "Interval between fetching media from queue.")
	if err := viper.BindPFlag("server.interval", startCmd.Flags().Lookup("interval")); err != nil {
		log.Fatal().Err(err).Msg("bind interval flag")
	}
}
