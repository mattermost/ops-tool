package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/ops-tool/version"
)

type healthResponse struct {
	Info *version.Info `json:"info"`
}

func indexHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte("This is the ops tool server."))
}

func healthHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := json.NewEncoder(w).Encode(healthResponse{Info: version.Full()})
	if err != nil {
		LogError(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func hookHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	command, err := ParseSlashCommand(r)
	if err != nil {
		WriteErrorResponse(w, NewError("Unable to parse incoming slash command info", err))
		return
	}
	if command.Token != Config.Token {
		WriteErrorResponse(w, NewError("Invalid comand token! Please check slash command tokens!", err))
		return
	}
	LogInfo("Received command: %s at channel %s from %s", command.Text, command.ChannelName, command.Username)
	if command.Text == "" || command.Text == "help" {
		msg := "| Command | Slash Command | Description |\n| :-- | :-- | :-- |"
		for key, opsCommand := range Commands {
			if opsCommand.CanTrigger(command.Username) {
				msg += fmt.Sprintf("\n| __%s__ | `%s %s` | *%s* |", opsCommand.Name, command.Command, key, opsCommand.Description)
			}

		}
		WriteEnrichedResponse(w, "Supported Commands", msg, "#0000ff", model.CommandResponseTypeEphemeral)
	} else {
		opsCommand, found := Commands[command.Text]
		if !found {
			WriteErrorResponse(w, NewError("Command not found", err))
			return
		}
		if !opsCommand.CanTrigger(command.Username) {
			WriteErrorResponse(w, NewError("You do not have permission to execute "+command.Command, err))
			return
		}
		output, err := opsCommand.Execute(command)
		if err != nil {
			LogError("Error occurred while executing command! %v", err)
			WriteErrorResponse(w, NewError("Command execution failed!", err))
		} else if opsCommand.Response.Generate {
			msgColor := "#000000"
			for _, responseColor := range opsCommand.Response.Colors {
				if responseColor.Status == output.Status {
					msgColor = responseColor.Color
					break
				}
			}

			buf := bytes.NewBufferString("")
			err = opsCommand.Response.Template.Execute(buf, output)
			WriteEnrichedResponse(w, opsCommand.Name, buf.String(), msgColor, opsCommand.Response.Type)
		}

	}

}

func Start() {
	LoadConfig("config.yaml")
	LoadCommands()
	LogInfo("Starting OpsTool")

	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/healthz", healthHandler)
	router.POST("/hook", hookHandler)

	LogInfo("Running OpsTool on port " + Config.Listen)
	if err := http.ListenAndServe(Config.Listen, router); err != nil {
		LogError(err.Error())
	}
}
