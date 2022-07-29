package config

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type RawMessage struct {
	unmarshal func(interface{}) error
}

func (msg *RawMessage) UnmarshalYAML(unmarshal func(interface{}) error) error {
	msg.unmarshal = unmarshal
	return nil
}

func (msg *RawMessage) Unmarshal(v interface{}) error {
	return msg.unmarshal(v)
}

type Config struct {
	Listen                string                   `yaml:"listen"`
	BaseURL               string                   `yaml:"base_url"`
	PluginsConfig         []PluginConfig           `yaml:"plugins"`
	CommandConfigurations []CommandConfig          `yaml:"commands"`
	ScheduledCommands     []ScheduledCommandConfig `yaml:"scheduler"`
}

type PluginConfig struct {
	Name   string     `yaml:"name"`
	File   string     `yaml:"file"`
	Config RawMessage `yaml:"config"`
}

type CommandConfig struct {
	Command           string   `yaml:"command"`
	Token             string   `yaml:"token"`
	DialogURL         string   `yaml:"dialog_url"`
	DialogResponseURL string   `yaml:"dialog_response_url"`
	Plugins           []string `yaml:"plugins"`
}

type ScheduledCommandConfig struct {
	Name     string   `yaml:"name"`
	Channel  string   `yaml:"channel"`
	Provider string   `yaml:"provider"`
	Command  string   `yaml:"command"`
	Args     []string `yaml:"args"`
	Cron     string   `yaml:"cron"`
	Hook     string   `yaml:"hook"`
}

func Load(path string) (*Config, error) {
	fmt.Println("Loading config from", path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config file %s", path)
	}

	fmt.Println("Parsing config")
	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal config file %s", path)
	}

	return &config, nil
}
