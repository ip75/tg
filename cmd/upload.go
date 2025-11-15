/*
Copyright © 2025 <admin@goswami.ru>
*/
package cmd

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var audioPath string

// uploadCmd represents the upload command
var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload media to telegram",
	Long:  `Upload audio to telegram without 50M limit. Limit 2G. app_id and app_hash should be specified in config file to work.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := viper.GetViper()
		tgAppID := cfg.GetInt("telegram.app_id")
		tgAppHash := cfg.GetString("telegram.app_hash")
		mtprotoGroupID := cfg.GetInt64("telegram.mtproto_group_id")
		AccessHash := cfg.GetInt64("telegram.access_hash")
		jobs := cfg.GetInt("server.jobs")
		tgBotToken := cfg.GetString("telegram.bot_token")

		if cfg.InConfig("telegram.app_id") == false || cfg.InConfig("telegram.app_hash") == false || cfg.InConfig("telegram.bot_token") == false {
			log.Error().Msg("app_id, app_hash and bot_token required in config file")
			return
		}

		if len(audioPath) == 0 {
			log.Error().Msg("empty audio path")
			return
		}

		client := telegram.NewClient(
			tgAppID,
			tgAppHash,
			telegram.Options{
				OnDead: func() {
					log.Error().Msg("telegram client dead")
				},
				RetryInterval:  time.Second,
				MaxRetries:     5,
				SessionStorage: &SessionCache{},
			})

		ctx, cancel := context.WithCancelCause(context.Background())

		err := client.Run(ctx, func(ctx context.Context) error {
			// Checking auth status.
			status, err := client.Auth().Status(ctx)
			if err != nil {
				return err
			}
			defer func() {
				if err != nil {
					cancel(err)
				}
			}()

			// Can be already authenticated if we have valid session in
			// session storage.
			if !status.Authorized {
				// Otherwise, perform bot authentication.
				if _, err := client.Auth().Bot(ctx, tgBotToken); err != nil {
					return err
				}
			}
			status, err = client.Auth().Status(ctx)
			if err != nil {
				return err
			}
			log.Info().Interface("status", status).Msg("Authenticated")

			// Helper for uploading. Automatically uses big file upload when needed.
			f, err := uploader.
				NewUploader(client.API()).
				WithThreads(jobs).
				FromPath(ctx, audioPath)
			if err != nil {
				return fmt.Errorf("upload %q: %w", audioPath, err)
			}

			// Helper for sending messages.
			sender := message.NewSender(client.API())

			r := sender.To(&tg.InputPeerChannel{
				ChannelID:  mtprotoGroupID,
				AccessHash: AccessHash,
			})

			// if _, err := r.Reply(4).Media(ctx, message.Audio(f, styling.Hashtag("#ШП"), styling.Plain("\n"), styling.Plain(path.Base(audioPath))).
			// 	Performer("Бхакти Вигьяна Госвами").
			// 	Title(filepath.Base(audioPath))); err != nil {
			// 	return fmt.Errorf("send media: %w", err)
			// }

			if _, err := r.Reply(4).Media(ctx, message.UploadedDocument(f, styling.Hashtag("#ШП"), styling.Plain("\n"), styling.Plain(path.Base(audioPath))).
				MIME(message.DefaultAudioMIME).
				Attributes(&tg.DocumentAttributeAudio{
					Title:     filepath.Base(audioPath),
					Performer: "Бхакти Вигьяна Госвами",
				}).
				Filename(path.Base(audioPath))); err != nil {
				return fmt.Errorf("send media: %w", err)
			}

			return nil
		})

		if err != nil {
			log.Fatal().Err(err).Msg("client error")
		}
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVarP(&audioPath, "path", "p", "", "path to audio file")
}

type SessionCache struct {
	atomic.Value
}

func (c *SessionCache) LoadSession(ctx context.Context) ([]byte, error) {
	s := c.Value.Load()
	if s == nil {
		return nil, session.ErrNotFound
	}
	if s, ok := s.([]byte); ok {
		return s, nil
	}
	return nil, session.ErrNotFound
}

func (c *SessionCache) StoreSession(ctx context.Context, data []byte) error {
	c.Value.Store(data)
	return nil
}
