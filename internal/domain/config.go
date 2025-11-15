package domain

import "time"

const DefaultConfigSlug = "goswami.ru"

type BotSettings struct {
	BotToken       string `json:"bot_token"`        // Telegram bot token
	GroupID        int    `json:"group_id"`         // Chat ID of telegram bot
	MtprotoGroupID int    `json:"mtproto_group_id"` // ChatID without trailing -100
	AccessHash     int    `json:"access_hash"`      // Group access hash. See GetChannelInfo
	MediaPath      string `json:"audio"`            // Path to directory with audio
	AssetsPath     string `json:"assets"`           // Path to directory with covers
	AppID          int    `json:"app_id"`           // telegram application ID. https://my.telegram.org/apps
	AppHash        string `json:"app_hash"`         // telegram application hash. https://my.telegram.org/apps
	UploadThreads  int    `json:"upload_threads"`   // number of threads to upload audio files
	Performer      string `json:"performer"`        // performer of audio
}

type Config struct {
	ID               int
	Slug             string    // unique name of config
	RecentUploadTime time.Time // Time of uploading recent audio
	Settings         BotSettings
}
