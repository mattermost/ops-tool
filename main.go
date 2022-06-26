package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/mattermost/ops-tool/server"
)

func main() {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	server.Start()

	<-signalChanel

	server.Stop()
}
