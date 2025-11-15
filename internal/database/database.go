package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/bvgm/tg/internal/database/gen"
	"gitlab.com/bvgm/tg/internal/domain"
)

var ErrEmptyQueue = errors.New("media to publish not found")
var ErrNoSingleInstance = errors.New("no single instance data for this media")

type Tgdb struct {
	pool    *pgxpool.Pool
	queries *gen.Queries
}

func New(urn string) (Tgdb, error) {
	d := Tgdb{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := d.connect(ctx, urn); err != nil {
		return Tgdb{}, fmt.Errorf("New connect to database: %w", err)
	}

	if err := d.Ping(ctx); err != nil {
		return Tgdb{}, fmt.Errorf("ping to database: %s", err)
	}

	return d, nil
}

func (d *Tgdb) connect(ctx context.Context, urn string) error {

	pool, err := pgxpool.New(ctx, urn)
	if err != nil {
		return fmt.Errorf("create connection pool: %w", err)
	}

	d.pool = pool
	d.queries = gen.New(pool)

	return nil
}

func (d *Tgdb) Close() {
	d.pool.Close()
}

func (d *Tgdb) Ping(ctx context.Context) error {
	if err := d.pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	return nil
}

func (d *Tgdb) SetRecentUploadTime(ctx context.Context, slug string, rut time.Time) error {
	if err := d.queries.SetRecentUploadTime(ctx, gen.SetRecentUploadTimeParams{
		Slug:             slug,
		RecentUploadTime: rut,
	}); err != nil {
		return fmt.Errorf("set recent update time: %w", err)
	}

	return nil
}
func (d *Tgdb) GetRecentUploadTime(ctx context.Context, slug string) (time.Time, error) {
	rut, err := d.queries.GetRecentUploadTime(ctx, slug)
	if err != nil {
		return time.Time{}, fmt.Errorf("get recent update time: %w", err)
	}
	return rut, nil
}

func (d *Tgdb) ListAllTopics(ctx context.Context) ([]domain.Topic, error) {
	topics, err := d.queries.ListAllTopics(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all topics: %w", err)
	}

	return genTopics(topics), nil
}

func (d *Tgdb) ListMediaQueue(ctx context.Context, limit int32, cursor uint64) ([]domain.Audio, uint64, error) {
	audioToPublish, err := d.queries.ListMediaQueue(ctx, gen.ListMediaQueueParams{
		ID:    cursor, // tg_queue.id is a cursor
		Limit: limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list queue to publish: %w", err)
	}
	if len(audioToPublish) == 0 {
		return nil, 0, ErrEmptyQueue
	}

	res := make([]domain.Audio, 0, len(audioToPublish))
	for _, a := range audioToPublish {

		res = append(res, domain.Audio{
			MediaID: a.MediaID,
			Title:   a.Title,
			Teaser:  a.Teaser,
			Path: func() string {
				if a.FileUrl != nil {
					return *a.FileUrl
				}
				return ""
			}(),
			MessageThreadID: a.MessageThreadID,
			TagID:           a.TagID,
			Tag:             a.Tag,
			OccurrenceDate:  a.OccurrenceDate,
			IssueDate:       a.IssueDate,
			Duration:        a.Duration,
			Size:            a.Size,
		})
	}

	return res, audioToPublish[len(audioToPublish)-1].Cursor, nil
}

func (d *Tgdb) AddAudioToFailedQueue(ctx context.Context, a domain.Audio, err error) error {
	if err := d.queries.AddMediaToFailedQueue(ctx, gen.AddMediaToFailedQueueParams{
		Error:           err.Error(),
		MediaID:         a.MediaID,
		MessageThreadID: a.MessageThreadID,
		TagID:           a.TagID,
	}); err != nil {
		return fmt.Errorf("add audio to failed queue: %w", err)
	}

	return nil
}

func (d *Tgdb) RemoveFromQueue(ctx context.Context, mediaID, tagID int) error {
	if err := d.queries.RemoveMediaQueue(ctx, gen.RemoveMediaQueueParams{
		MediaID: mediaID,
		TagID:   tagID,
	}); err != nil {
		return fmt.Errorf("remove id from queue: %w", err)
	}
	return nil
}

func (d *Tgdb) MakeTopicPublished(ctx context.Context, MessageThreadID int, ID uint64) error {
	if err := d.queries.MakeTopicPublished(ctx, gen.MakeTopicPublishedParams{
		MessageThreadID: MessageThreadID,
		ID:              ID,
	}); err != nil {
		return fmt.Errorf("mark topic as published: %w", err)
	}
	return nil
}

// INFO: For future use when will be many chatbots on one server
func (d *Tgdb) GetConfig(ctx context.Context, slug string) (*domain.Config, error) {
	cfg, err := d.queries.GetConfig(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("mark topic as published: %w", err)
	}

	settings := domain.BotSettings{UploadThreads: 2}
	if cfg.Settings != nil {
		err = json.Unmarshal(cfg.Settings, &settings)
		if err != nil {
			return nil, fmt.Errorf("unmarshal settings: %w", err)
		}
	}

	return &domain.Config{
		Slug:             cfg.Slug,
		RecentUploadTime: cfg.RecentUploadTime,
		Settings:         settings,
	}, nil
}

// link media to telegragm audio. This is important to make single instance storage for audio files.
// TgAudioID can be used for sending audio files without uploading media each time
func (d *Tgdb) LinkMediaToTelegram(ctx context.Context, MediaID int, TgAudioID string) error {
	if err := d.queries.LinkMediaToTelegram(ctx, gen.LinkMediaToTelegramParams{
		MediaID: MediaID,
		Value:   TgAudioID,
	}); err != nil {
		return fmt.Errorf("link media to telegram audio file: %w", err)
	}
	return nil
}

func (d *Tgdb) PopulateMedia(ctx context.Context, t time.Time, tagID int) error {
	if tagID == 0 {
		if err := d.queries.PopulateMedia(ctx, t); err != nil {
			return fmt.Errorf("populate database: %w", err)
		}
		return nil
	}

	if err := d.queries.PopulateMediaWithTagID(ctx, gen.PopulateMediaWithTagIDParams{
		OccurrenceDate: t,
		ID:             tagID,
	}); err != nil {
		return fmt.Errorf("populate database: %w", err)
	}

	return nil
}

func (d *Tgdb) GetSingleInstanceAudio(ctx context.Context, mediaID int) (*string, error) {
	r, err := d.queries.GetMediaDataTelegram(ctx, mediaID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNoSingleInstance
		}
		return nil, fmt.Errorf("get single instance audio: %w", err)
	}
	return &r.Value, nil
}

func genTopics(topics []gen.ListAllTopicsRow) []domain.Topic {
	mTop := make([]domain.Topic, 0, len(topics))
	for _, topic := range topics {
		mTop = append(mTop, genTopic(topic))
	}

	return mTop
}

func genTopic(topic gen.ListAllTopicsRow) domain.Topic {
	return domain.Topic{
		ID:                topic.ID,
		MessageThreadID:   topic.MessageThreadID,
		TagID:             topic.TagID,
		Name:              topic.Name,
		IconCustomEmojiID: topic.IconCustomEmojiID,
		CreatedAt:         topic.Created,
		Tag:               topic.Tag,
	}
}
