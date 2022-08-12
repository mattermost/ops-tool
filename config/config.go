package config

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/mattermost/ops-tool/log"
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
	Command              string          `yaml:"command"`
	Token                string          `yaml:"token"`
	DialogURL            string          `yaml:"dialog_url"`
	DialogResponseURL    string          `yaml:"dialog_response_url"`
	SchedulerResponseURL string          `yaml:"scheduler_response_url"`
	Plugins              []CommandPlugin `yaml:"plugins"`
}

type ScheduledCommandConfig struct {
	Name        string `yaml:"name"`
	Cron        string `yaml:"cron"`
	Command     string `yaml:"command"`
	Channel     string `yaml:"channel"`
	ResponseURL string `yaml:"response_url"`
}

func Load(ctx context.Context, path string) (*Config, error) {
	log.FromContext(ctx).Info("loading config from " + path)
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read config file %s", path)
	}

	// Try to use the config file as a template.
	// If anything fails at this stage only log error but let the program continue
	// turns all env variable into template variables
	log.FromContext(ctx).Debug("parsing config file as template")
	content = parseConfigTemplate(ctx, content)

	log.FromContext(ctx).Debug("unmarshalling config file")
	var config Config
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal config file %s", path)
	}

	return &config, nil
}

func parseConfigTemplate(ctx context.Context, content []byte) []byte {
	t, err := template.New("config").Parse(string(content))
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("failed to create template from config file")
		return content
	}

	m := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			m[e[:i]] = e[i+1:]
		}
	}

	buf := &bytes.Buffer{}
	if err = t.Execute(buf, map[string]map[string]string{
		"Env": m,
	}); err != nil {
		log.FromContext(ctx).WithError(err).Error("failed to execute config template")
		return content
	}

	return buf.Bytes()
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
