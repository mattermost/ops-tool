package server

import (
	"fmt"
	"strings"

	"github.com/mattermost/ops-tool/slashcommand"
)

func (s *Server) fullHelp() string {
	builder := strings.Builder{}
	first := true
	for i := range s.commands {
		if !first {
			builder.WriteString("\n")
		}
		fmt.Fprintf(&builder, "## /%s\n", s.commands[i].Command)
		builder.WriteString(s.helpForCommand(s.commands[i]))
		first = false
	}

	return builder.String()
}

func (s *Server) helpForCommand(cmd slashcommand.SlashCommand) string {
	builder := strings.Builder{}
	builder.WriteString("| **Command** | **Name** | **Description** | **Usage** |\n")
	builder.WriteString("|:---:|:---:|:---:|:---:|\n")
	for i := range cmd.Commands {
		fmt.Fprintf(
			&builder,
			"| /%s %s | %s | %s | %s |\n",
			cmd.Command, cmd.Commands[i].Command,
			cmd.Commands[i].Name,
			cmd.Commands[i].Description,
			cmd.Commands[i].Usage,
		)
	}
	return builder.String()
}
