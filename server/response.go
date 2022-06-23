package server

import (
	"encoding/json"
	"net/http"

	"github.com/mattermost/mattermost-server/v6/model"
)

func GenerateStandardSlashResponse(text string, respType string) string {
	response := model.CommandResponse{
		ResponseType: respType,
		Text:         text,
		GotoLocation: "",
	}

	b, err := json.Marshal(response)
	if err != nil {
		LogError("Unable to marshal response")
		return ""
	}
	return string(b)
}

func GenerateEnrichedSlashResponse(title, text, color, respType string) []byte {
	msgAttachment := &model.SlackAttachment{
		Fallback: text,
		Color:    color,
		Text:     text,
		Title:    title,
	}

	response := model.CommandResponse{
		ResponseType: respType,
		Text:         "",
		Attachments:  []*model.SlackAttachment{msgAttachment},
		GotoLocation: "",
	}

	b, err := json.Marshal(response)
	if err != nil {
		LogError("Unable to marshal response")
		return nil
	}

	return b
}

func WriteErrorResponse(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(GenerateStandardSlashResponse(err.Error(), model.CommandResponseTypeEphemeral)))
}

func WriteResponse(w http.ResponseWriter, resp string, style string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(GenerateStandardSlashResponse(resp, style)))
}

func WriteEnrichedResponse(w http.ResponseWriter, title, resp, color, style string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(GenerateEnrichedSlashResponse(title, resp, color, style))
}
