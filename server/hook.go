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
	"github.com/mattermost/ops-tool/slashcommand"
)

func (s *Server) hookHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	log := log.FromContext(ctx)

	slashCommand, err := model.ParseMattermostSlashCommand(r)
	if err != nil {
		log.WithError(err).Error("unable to parse mattermost slash command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Debugf("slashCommand received: %#v", *slashCommand)

	rootCommand, cmdText, args, err := ParseCommand(slashCommand.Command + " " + slashCommand.Text)
	if err != nil {
		log.WithError(err).Warn("unable to parse user command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debugf("Root Command: %s", rootCommand)
	log.Debugf("Command: " + cmdText)
	log.Debugf("Args: %#v", args)

	cmd := s.findCommand(rootCommand)
	if cmd == nil {
		log.WithError(err).Error("Command not found, sending full help")
		help := s.fullHelp()
		WriteResponse(w, help, mmmodel.CommandResponseTypeEphemeral)
		return
	}

	// make sure the token is valid
	if !strings.EqualFold(cmd.Token, slashCommand.Token) {
		log.Error("Invalid token - possibly a crafted command")
		WriteErrorResponse(w, NewError("Invalid command token! Please check slash command token", err))
		return
	}

	response, err := cmd.Execute(ctx, slashCommand, cmdText, args)
	if err != nil {
		if err == slashcommand.ErrCommandNotFound {
			log.Debug("Command not found, sending help")

			help := "**Command not found. Available commands:**\n\n" + s.helpForCommand(*cmd)
			WriteResponse(w, help, mmmodel.CommandResponseTypeEphemeral)
			return
		}

		log.WithError(err).Error("Error occurred while executing command")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch response.Type {
	case model.CommandResponseTypeInChannel, model.CommandResponseTypeEphemeral:
		msg := response.Message
		WriteEnrichedResponse(w, msg.Title, msg.Body, msg.Color, msg.ResponseType)
		return
	case model.CommandResponseTypeDialog:
		request, err := s.DialogStore.Create(
			slashCommand,
			rootCommand,
			cmdText,
			args,
			response.Dialog,
		)
		request.URL = s.Config.BaseURL + "/dialog"
		log.Debug("dialog response to: " + request.URL)
		if err != nil {
			log.WithError(err).Error("unable to create dialog")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		s.SendDialogRequest(ctx, cmd.DialogURL, request)
		return
	}
	// TODO: error for unknown response type?
}

func (s *Server) SendDialogRequest(ctx context.Context, url string, request *mmmodel.OpenDialogRequest) {
	b, err := json.Marshal(request)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("unable to marshal dialog request")
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("unable to create dialog request")
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("unable to send dialog request")
		return
	}

	if response.StatusCode != 200 {
		log.FromContext(ctx).With("status", response.Status).Error("unable to send dialog request")
		return
	}
}

func WriteEnrichedResponse(w http.ResponseWriter, title, resp, color, style string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(GenerateEnrichedSlashResponse(title, resp, color, style))
}

func GenerateEnrichedSlashResponse(title, text, color, respType string) []byte {
	msgAttachment := &mmmodel.SlackAttachment{
		Fallback: text,
		Color:    color,
		Text:     text,
		Title:    title,
	}

	response := mmmodel.CommandResponse{
		ResponseType: respType,
		Text:         "",
		Attachments:  []*mmmodel.SlackAttachment{msgAttachment},
		GotoLocation: "",
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.Default().Error("Unable to marshal response")
		return nil
	}

	return b
}

func GenerateStandardSlashResponse(text string, respType string) string {
	response := mmmodel.CommandResponse{
		ResponseType: respType,
		Text:         text,
		GotoLocation: "",
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.Default().Error("Unable to marshal response")
		return ""
	}
	return string(b)
}

func WriteResponse(w http.ResponseWriter, resp string, style string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(GenerateStandardSlashResponse(resp, style)))
}

func WriteErrorResponse(w http.ResponseWriter, err *AppError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(GenerateStandardSlashResponse(err.Error(), mmmodel.CommandResponseTypeEphemeral)))
}
