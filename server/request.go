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

	request, error := http.NewRequest("POST", hook, bytes.NewBuffer(data))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		LogError("[%s]Error occured while sending data to %s. %v", title, hook, error)
		return
	}

	if response.StatusCode != 200 {
		LogError("[%s]Got %d while invoking %s.", title, response.StatusCode, hook)
	}
}
