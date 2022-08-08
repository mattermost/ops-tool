package model

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type Dialog struct {
	Title       string          `yaml:"title"`
	URL         string          `yaml:"url"`
	CallbackURL string          `yaml:"callbackUrl"`
	MMHookURL   string          `yaml:"hook"`
	Text        string          `yaml:"introduction_text"`
	Elements    []DialogElement `yaml:"elements"`
}

type DialogElement struct {
	DisplayName string          `yaml:"display_name"`
	Name        string          `yaml:"name"`
	Type        string          `yaml:"type"`
	SubType     string          `yaml:"subtype"`
	Default     string          `yaml:"default"`
	Placeholder string          `yaml:"placeholder"`
	HelpText    string          `yaml:"help_text"`
	Optional    bool            `yaml:"optional"`
	MinLength   int             `yaml:"min_length"`
	MaxLength   int             `yaml:"max_length"`
	Options     []*DialogOption `yaml:"options"`
}

type DialogOption struct {
	Text  string `yaml:"text"`
	Value string `yaml:"value"`
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

func ParseDialogSubmission(r *http.Request) (*DialogSubmission, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read request body")
	}
	dialogSubmission := &DialogSubmission{}

	err = json.Unmarshal(body, dialogSubmission)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal request body")
	}

	return dialogSubmission, nil
}
