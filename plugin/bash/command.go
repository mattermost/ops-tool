package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strings"

	"github.com/mattermost/ops-tool/log"
	"github.com/mattermost/ops-tool/model"
)

type BashVars struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type BashResponse struct {
	Type           string             `yaml:"type"`
	Colors         []model.Color      `yaml:"colors"`
	TemplateString string             `yaml:"template"`
	Template       *template.Template `yaml:"-"`
	Generate       bool               `yaml:"-"`
}

type BashCommand struct {
	Command     string     `yaml:"command"`
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Usage       string     `yaml:"usage"`
	Variables   []BashVars `yaml:"vars"`
	Provides    []string   `yaml:"provides"`
	Exec        []string   `yaml:"exec"`

	// Responses
	Dialog   model.Dialog `yaml:"dialog"`
	Response BashResponse `yaml:"response"`

	// Deprecated: prefer having  a complete command instead
	Subcommand string `yaml:"subcommand"`
}

type BashCommandOutput struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

func (c *BashCommand) Execute(ctx context.Context, metadata BashMetadata, args map[string]string, submission map[string]string) (*BashCommandOutput, error) {
	log := log.FromContext(ctx)

	var output BashCommandOutput
	for _, step := range c.Exec {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := exec.Command(step)
		cmd.Env = append(cmd.Env, os.Environ()...)
		cmd.Env = append(cmd.Env, fmt.Sprintf("CHANNEL_ID=%s", metadata.ChannelID))
		cmd.Env = append(cmd.Env, fmt.Sprintf("TEAM_ID=%s", metadata.TeamID))
		cmd.Env = append(cmd.Env, fmt.Sprintf("USER_ID=%s", metadata.UserID))
		for _, cmdVar := range c.Variables {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", cmdVar.Name, cmdVar.Value))
		}
		for name, value := range args {
			cmd.Env = append(cmd.Env, fmt.Sprintf("ARG_%s=%s", strings.ToUpper(name), value))
		}
		for name, value := range submission {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, value))
		}

		log.Printf("Env : %#v", cmd.Env)

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			log.WithError(err).Errorf("Error while executing command %s.\n%v\n", c.Name, stderr.String(), err)
			return nil, err
		}
		data := stdout.Bytes()
		err = json.Unmarshal(data, &output)
		if err != nil {
			log.WithError(err).Errorf("Error while deserializing command output %s.\n%v", c.Name, string(data), err)
			return nil, err
		}
	}

	return &output, nil
}
