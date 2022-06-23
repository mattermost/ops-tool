package server

import (
	"os"

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
	Type     string  `yaml:"type"`
	Colors   []Color `yaml:"colors"`
	Template string  `yaml:"template"`
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
			Commands[command.Command] = &command
		}
	}
}
