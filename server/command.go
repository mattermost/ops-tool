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
	Dialog           *OpsCommandDialog    `yaml:"dialog"`
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

type OpsCommandDialog struct {
	Title       string                     `yaml:"title"`
	URL         string                     `yaml:"url"`
	CallbackURL string                     `yaml:"callbackUrl"`
	MMHookURL   string                     `yaml:"hook"`
	Text        string                     `yaml:"introduction_text"`
	Elements    []*OpsCommandDialogElement `yaml:"elements"`
}

type OpsCommandDialogElement struct {
	DisplayName string `yaml:"display_name"`
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	SubType     string `yaml:"subtype"`
	Default     string `yaml:"default"`
	Placeholder string `yaml:"placeholder"`
	HelpText    string `yaml:"help_text"`
	Optional    bool   `yaml:"optional"`
	MinLength   int    `yaml:"min_length"`
	MaxLength   int    `yaml:"max_length"`
}

var Providers map[string]*OpsCommand = make(map[string]*OpsCommand)

func (opsCmd *OpsCommand) CanTrigger(username string) bool {
	canTrigger := true
	if opsCmd.Users != nil {
		canTrigger = false
		for _, user := range opsCmd.Users {
			if user == username {
				canTrigger = true
				break
			}
		}
	}
	return canTrigger
}

func (opsCmd *OpsCommand) Execute(mmCommand *MMSlashCommand, args []string, envValues map[string]string) (*OpsCommandOutput, error) {
	output := &OpsCommandOutput{}
	for _, step := range opsCmd.Exec {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		LogInfo("Will execute %s", step)
		cmd := exec.Command(step, args...)
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, fmt.Sprintf("CHANNEL_NAME=%s", mmCommand.ChannelName))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TEAM_NAME=%s", mmCommand.TeamName))
		cmd.Env = append(cmd.Env, fmt.Sprintf("USER_NAME=%s", mmCommand.Username))
		cmd.Env = append(cmd.Env, fmt.Sprintf("COMMAND_TEXT=%s", mmCommand.Text))
		for _, cmdVar := range opsCmd.Variables {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", cmdVar.Name, cmdVar.Value))
		}
		for envKey, envValue := range envValues {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envKey, envValue))
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
	loadCommands(Config.CommandConfigurations, []OpsCommandVariable{})
}

func loadCommands(commandsConfig []string, variables []OpsCommandVariable) []*OpsCommand {
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
		for i := range commands {
			command := commands[i]
			command.Response.Generate = false

			// inherit variables from parent and append our own
			command.Variables = dedup(append(variables, command.Variables...))

			LogInfo("Command %s[%s]=%s", command.Name, command.Command, command.Description)

			switch {
			case len(command.Provides) > 0:
				command.ProvidedCommands = loadCommands(command.Provides, command.Variables)
				Providers[command.Command] = &command
			case command.Response.TemplateString != "":
				command.Response.Generate = true
				t, err := createTemplate(command.Name, command.Response.TemplateString)
				if err != nil {
					LogCritical("Error rendering template file for command %s err= %v", command.Name, err)
				}
				command.Response.Template = t
			}

			providedCommands = append(providedCommands, &command)
		}
	}
	return providedCommands
}

func dedup(opsCommandVariable []OpsCommandVariable) []OpsCommandVariable {
	keys := make(map[string]int, 0)
	deduped := make([]OpsCommandVariable, 0)
	for i := range opsCommandVariable {
		// if we already saw this variable, replace its value
		if key, ok := keys[opsCommandVariable[i].Name]; ok {
			deduped[key] = opsCommandVariable[i]
		} else {
			keys[opsCommandVariable[i].Name] = i
			deduped = append(deduped, opsCommandVariable[i])
		}
	}

	return deduped
}
