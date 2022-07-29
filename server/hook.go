package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	mmmodel "github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/ops-tool/model"
	"github.com/mattn/go-shellwords"
)

func (s *Server) hookHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	slashCommand, err := model.ParseSlashCommand(r)
	if err != nil {
		fmt.Println("parse slash command" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("slashCommand received: %#v", *slashCommand)

	res, err := shellwords.Parse(strings.TrimPrefix(slashCommand.Command, "/") + " " + slashCommand.Text)
	if err != nil {
		fmt.Println("parse slash command" + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmdText := ""
	args := map[string]string{}
	parsingArgs := false
	for _, x := range res[1:] {
		if !parsingArgs {
			if strings.HasPrefix(x, "--") {
				parsingArgs = true
			} else {
				cmdText += x + " "
			}
		}

		if parsingArgs {
			if !strings.HasPrefix(x, "--") {
				// ignore string that does not start with --
				continue
			}

			parts := strings.Split(x, "=")
			name := strings.TrimSpace(strings.TrimPrefix(parts[0], "--"))
			if len(parts) == 1 {
				args[name] = "true"
			} else {
				args[name] = strings.TrimSpace(parts[1])
			}
		}
	}
	cmdText = strings.TrimSpace(cmdText)

	log.Println("Command: " + cmdText)
	log.Printf("Args: %#v\n", args)
	for _, cmd := range s.commands {
		if strings.EqualFold(res[0], cmd.Command) {
			response, err := cmd.Execute(slashCommand, cmdText, args)
			if err != nil {
				log.Println("execute command: " + err.Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			switch response.Type {
			case model.CommandResponseTypeInChannel, model.CommandResponseTypeEphemeral:
				msg := response.Message
				WriteEnrichedResponse(w, msg.Title, msg.Body, msg.Color, msg.ResponseType)
				return
			case model.CommandResponseTypeDialog:
				// create a dialog
				request, err := s.DialogStore.Create(
					slashCommand,
					res[0],
					cmdText,
					args,
					response.Dialog,
				)
				request.URL = s.Config.BaseURL + "/dialog"
				log.Println("dialog response to: " + request.URL)
				if err != nil {
					log.Println("create dialog" + err.Error())
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				s.SendDialogRequest(cmd.DialogURL, request)
			default:
				// err
			}
			break
		}
	}
}

func (s *Server) SendDialogRequest(url string, request *mmmodel.OpenDialogRequest) {
	b, err := json.Marshal(request)
	if err != nil {
		log.Printf("Unable to marshal dialog request %v", err)
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	if err != nil {
		log.Printf("Error occurred while creating request to %s. %v", url, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		log.Printf("Error occurred while sending data to %s. %v", url, err)
		return
	}

	if response.StatusCode != 200 {
		log.Printf("Got %d while invoking %s.", response.StatusCode, url)
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
		log.Printf("Unable to marshal response")
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
		log.Println("Unable to marshal response")
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
