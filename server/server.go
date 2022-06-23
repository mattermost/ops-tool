package server

import (
	"encoding/json"
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
	LogInfo("Received command: %s at channel %s from %s", command.Command, command.ChannelName, command.Username)
	WriteResponse(w, "Received", model.CommandResponseTypeEphemeral)
}

func Start() {
	LoadConfig("config.yaml")
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
