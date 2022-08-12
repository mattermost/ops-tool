package server

import (
	"context"

	"github.com/go-co-op/gocron"
	"github.com/mattermost/ops-tool/config"
	"github.com/mattermost/ops-tool/log"
)

func (s *Server) scheduledCommandHandler(scheduledCommand config.ScheduledCommandConfig, job gocron.Job) {
	ctx := context.Background()
	log := log.FromContext(ctx)

	log.Printf("%s's last run: %s; next run: %s", scheduledCommand.Name, job.LastRun(), job.NextRun())

	rootCmd, cmdText, args, err := ParseCommand(scheduledCommand.Command)
	if err != nil {
		log.Printf("Failed to parse command: %s", err)
		return
	}

	cmd := s.findCommand(rootCmd)
	if cmd == nil {
		log.Printf("command not found: %s", rootCmd)
		return
	}

	response, err := cmd.Execute(ctx, nil, cmdText, args)
	if err != nil {
		log.Printf("error executing command: %s", err.Error())
		return
	}

	dest := scheduledCommand.ResponseURL
	if dest == "" {
		dest = cmd.SchedulerResponseURL
	}

	SendViaIncomingHook(dest, scheduledCommand.Channel, scheduledCommand.Name, response.Message.Body, response.Message.Color)
}
