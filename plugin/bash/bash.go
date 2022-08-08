package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"

	"github.com/mattermost/ops-tool/config"
	"github.com/mattermost/ops-tool/log"
	"github.com/mattermost/ops-tool/model"
	"github.com/mattermost/ops-tool/plugin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type BashPlugin struct {
	commands []BashCommand
}

type BashMetadata struct {
	ChannelID string
	UserID    string
	TeamID    string
}

type Config struct {
	Files []string `yaml:"files"`
}

func New(cfg config.RawMessage) (plugin.Interface, error) {
	var localCfg Config
	err := cfg.Unmarshal(&localCfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	plgn := &BashPlugin{}

	err = plgn.LoadFromFiles(localCfg.Files, "", []BashVars{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to load from files")
	}

	return plgn, nil
}

func (p *BashPlugin) LoadFromFiles(files []string, parent string, vars []BashVars) error {
	// load file content
	for _, file := range files {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return errors.Wrapf(err, "failed to read file %s", file)
		}

		// unmarshal yaml
		var cmds []BashCommand
		err = yaml.Unmarshal(content, &cmds)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal file %s", file)
		}

		for _, cmd := range cmds {
			if len(cmd.Variables) > 0 {
				vars = append(vars, cmd.Variables...)
			}

			if len(cmd.Provides) != 0 {
				err := p.LoadFromFiles(cmd.Provides, cmd.Command, vars)
				if err != nil {
					return errors.Wrapf(err, "failed to load from files")
				}
			} else {
				cmd.Command = parent + " " + cmd.Command
				if cmd.Subcommand != "" {
					cmd.Command += " " + cmd.Subcommand
				}

				cmd.Command = strings.TrimSpace(cmd.Command)

				if cmd.Response.TemplateString == "" {
					cmd.Response.TemplateString = "{{toHTMLUnescapedYaml .}}"
				}

				t, err := createTemplate(cmd.Name, cmd.Response.TemplateString)
				if err != nil {
					return errors.Wrap(err, "failed to create template")
				}
				cmd.Response.Template = t

				p.registerCommand(cmd)
			}
		}
	}

	return nil
}

func (p *BashPlugin) registerCommand(cmd BashCommand) {
	p.commands = append(p.commands, cmd)
}

func (p *BashPlugin) RegisterSlashCommand() []model.Command {
	commands := make([]model.Command, 0, len(p.commands))

	for i := range p.commands {
		cmd := p.commands[i]
		commands = append(commands, model.Command{
			Command:     cmd.Command,
			Name:        cmd.Name,
			Description: cmd.Description,
			Usage:       cmd.Usage,
			CommandHandler: func(ctx context.Context, mmCommand *model.MMSlashCommand, args map[string]string) (*model.CommandResponse, error) {
				log := log.FromContext(ctx)
				log.Println("Command:", cmd.Name)

				// if we have a dialog, return it
				if cmd.Dialog.Title != "" {
					return &model.CommandResponse{
						Type:   model.CommandResponseTypeDialog,
						Dialog: cmd.Dialog,
					}, nil
				}

				cmdOutput, err := cmd.Execute(ctx, BashMetadata{
					ChannelID: mmCommand.ChannelID,
					UserID:    mmCommand.UserID,
					TeamID:    mmCommand.TeamID,
				}, args, nil)
				if err != nil {
					return nil, err
				}

				msgColor := "#000000"
				for _, responseColor := range cmd.Response.Colors {
					if responseColor.Status == cmdOutput.Status {
						msgColor = responseColor.Color
						break
					}
				}

				buf := bytes.NewBufferString("")
				err = cmd.Response.Template.Execute(buf, cmdOutput)
				if err != nil {
					return nil, errors.Wrap(err, "failed to execute template")
				}

				return &model.CommandResponse{
					Type: model.CommandResponseType(cmd.Response.Type),
					Message: model.Message{
						Title:        cmd.Name,
						Color:        msgColor,
						ResponseType: cmd.Response.Type,
						Body:         buf.String(),
					},
				}, nil
			},
			DialogHandler: func(ctx context.Context, submission *model.DialogSubmission, args map[string]string) (*model.CommandResponse, error) {
				log := log.FromContext(ctx)

				log.Println("Dialog:", cmd.Name)

				cmdOutput, err := cmd.Execute(ctx, BashMetadata{
					ChannelID: submission.ChannelID,
					UserID:    submission.UserID,
					TeamID:    submission.TeamID,
				}, args, submission.Submission)
				if err != nil {
					return nil, err
				}

				msgColor := "#000000"
				for _, responseColor := range cmd.Response.Colors {
					if responseColor.Status == cmdOutput.Status {
						msgColor = responseColor.Color
						break
					}
				}

				buf := bytes.NewBufferString("")
				err = cmd.Response.Template.Execute(buf, cmdOutput)
				if err != nil {
					return nil, errors.Wrap(err, "failed to execute template")
				}

				return &model.CommandResponse{
					Type: model.CommandResponseType(cmd.Response.Type),
					Message: model.Message{
						Title:        cmd.Name,
						Color:        msgColor,
						ResponseType: cmd.Response.Type,
						Body:         buf.String(),
					},
				}, nil
			},
		})
	}
	return commands
}
