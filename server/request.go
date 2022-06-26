package server

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
)

func GenerateIncomingWebhookRequest(title, text, color string) []byte {
	msgAttachment := &model.SlackAttachment{
		Fallback: text,
		Color:    color,
		Text:     text,
		Title:    title,
	}

	request := model.IncomingWebhookRequest{
		Attachments: []*model.SlackAttachment{msgAttachment},
	}

	b, err := json.Marshal(request)
	if err != nil {
		LogError("Unable to marshal request")
		return nil
	}

	return b
}

func SendViaIncomingHook(hook, title, text, color string) {
	data := GenerateIncomingWebhookRequest(title, text, color)

	request, err := http.NewRequest("POST", hook, bytes.NewBuffer(data))
	if err != nil {
		LogError("[%s]Error occurred creating request to %s. %v", title, hook, err)
		return
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		LogError("[%s]Error occurred while sending data to %s. %v", title, hook, err)
		return
	}

	if response.StatusCode != 200 {
		LogError("[%s]Got %d while invoking %s.", title, response.StatusCode, hook)
	}
}
