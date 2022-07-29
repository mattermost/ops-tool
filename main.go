package main

import (
	"context"
	"fmt"
	"net/http"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := server.New()

	go func() {
		select {
		case <-signalChanel:
			fmt.Println("Received an interrupt, stopping...")
		case <-ctx.Done():
			fmt.Println("Context done, stopping...")
		}
		srv.Stop()
	}()

	err := srv.Start(ctx)
	if err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
