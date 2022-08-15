package slashcommand

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mattermost/ops-tool/config"
	"github.com/mattermost/ops-tool/log"
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

	AccessControl model.AccessControl
}

var ErrCommandNotFound = errors.New("command not found")
var ErrTeamUnauthorized = errors.New("you cannot use this command from this team")
var ErrChannelUnauthorized = errors.New("you cannot use this command from this channel")
var ErrUserUnauthorized = errors.New("you are not authorized to use this command")

func (s *SlashCommand) Execute(ctx context.Context, mmCommand *model.MMSlashCommand, cmd string, args map[string]string) (*model.CommandResponse, error) {
	// mmCommand is nil for scheduled commands, no access control needed
	if mmCommand != nil && !s.AccessControl.IsEmpty() {
		err := s.accessControl(mmCommand)
		if err != nil {
			log.FromContext(ctx).
				WithError(err).
				Warn("restricted by access control")
			return nil, err
		}
	}

	if mmCommand == nil {
		mmCommand = &model.MMSlashCommand{}
	}

	for _, command := range s.Commands {
		if strings.EqualFold(command.Command, cmd) {
			ctx = enhanceContext(ctx, s, command)

			log.FromContext(ctx).Debugf("Executing command: %s", command.Command)
			return command.CommandHandler(ctx, mmCommand, args)
		}
	}

	return nil, ErrCommandNotFound
}

func (s *SlashCommand) accessControl(mmCommand *model.MMSlashCommand) error {
	if len(s.AccessControl.TeamID) > 0 && !contains(s.AccessControl.TeamID, mmCommand.TeamID) {
		return ErrTeamUnauthorized
	}
	if len(s.AccessControl.TeamName) > 0 && !contains(s.AccessControl.TeamName, mmCommand.TeamName) {
		return ErrTeamUnauthorized
	}
	if len(s.AccessControl.ChannelID) > 0 && !contains(s.AccessControl.ChannelID, mmCommand.ChannelID) {
		return ErrChannelUnauthorized
	}
	if len(s.AccessControl.ChannelName) > 0 && !contains(s.AccessControl.ChannelName, mmCommand.ChannelName) {
		return ErrChannelUnauthorized
	}
	if len(s.AccessControl.UserID) > 0 && !contains(s.AccessControl.UserID, mmCommand.UserID) {
		return ErrUserUnauthorized
	}
	if len(s.AccessControl.UserName) > 0 && !contains(s.AccessControl.UserName, mmCommand.Username) {
		return ErrUserUnauthorized
	}

	return nil
}

func (s *SlashCommand) ExecuteDialog(ctx context.Context, submission *model.DialogSubmission, cmd string, args map[string]string) (*model.CommandResponse, error) {
	for _, command := range s.Commands {
		if strings.EqualFold(command.Command, cmd) {
			ctx = enhanceContext(ctx, s, command)

			log.FromContext(ctx).Debugf("Executing dialog command: %s", command.Command)
			return command.DialogHandler(ctx, submission, args)
		}
	}

	return nil, fmt.Errorf("command %s not found", cmd)
}

func enhanceContext(ctx context.Context, s *SlashCommand, command model.Command) context.Context {
	ctx = log.WithPlugin(ctx, command.Plugin)
	ctx = log.WithSlashCommand(ctx, s.Command)
	return ctx
}

func Load(ctx context.Context, plugins []plugin.Plugin, cfg []config.CommandConfig) ([]SlashCommand, error) {
	commands := make([]SlashCommand, len(cfg))

	for i, commandCfg := range cfg {
		sCmd := SlashCommand{
			Command:              commandCfg.Command,
			Token:                commandCfg.Token,
			DialogURL:            commandCfg.DialogURL,
			DialogResponseURL:    commandCfg.DialogResponseURL,
			SchedulerResponseURL: commandCfg.SchedulerResponseURL,
			Commands:             []model.Command{},
			AccessControl:        commandCfg.AccessControl,
		}
		ctx = log.WithSlashCommand(ctx, sCmd.Command)

		for _, cmdPlugins := range commandCfg.Plugins {
			for _, plugin := range plugins {
				if plugin.Name == cmdPlugins.Name {
					ctx = log.WithPlugin(ctx, plugin.Name)
					pluginCmds := plugin.RegisterSlashCommand()
					for i := range pluginCmds {
						if len(cmdPlugins.Only) > 0 && !contains(cmdPlugins.Only, pluginCmds[i].Command) {
							log.FromContext(ctx).Debugf("Skipping command %s because it is not in the only list", pluginCmds[i].Command)
							continue
						} else if len(cmdPlugins.Exclude) > 0 && contains(cmdPlugins.Exclude, pluginCmds[i].Command) {
							log.FromContext(ctx).Debugf("Skipping command %s because it is in the exclude list", pluginCmds[i].Command)
							continue
						}

						pluginCmds[i].Plugin = plugin.Name
						sCmd.Commands = append(sCmd.Commands, pluginCmds[i])
					}
				}
			}
		}

		commands[i] = sCmd
	}

	return commands, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
