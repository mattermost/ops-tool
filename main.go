package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattermost/ops-tool/log"
	"github.com/mattermost/ops-tool/server"
	"github.com/mattermost/ops-tool/version"
)

func main() {
	log.AttachVersion(version.Full())

	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := server.New()

	go func() {
		select {
		case <-signalChanel:
			log.Default().Println("Received an interrupt, stopping...")
		case <-ctx.Done():
			log.Default().Println("Context done, stopping...")
		}
		srv.Stop()
	}()

	err := srv.Start(ctx)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
