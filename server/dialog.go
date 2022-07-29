package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	mmmodel "github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/ops-tool/log"
	"github.com/mattermost/ops-tool/model"
	"github.com/pkg/errors"
)

func (s *Server) dialogHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log := log.FromContext(ctx)

	dialogSubmission, err := model.ParseDialogSubmission(r)
	if err != nil {
		log.WithError(err).Error("unable to parse dialog submission")
		WriteErrorResponse(w, NewError("Unable to parse dialog submission", err))
		return
	}
	log.Debugf("Received valid dialog submission for session %s, canceled: %t", dialogSubmission.CallbackID, dialogSubmission.Canceled)

	dialogSession, err := s.DialogStore.Get(dialogSubmission.CallbackID)
	if err != nil {
		log.WithError(err).Error("unable to get dialog session")
		WriteResponse(w, "Session not found! Trigger slash command again!", mmmodel.CommandResponseTypeEphemeral)
		return
	}

	s.DialogStore.Delete(dialogSession.CallbackID)
	log.Debugf("Session %s is terminated. %d session is active.", dialogSession.CallbackID, s.DialogStore.Count())
	if dialogSubmission.Canceled {
		return
	}

	for _, cmd := range s.commands {
		if strings.EqualFold(dialogSession.RootCommand, cmd.Command) {
			response, err := cmd.ExecuteDialog(ctx, dialogSubmission, dialogSession.Command, dialogSession.CommandArgs)
			if err != nil {
				log.Println("execute dialog: " + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			switch response.Type {
			case model.CommandResponseTypeInChannel, model.CommandResponseTypeEphemeral:
				msg := response.Message
				channel := dialogSession.ChannelName
				if response.Type == model.CommandResponseTypeEphemeral {
					channel = "@" + dialogSession.Username
				}

				SendViaIncomingHook(cmd.DialogResponseURL, channel, msg.Title, msg.Body, msg.Color)
				return
			default:
				// err
			}
			break
		}
	}
}

func SendViaIncomingHook(hook, channelName, title, text, color string) error {
	data, err := GenerateIncomingWebhookRequest(channelName, title, text, color)
	if err != nil {
		return errors.Wrap(err, "unable to generate incoming webhook request")
	}

	request, err := http.NewRequest("POST", hook, bytes.NewBuffer(data))
	if err != nil {
		return errors.Wrap(err, "unable to create request")
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return errors.Wrap(err, "unable to send request")
	}

	if response.StatusCode != 200 {
		return errors.New("unexpected response status code: " + response.Status)
	}
	return nil
}

func GenerateIncomingWebhookRequest(channelName, title, text, color string) ([]byte, error) {
	msgAttachment := &mmmodel.SlackAttachment{
		Fallback: text,
		Color:    color,
		Text:     text,
		Title:    title,
	}

	request := mmmodel.IncomingWebhookRequest{
		ChannelName: channelName,
		Attachments: []*mmmodel.SlackAttachment{msgAttachment},
	}

	b, err := json.Marshal(request)

	return b, errors.Wrap(err, "unable to marshal request")
}
