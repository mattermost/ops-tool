package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"gopkg.in/yaml.v2"
)

type OpsCommand struct {
	Command     string               `yaml:"command"`
	Name        string               `yaml:"name"`
	Description string               `yaml:"description"`
	Variables   []OpsCommandVariable `yaml:"vars"`
	Exec        []string             `yaml:"exec"`
	Response    OpsCommandResponse   `yaml:"response"`
	Users       []string             `yaml:"users"`
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

var Commands map[string]*OpsCommand = make(map[string]*OpsCommand)

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

func (opsCmd *OpsCommand) Execute(mmCommand *MMSlashCommand) (map[string]string, error) {
	var output map[string]string = make(map[string]string)
	for _, step := range opsCmd.Exec {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := exec.Command(step)
		cmd.Env = append(cmd.Env, fmt.Sprintf("CHANNEL_NAME=%s", mmCommand.ChannelName))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TEAM_NAME=%s", mmCommand.TeamName))
		cmd.Env = append(cmd.Env, fmt.Sprintf("USER_NAME=%s", mmCommand.Username))
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
		err = json.Unmarshal(data, &output)
		if err != nil {
			LogError("Error while deserializing command output %s.Err:%s.\n%v", opsCmd.Name, string(data), err)
			return nil, err
		}
	}
	return output, nil
}

func LoadCommands() {

	for _, commandConfiguration := range Config.CommandConfigurations {
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
		for _, command := range commands {
			LogInfo("Command %s[%s]=%s", command.Name, command.Command, command.Description)
			if command.Response.TemplateString != "" {
				command.Response.Generate = true
				t := template.New(command.Name)
				t, err := t.Parse(command.Response.TemplateString)
				if err != nil {
					LogCritical("Error rendering template file for command %s err= %v", command.Name, err)
				}
				command.Response.Template = t
			} else {
				command.Response.Generate = false
			}
			Commands[command.Command] = &command
		}
	}
}
