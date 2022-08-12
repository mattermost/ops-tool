package config

import (
	"fmt"
	"io/ioutil"

	"github.com/mattermost/ops-tool/model"
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
	Command              string              `yaml:"command"`
	Token                string              `yaml:"token"`
	DialogURL            string              `yaml:"dialog_url"`
	DialogResponseURL    string              `yaml:"dialog_response_url"`
	SchedulerResponseURL string              `yaml:"scheduler_response_url"`
	Plugins              []CommandPlugin     `yaml:"plugins"`
	AccessControl        model.AccessControl `yaml:"access_control"`
}

type ScheduledCommandConfig struct {
	Name        string `yaml:"name"`
	Cron        string `yaml:"cron"`
	Command     string `yaml:"command"`
	Channel     string `yaml:"channel"`
	ResponseURL string `yaml:"response_url"`
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

type CommandPlugin struct {
	Name    string
	Only    []string
	Exclude []string
}

type plainCommandPlugin struct {
	Name    string   `yaml:"name"`
	Only    []string `yaml:"only"`
	Exclude []string `yaml:"exclude"`
}

// UnmarshalYAML implements the Unmarshaler interface.
// The CommandPlugin can either be a string corresponding to the name, or a CommandPlugin
func (cp *CommandPlugin) UnmarshalYAML(unmarshal func(interface{}) error) error {
	name := ""
	err := unmarshal(&name)
	if err == nil {
		*cp = CommandPlugin{Name: name}
		return nil
	}

	plugin := plainCommandPlugin{}
	err = unmarshal(&plugin)
	if err != nil {
		return err
	}

	cp.Name = plugin.Name
	cp.Only = plugin.Only
	cp.Exclude = plugin.Exclude

	if len(cp.Only) > 0 && len(cp.Exclude) > 0 {
		return errors.New("only and exclude cannot be both set")
	}
	return nil
}
