package server

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type OpsToolConfig struct {
	Listen string `yaml:"listen"`
	Token  string `yaml:"token"`
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
		LogCritical("Error reading config file=" + fileName + ", err=" + err.Error())
	}
	err = yaml.Unmarshal(content, Config)
	if err != nil {
		LogCritical("Error decoding config file=" + fileName + ", err=" + err.Error())
	}
}
