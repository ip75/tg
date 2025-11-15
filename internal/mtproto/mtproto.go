package mtproto

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/rs/zerolog/log"
	"gitlab.com/bvgm/tg/internal/domain"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const defaultRateLimit = time.Second

type SesstionParams struct {
	TgAppID        int
	TgAppHash      string
	MtprotoGroupID int64
	AccessHash     int64
	TgBotToken     string
	Threads        int
	RateLimit      time.Duration
}

type SingleInstanceFile struct {
	MediaID          int
	SerializedObject string
}

func (sif *SingleInstanceFile) create() (tg.InputFileClass, error) {
	file, err := Unmarshal(sif.SerializedObject)
	if err != nil {
		return nil, fmt.Errorf("unmarshal single instance file: %w", err)
	}
	return file, nil
}

type PublishAudioFunc func(audio domain.Audio, tok *string) (string, error)

type MTProtoClient struct {
	client *telegram.Client
	sess   SesstionParams
	sCtx   context.Context
	logger *zap.Logger
	waiter *floodwait.Waiter
}

func (c *MTProtoClient) Client() *telegram.Client {
	return c.client
}

func New(ctx context.Context, p SesstionParams) (*MTProtoClient, error) {

	logger, err := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "console",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build()
	if err != nil {
		return nil, fmt.Errorf("create zap logger: %w", err)
	}

	waiter := floodwait.NewWaiter().
		WithMaxWait(time.Hour).
		WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
			// Notifying about flood wait.
			fmt.Println("Got FLOOD_WAIT. Will retry after ", wait.Duration.String())
			log.Warn().Dur("wait", wait.Duration).Msg("Flood wait")
		})

	if p.RateLimit == 0 {
		p.RateLimit = defaultRateLimit
	}
	client := telegram.NewClient(
		p.TgAppID,
		p.TgAppHash,
		telegram.Options{
			OnDead: func() {
				log.Error().Msg("telegram client dead")
			},
			RetryInterval:   time.Second * 7,
			MaxRetries:      5,
			DialTimeout:     time.Second * 10,
			ExchangeTimeout: time.Second * 10,
			SessionStorage:  &SessionCache{},
			Logger:          logger,
			Middlewares: []telegram.Middleware{
				// Setting up general rate limits to less likely get flood wait errors.
				ratelimit.New(rate.Every(p.RateLimit), 5),
				// Handler of FLOOD_WAIT that will automatically retry request.
				waiter,
			},
		})

	// client.Pool(5)

	return &MTProtoClient{
		client: client,
		sess:   p,
		logger: logger,
		waiter: waiter,
	}, nil
}

func (c *MTProtoClient) Close() {

	if err := c.logger.Sync(); err != nil {
		log.Error().Err(err).Msg("close zap logger")
	}
}

// ctx - application context.
func (c *MTProtoClient) StartSession(ctx context.Context, queueProcessor func(PublishAudioFunc) error) error {
	log.Info().Msg("creating mtproto session")

	err := c.waiter.Run(ctx, func(ctx context.Context) error {
		return c.client.Run(ctx, func(ctx context.Context) error {
			c.sCtx = ctx
			// Checking auth status.
			status, err := c.client.Auth().Status(ctx)
			if err != nil {
				return fmt.Errorf("get auth status: %w", err)
			}

			// Can be already authenticated if we have valid session in
			// session storage.
			if !status.Authorized {
				// Otherwise, perform bot authentication.
				if _, err := c.client.Auth().Bot(ctx, c.sess.TgBotToken); err != nil {
					return fmt.Errorf("bot auth: %w", err)
				}
			}
			status, err = c.client.Auth().Status(ctx)
			if err != nil {
				return fmt.Errorf("get auth status: %w", err)
			}
			log.Info().Interface("status", status).Msg("Authenticated")

			return queueProcessor(c.PublishAudio)
		})
	})

	if err != nil {
		return fmt.Errorf("stop session: %w", err)
	}

	log.Info().Msg("session closed")
	return nil
}

// ctx - Must be sesstion context
// https://core.telegram.org/api/forum
// https://core.telegram.org/constructor/inputReplyToMessage - to send to topic
func (c *MTProtoClient) PublishAudio(audio domain.Audio, tok *string) (string, error) {
	log.Info().Bool("single_instance", tok != nil).Msg("sending media to group")

	var (
		f   tg.InputFileClass
		err error
	)

	if tok != nil {
		sif := &SingleInstanceFile{SerializedObject: *tok, MediaID: audio.MediaID}
		f, err = sif.create()
		if err != nil {
			log.Warn().Err(err).Msg("single instance file object exist, but can't create. Try to uplad again.")
		}
	}

	if f == nil {
		// Helper for uploading. Automatically uses big file upload when needed.
		f, err = uploader.
			NewUploader(c.client.API()).
			WithThreads(c.sess.Threads).
			FromPath(c.sCtx, audio.Path)
		if err != nil {
			return "", fmt.Errorf("upload %q: %w", audio.Path, err)
		}
	}

	siID, err := Marshal(f)
	if err != nil {
		return "", fmt.Errorf("serialize uploaded audio for single instance: %w", err)
	}

	// Helper for sending messages.
	sender := message.NewSender(c.client.API())

	r := sender.To(&tg.InputPeerChannel{
		ChannelID:  c.sess.MtprotoGroupID,
		AccessHash: c.sess.AccessHash,
	})

	caption := []message.StyledTextOption{styling.Plain(audio.Title), styling.Plain("\n")}
	caption = append(caption, styling.Hashtag(audio.HashTag()))

	// https://github.com/gotd/td/pull/1597 - message.Audio does not allow to set filename attribute
	if _, err := r.Reply(audio.MessageThreadID).
		Media(c.sCtx, message.UploadedDocument(f,
			caption...,
		).MIME(message.DefaultAudioMIME).
			Filename(filepath.Base(audio.Path)).
			Attributes(&tg.DocumentAttributeAudio{
				Title:     audio.Title,
				Performer: audio.Performer,
				Duration: func() int {
					if audio.Duration == nil {
						return 0
					}
					return int(audio.Duration.Seconds())
				}(),
			})); err != nil {
		return "", fmt.Errorf("send media: %w", err)
	}

	return siID, nil
}

// example to get channel with ID:
// channel, err := getChannel(ctx, client.API(), cfg.GetInt64("telegram.mtproto_group_id"))
//
//	if err != nil {
//		return fmt.Errorf("get channel: %w", err)
//	}
func (c *MTProtoClient) GetChannelInfo(ctx context.Context, channelID int64) (*tg.Channel, error) {
	inputChannel := &tg.InputChannel{
		ChannelID:  channelID,
		AccessHash: 0,
	}
	channels, err := c.client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})

	if err != nil {
		return nil, fmt.Errorf("fetch channel: %w", err)
	}

	if len(channels.GetChats()) == 0 {
		return nil, fmt.Errorf("no channels found")
	}

	channel := channels.GetChats()[0].(*tg.Channel)
	return channel, nil
}
