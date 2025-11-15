package domain

import "time"

const DefaultTopicEmojiID = "5377317729109811382"

type Topic struct {
	ID                uint64
	MessageThreadID   int
	Name              string
	TagID             int
	Tag               string
	IconCustomEmojiID *string
	CreatedAt         *time.Time
}
