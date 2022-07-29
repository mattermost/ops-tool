package model

import (
	"net/http"

	"github.com/gorilla/schema"
)

type MMSlashCommand struct {
	ChannelID   string `schema:"channel_id"`
	ChannelName string `schema:"channel_name"`
	Command     string `schema:"command"`
	TeamName    string `schema:"team_domain"`
	TeamID      string `schema:"team_id"`
	Text        string `schema:"text"`
	Token       string `schema:"token"`
	UserID      string `schema:"user_id"`
	Username    string `schema:"user_name"`
	ResponseURL string `schema:"response_url"`
	TriggerID   string `schema:"trigger_id"`
}

func ParseSlashCommand(r *http.Request) (*MMSlashCommand, error) {
	err := r.ParseForm()
	if err != nil {
		return nil, err
	}
	inCommand := &MMSlashCommand{}
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	err = decoder.Decode(inCommand, r.Form)
	if err != nil {
		return nil, err
	}

	return inCommand, nil
}
