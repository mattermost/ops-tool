package slashcommand

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/mattermost/ops-tool/config"
	"github.com/mattermost/ops-tool/model"
	"github.com/mattermost/ops-tool/plugin"
)

type SlashCommand struct {
	// Command is the root command that will be used to dispatch the received slash command
	Command string
	// Token is the token you get from Mattermost when creating a slash command
	Token string
	// DialogURL is the URL that will be used to open the dialog
	DialogURL string
	// DialogResponseURL is the URL that will be used to send the response to the dialog
	DialogResponseURL string
	// ScheduledResponseURL is the URL that will be used to send the response from the scheduled commands
	SchedulerResponseURL string

	Commands []model.Command
}

var ErrCommandNotFound = errors.New("command not found")

func (s *SlashCommand) Execute(mmCommand *model.MMSlashCommand, cmd string, args map[string]string) (*model.CommandResponse, error) {
	for _, command := range s.Commands {
		if strings.EqualFold(command.Command, cmd) {
			log.Println("Executing command: " + command.Command)
			return command.CommandHandler(mmCommand, args)
		}
	}

	return nil, ErrCommandNotFound
}

func (s *SlashCommand) ExecuteDialog(submission *model.DialogSubmission, cmd string, args map[string]string) (*model.CommandResponse, error) {
	for _, command := range s.Commands {
		if strings.EqualFold(command.Command, cmd) {
			log.Println("Executing command: " + command.Command)
			return command.DialogHandler(submission, args)
		}
	}

	return nil, fmt.Errorf("command %s not found", cmd)
}

func Load(plugins []plugin.Plugin, cfg []config.CommandConfig) ([]SlashCommand, error) {
	commands := make([]SlashCommand, len(cfg))

	for i, commandCfg := range cfg {
		sCmd := SlashCommand{
			Command:              commandCfg.Command,
			Token:                commandCfg.Token,
			DialogURL:            commandCfg.DialogURL,
			DialogResponseURL:    commandCfg.DialogResponseURL,
			SchedulerResponseURL: commandCfg.SchedulerResponseURL,
			Commands:             []model.Command{},
		}

		for _, cmdPlugins := range commandCfg.Plugins {
			for _, plugin := range plugins {
				if plugin.Name == cmdPlugins {
					pluginCmds := plugin.RegisterSlashCommand()
					log.Printf("plugin %s registered %d commands", plugin.Name, len(pluginCmds))
					sCmd.Commands = append(sCmd.Commands, pluginCmds...)
				}
			}
		}

		commands[i] = sCmd
	}

	return commands, nil
}
