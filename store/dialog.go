package store

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	mmmodel "github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/ops-tool/model"
)

type DialogStore interface {
	Create(slashCommand *model.MMSlashCommand, rootCommand, command string, commandsArgs map[string]string, dialog model.Dialog) (*mmmodel.OpenDialogRequest, error)
	Get(callbackID string) (*DialogSession, error)
	Delete(callbackID string) error
	Count() int
}

type DialogSession struct {
	CallbackID  string
	CreatedAt   time.Time
	RootCommand string
	Command     string
	CommandArgs map[string]string
	ChannelName string
	Username    string
}

func NewInMemoryDialogStore() DialogStore {
	return &InMemoryDialogStore{
		dialogs: make(map[string]*DialogSession),
	}
}

type InMemoryDialogStore struct {
	dialogs map[string]*DialogSession
}

func (s *InMemoryDialogStore) Create(
	slashcommand *model.MMSlashCommand,
	rootCommand, command string,
	commandArgs map[string]string,
	dialog model.Dialog,
) (*mmmodel.OpenDialogRequest, error) {
	callbackID := uuid.NewString()
	elements := make([]mmmodel.DialogElement, 0)
	for _, opsElem := range dialog.Elements {
		options := make([]*mmmodel.PostActionOptions, 0)
		for _, option := range opsElem.Options {
			options = append(options, &mmmodel.PostActionOptions{
				Text:  option.Text,
				Value: option.Value,
			})
		}

		elements = append(elements, mmmodel.DialogElement{
			Name:        opsElem.Name,
			DisplayName: opsElem.DisplayName,
			Type:        opsElem.Type,
			SubType:     opsElem.SubType,
			Default:     opsElem.Default,
			Optional:    opsElem.Optional,
			HelpText:    opsElem.HelpText,
			Placeholder: opsElem.Placeholder,
			MinLength:   opsElem.MinLength,
			MaxLength:   opsElem.MaxLength,
			Options:     options,
		})
	}

	request := &mmmodel.OpenDialogRequest{
		TriggerId: slashcommand.TriggerID,
		Dialog: mmmodel.Dialog{
			CallbackId:       callbackID,
			Title:            dialog.Title,
			IntroductionText: dialog.Text,
			Elements:         elements,
			NotifyOnCancel:   true, // We need to remove dialog session from map to save memory.
		},
	}

	s.dialogs[callbackID] = &DialogSession{
		CallbackID:  callbackID,
		CreatedAt:   time.Now(),
		RootCommand: rootCommand,
		Command:     command,
		CommandArgs: commandArgs,
		ChannelName: slashcommand.ChannelName,
		Username:    slashcommand.Username,
	}

	go func() {
		// Everytime we create a dialog, clean older ones
		for callbackID, d := range s.dialogs {
			if time.Since(d.CreatedAt) > time.Minute*5 {
				s.Delete(callbackID)
			}
		}
	}()

	return request, nil
}

func (s *InMemoryDialogStore) Delete(callbackID string) error {
	delete(s.dialogs, callbackID)
	return nil
}

func (s *InMemoryDialogStore) Get(callbackID string) (*DialogSession, error) {
	session, found := s.dialogs[callbackID]
	if !found {
		return nil, fmt.Errorf("dialog session not found with callbackID %s", callbackID)
	}

	return session, nil
}

func (s *InMemoryDialogStore) Count() int {
	return len(s.dialogs)
}
