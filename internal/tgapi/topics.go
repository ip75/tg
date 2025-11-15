package tgapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
	"gitlab.com/bvgm/tg/internal/domain"
)

type TelegramPublisher struct {
	list   []domain.Topic
	config *viper.Viper
}

func New(top []domain.Topic) TelegramPublisher {

	return TelegramPublisher{
		config: viper.GetViper(),
		list:   top,
	}
}

func (t *TelegramPublisher) CreateGroupTopic(topic domain.Topic) (*forumTopic, error) {

	u := url.URL{
		Scheme: "https",
		Host:   "api.telegram.org",
		Path:   fmt.Sprintf("/bot%v/createForumTopic", t.config.GetString("telegram.bot_token")),
	}

	jb, err := json.Marshal(createForumTopic{
		ChatID:  t.config.GetInt("telegram.group_id"),
		Name:    topic.Name,
		EmojiID: topic.IconCustomEmojiID,
	})

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(jb))
	if err != nil {
		return nil, fmt.Errorf("create http request to publish topic: %w", err)
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create forum topic: %w", err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read http response: %w", err)
	}

	var response createForumTopicResponse

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("JSON parse created topic: %w", err)
	}

	if !response.Ok {
		var status createForumTopicResponseResult
		if err := json.Unmarshal(body, &status); err != nil {
			return nil, fmt.Errorf("JSON parse status: %w", err)
		}
		return nil, fmt.Errorf("create forum topic: %s (%d)", status.Message, status.ErrCode)
	}
	return &response.Topic, nil
}
