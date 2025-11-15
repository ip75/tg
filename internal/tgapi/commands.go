package tgapi

type createForumTopic struct {
	ChatID  int     `json:"chat_id"`
	Name    string  `json:"name"`
	EmojiID *string `json:"icon_custom_emoji_id,omitempty"`
}

type forumTopic struct {
	MessageThreadID int    `json:"message_thread_id"`    //	Unique identifier of the forum topic
	Name            string `json:"name"`                 //	Name of the topic
	IconColor       int    `json:"icon_color"`           //	Color of the topic icon in RGB format
	IconEmojiID     string `json:"icon_custom_emoji_id"` //	Optional. Unique identifier of the custom emoji shown as the topic icon
}

type createForumTopicResponse struct {
	Ok    bool       `json:"ok"`
	Topic forumTopic `json:"result"`
}

type createForumTopicResponseResult struct {
	Ok      bool   `json:"ok"`
	ErrCode int    `json:"error_code"`
	Message string `json:"description"`
}
