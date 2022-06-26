package server

import (
	"encoding/json"
	"io/ioutil"
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

type DialogSubmission struct {
	Type       string `json:"type"`
	CallbackID string `json:"callback_id"`
	State      string `json:"state"`
	UserID     string `json:"user_id"`
	ChannelID  string `json:"channel_id"`
	TeamID     string `json:"team_id"`
	//nolint:misspell // cancelled is misspelled but it is sent from mattermost-server.
	Canceled   bool              `json:"cancelled"`
	Submission map[string]string `json:"submission"`
}

type DialogSession struct {
	CallbackID   string
	MMHookURL    string
	SlashCommand *MMSlashCommand
	OpsCommand   *OpsCommand
}

var DialogSessions map[string]*DialogSession = make(map[string]*DialogSession, 0)

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

func ParseDialogSubmission(r *http.Request) (*DialogSubmission, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		LogError("Error reading body: %v", err)
		return nil, err
	}
	dialogSubmission := &DialogSubmission{}

	err = json.Unmarshal(body, dialogSubmission)
	if err != nil {
		LogError("Unable to unmarshal dialog submission %v", err)
		return nil, err
	}
	LogInfo(string(body))
	return dialogSubmission, nil
}
