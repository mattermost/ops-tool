package server

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type OpsToolConfig struct {
	Listen                string              `yaml:"listen"`
	Token                 string              `yaml:"token"`
	CommandConfigurations []string            `yaml:"commands"`
	ScheduledCommands     []*ScheduledCommand `yaml:"scheduler"`
}

type ScheduledCommand struct {
	Name     string   `yaml:"name"`
	Channel  string   `yaml:"channel"`
	Provider string   `yaml:"provider"`
	Command  string   `yaml:"command"`
	Args     []string `yaml:"args"`
	Cron     string   `yaml:"cron"`
	Hook     string   `yaml:"hook"`
}

var Config *OpsToolConfig = &OpsToolConfig{}

func FindConfigFile(fileName string) string {
	if _, err := os.Stat(fileName); err == nil {
		fileName, _ = filepath.Abs(fileName)
	} else if _, err := os.Stat("./config/" + fileName); err == nil {
		fileName, _ = filepath.Abs("./config/" + fileName)
	}
	return fileName
}

func LoadConfig(fileName string) {
	fileName = FindConfigFile(fileName)
	LogInfo("Loading " + fileName)

	content, err := os.ReadFile(fileName)
	if err != nil {
		LogCritical("Error reading config file=%s err= %v", fileName, err)
	}
	err = yaml.Unmarshal(content, Config)
	if err != nil {
		LogCritical("Error decoding config file=%s err= %v", fileName, err)
	}
}
