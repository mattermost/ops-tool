package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"

	"gopkg.in/yaml.v2"
)

type OpsCommand struct {
	Command          string               `yaml:"command"`
	SubCommand       string               `yaml:"subcommand"`
	Name             string               `yaml:"name"`
	Description      string               `yaml:"description"`
	Provides         []string             `yaml:"provides"`
	ProvidedCommands []*OpsCommand        `yaml:"-"`
	Variables        []OpsCommandVariable `yaml:"vars"`
	Exec             []string             `yaml:"exec"`
	Response         OpsCommandResponse   `yaml:"response"`
	Users            []string             `yaml:"users"`
}

type OpsCommandVariable struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type OpsCommandResponse struct {
	Type           string             `yaml:"type"`
	Colors         []Color            `yaml:"colors"`
	TemplateString string             `yaml:"template"`
	Template       *template.Template `yaml:"-"`
	Generate       bool               `yaml:"-"`
}

type Color struct {
	Color  string `yaml:"color"`
	Status string `yaml:"status"`
}

type OpsCommandOutput struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

var Providers map[string]*OpsCommand = make(map[string]*OpsCommand)

func (cmd *OpsCommand) CanTrigger(username string) bool {
	canTrigger := true
	if cmd.Users != nil {
		canTrigger = false
		for _, user := range cmd.Users {
			if user == username {
				canTrigger = true
				break
			}
		}
	}
	return canTrigger
}

func (opsCmd *OpsCommand) Execute(mmCommand *MMSlashCommand, args []string) (*OpsCommandOutput, error) {
	output := &OpsCommandOutput{}
	for _, step := range opsCmd.Exec {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		script := step
		for i := range args {
			script = fmt.Sprintf("%s %s", script, args[i])
		}
		LogInfo("Will execute %s", script)
		cmd := exec.Command("/bin/bash", "-c", script)
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, fmt.Sprintf("CHANNEL_NAME=%s", mmCommand.ChannelName))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TEAM_NAME=%s", mmCommand.TeamName))
		cmd.Env = append(cmd.Env, fmt.Sprintf("USER_NAME=%s", mmCommand.Username))
		cmd.Env = append(cmd.Env, fmt.Sprintf("COMMAND_TEXT=%s", mmCommand.Text))
		for _, cmdVar := range opsCmd.Variables {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", cmdVar.Name, cmdVar.Value))
		}
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			LogError("Error while executing command %s.Err:%s.\n%v", opsCmd.Name, stderr.String(), err)
			return nil, err
		}
		data := stdout.Bytes()
		err = json.Unmarshal(data, output)
		if err != nil {
			LogError("Error while deserializing command output %s.Err:%s.\n%v", opsCmd.Name, string(data), err)
			return nil, err
		}
	}
	return output, nil
}

func LoadCommands() {
	loadCommands(Config.CommandConfigurations)
}

func loadCommands(commandsConfig []string) []*OpsCommand {
	providedCommands := make([]*OpsCommand, 0)
	for _, commandConfiguration := range commandsConfig {
		LogInfo("Loading commands from " + commandConfiguration)
		content, err := os.ReadFile(commandConfiguration)
		if err != nil {
			LogCritical("Error reading command file=%s err= %v", commandConfiguration, err)
		}
		commands := []OpsCommand{}
		err = yaml.Unmarshal(content, &commands)
		if err != nil {
			LogCritical("Error reading command file=%s err= %v", commandConfiguration, err)
		}
		for i, _ := range commands {
			command := commands[i]
			command.Response.Generate = false

			LogInfo("Command %s[%s]=%s", command.Name, command.Command, command.Description)

			if len(command.Provides) > 0 {
				command.ProvidedCommands = loadCommands(command.Provides)
				Providers[command.Command] = &command
			} else if command.Response.TemplateString != "" {
				command.Response.Generate = true
				t, err := createTemplate(command.Name, command.Response.TemplateString)
				if err != nil {
					LogCritical("Error rendering template file for command %s err= %v", command.Name, err)
				}
				command.Response.Template = t
			} else {
			}
			providedCommands = append(providedCommands, &command)
		}
	}
	return providedCommands
}
