package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/ops-tool/version"
)

type healthResponse struct {
	Info *version.Info `json:"info"`
}

type HookResponse struct {
	Title        string
	Color        string
	ResponseType string
	Body         string
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

func helpHookHandler(slashCommand *MMSlashCommand) (*HookResponse, error) {
	keys := make([]string, 0, len(Providers))
	for key := range Providers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	msg := "| Command | Slash Command | Description |\n| :-- | :-- | :-- |"
	for key := range keys {
		opsCommand := Providers[keys[key]]
		msg += fmt.Sprintf("\n| __%s__ | `%s %s` | *%s* |", opsCommand.Name, slashCommand.Command, keys[key], opsCommand.Description)
	}
	return &HookResponse{
		Title:        "Provider List",
		Color:        "#000000",
		ResponseType: model.CommandResponseTypeEphemeral,
		Body:         msg,
	}, nil
}

func providerHelpHookHandler(slashCommand *MMSlashCommand, providerName string) (*HookResponse, error) {
	providerCommand, found := Providers[providerName]
	if !found {
		return nil, fmt.Errorf("%s not found", providerName)
	}
	sort.Slice(providerCommand.ProvidedCommands, func(i, j int) bool {
		return providerCommand.ProvidedCommands[i].Command < providerCommand.ProvidedCommands[j].Command
	})
	msg := "| Command | Slash Command | Description |\n| :-- | :-- | :-- |"
	for _, c := range providerCommand.ProvidedCommands {
		if c.CanTrigger(slashCommand.Username) {
			msg += fmt.Sprintf("\n| __%s__ | `%s %s %s %s` | *%s* |", c.Name, slashCommand.Command, providerCommand.Command, c.Command, c.SubCommand, c.Description)
		}
	}
	return &HookResponse{
		Title:        fmt.Sprintf("%s commands", providerCommand.Name),
		Color:        "#000000",
		ResponseType: model.CommandResponseTypeEphemeral,
		Body:         msg,
	}, nil
}

func providerCommandHookHandler(slashCommand *MMSlashCommand, providerName string, commandName string, args []string) (*HookResponse, error) {
	providerCommand, found := Providers[providerName]
	if !found {
		return nil, fmt.Errorf("%s not found", providerName)
	}
	subCommand := ""
	if len(args) > 0 {
		subCommand = args[0]
	}
	var opsCommand *OpsCommand
	commandFound := false
	for i := range providerCommand.ProvidedCommands {
		opsCommand = providerCommand.ProvidedCommands[i]
		if opsCommand.Command == commandName && (opsCommand.SubCommand == "" || opsCommand.SubCommand == subCommand) {
			commandFound = true
			break
		}
	}
	if !commandFound {
		return nil, fmt.Errorf("%s %s not found", providerName, commandName)
	}
	if opsCommand.SubCommand != "" {
		args = args[1:]
	}
	if !opsCommand.CanTrigger(slashCommand.Username) {
		return nil, fmt.Errorf("you are not allowed to execute %s %s", providerName, commandName)
	}
	if opsCommand.Dialog != nil {
		sendDialogCallback(opsCommand, slashCommand)
	} else {
		return executeCommand(opsCommand, slashCommand, args, map[string]string{})
	}
	return nil, nil
}

func executeCommand(opsCommand *OpsCommand, slashCommand *MMSlashCommand, args []string, envVars map[string]string) (*HookResponse, error) {
	output, err := opsCommand.Execute(slashCommand, args, envVars)
	if err != nil {
		return nil, err
	}
	if opsCommand.Response.Generate {
		msgColor := "#000000"
		for _, responseColor := range opsCommand.Response.Colors {
			if responseColor.Status == "" || responseColor.Status == output.Status { // Support default color
				msgColor = responseColor.Color
				break
			}
		}

		buf := bytes.NewBufferString("")
		err = opsCommand.Response.Template.Execute(buf, output)
		if err != nil {
			return nil, err
		}
		return &HookResponse{
			Title:        opsCommand.Name,
			Color:        msgColor,
			ResponseType: opsCommand.Response.Type,
			Body:         buf.String(),
		}, nil
	}
	return nil, nil
}

func sendDialogCallback(opsCommand *OpsCommand, slashCommand *MMSlashCommand) {
	callbackID := uuid.NewString()
	elements := make([]model.DialogElement, 0)
	for _, opsElem := range opsCommand.Dialog.Elements {
		elements = append(elements, model.DialogElement{
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
		})
	}

	request := &model.OpenDialogRequest{
		TriggerId: slashCommand.TriggerID,
		URL:       opsCommand.Dialog.CallbackURL,

		Dialog: model.Dialog{
			CallbackId:       callbackID,
			Title:            opsCommand.Dialog.Title,
			IntroductionText: opsCommand.Dialog.Text,
			Elements:         elements,
			NotifyOnCancel:   true, // We need to remove dialog session from map to save memory.
		},
	}
	SendDialogRequest(opsCommand.Dialog.URL, request)
	DialogSessions[callbackID] = &DialogSession{
		CallbackID:   callbackID,
		MMHookURL:    opsCommand.Dialog.MMHookURL,
		SlashCommand: slashCommand,
		OpsCommand:   opsCommand,
	}
	LogInfo("Dialog session is created for trigger %s with id %s", slashCommand.TriggerID, callbackID)
}

func hookHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	slashCommand, err := ParseSlashCommand(r)
	if err != nil {
		WriteErrorResponse(w, NewError("Unable to parse incoming slash command info", err))
		return
	}
	if slashCommand.Token != Config.Token {
		WriteErrorResponse(w, NewError("Invalid comand token! Please check slash command tokens!", err))
		return
	}
	LogInfo("Received command: %s at channel %s from %s", slashCommand.Text, slashCommand.ChannelName, slashCommand.Username)
	parsedCommand := strings.Fields(strings.TrimSpace(slashCommand.Text))
	var response *HookResponse

	switch len(parsedCommand) {
	case 0:
		response, err = helpHookHandler(slashCommand)
	case 1:
		response, err = providerHelpHookHandler(slashCommand, parsedCommand[0])
	default:
		response, err = providerCommandHookHandler(slashCommand, parsedCommand[0], parsedCommand[1], parsedCommand[2:])
	}

	if err != nil {
		LogError("Error while processing command %v", err)
		WriteErrorResponse(w, NewError("Command execution failed!", err))
	} else if response != nil {
		WriteEnrichedResponse(w, response.Title, response.Body, response.Color, response.ResponseType)
	}
}

func dialogHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	dialogSubmission, err := ParseDialogSubmission(r)
	if err != nil {
		WriteErrorResponse(w, NewError("Unable to parse dialog submission", err))
		return
	}
	LogInfo("Received valid dialog submission for session %s, canceled: %t", dialogSubmission.CallbackID, dialogSubmission.Canceled)
	dialogSession, found := DialogSessions[dialogSubmission.CallbackID]
	if !found {
		LogInfo("Dialog Session %s not found!", dialogSession.CallbackID)
		WriteResponse(w, "Session not found! Trigger slash command again!", model.CommandResponseTypeEphemeral)
		return
	}
	LogInfo("Found session executing command %s", dialogSession.OpsCommand.Name)
	delete(DialogSessions, dialogSession.CallbackID)
	LogInfo("Session %s is terminated. %d session is active.", dialogSession.CallbackID, len(DialogSessions))
	if dialogSubmission.Canceled {
		return
	}

	response, err := executeCommand(dialogSession.OpsCommand, dialogSession.SlashCommand, []string{}, dialogSubmission.Submission)
	if err != nil {
		LogError("Error while processing command %v", err)
		return
	} else if response != nil {
		WriteEnrichedResponse(w, response.Title, response.Body, response.Color, response.ResponseType)
	}
	if dialogSession.MMHookURL != "" {
		channel := dialogSession.SlashCommand.ChannelName
		if response.ResponseType == model.CommandResponseTypeEphemeral {
			channel = fmt.Sprintf("@%s", dialogSession.SlashCommand.Username)
		}
		SendViaIncomingHook(dialogSession.MMHookURL, channel, response.Title, response.Body, response.Color)
	}
}

func scheduledJobHandler(scheduledCommand *ScheduledCommand, job gocron.Job) {
	LogInfo("%s's last run: %s; next run: %s", scheduledCommand.Name, job.LastRun(), job.NextRun())
	providerCommand, found := Providers[scheduledCommand.Provider]
	if !found {
		return
	}
	var opsCommand *OpsCommand
	for i := range providerCommand.ProvidedCommands {
		if providerCommand.ProvidedCommands[i].Command == scheduledCommand.Command {
			opsCommand = providerCommand.ProvidedCommands[i]
			break
		}
	}
	if opsCommand == nil {
		return
	}
	output, err := opsCommand.Execute(&MMSlashCommand{}, scheduledCommand.Args, map[string]string{})
	if err != nil {
		LogError("Error occurred while executing command! %v", err)
	} else if opsCommand.Response.Generate {
		msgColor := "#000000"
		for _, responseColor := range opsCommand.Response.Colors {
			if responseColor.Status == output.Status {
				msgColor = responseColor.Color
				break
			}
		}

		buf := bytes.NewBufferString("")
		err = opsCommand.Response.Template.Execute(buf, &output)
		if err != nil {
			LogError("Error occurred while rendering response! %v", err)
		} else {
			SendViaIncomingHook(scheduledCommand.Hook, scheduledCommand.Channel, opsCommand.Name, buf.String(), msgColor)
		}
	}
}

var server *http.Server

func Start() {
	LoadConfig("config.yaml")
	LoadCommands()
	LogInfo("Starting OpsTool")

	LogInfo("Starting Scheduler")
	scheduler := gocron.NewScheduler(time.UTC)
	for _, scheduledCommand := range Config.ScheduledCommands {
		LogInfo("Scheduled Job %s for %s", scheduledCommand.Name, scheduledCommand.Cron)
		scheduler.Cron(scheduledCommand.Cron).DoWithJobDetails(scheduledJobHandler, scheduledCommand)
	}
	scheduler.StartAsync()
	LogInfo("Starting Http Router")
	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/healthz", healthHandler)
	router.POST("/hook", hookHandler)
	router.POST("/dialog", dialogHandler)

	LogInfo("Running OpsTool on port " + Config.Listen)
	server = &http.Server{Addr: Config.Listen, Handler: router}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			LogError(err.Error())
		}
	}()
}

func Stop() {
	if server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
	}
}
