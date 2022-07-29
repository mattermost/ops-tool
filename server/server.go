package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mattermost/ops-tool/config"
	"github.com/mattermost/ops-tool/plugin"
	"github.com/mattermost/ops-tool/slashcommand"
	"github.com/mattermost/ops-tool/store"
	"github.com/mattermost/ops-tool/version"
	"github.com/pkg/errors"
)

type healthResponse struct {
	Info *version.Info `json:"info"`
}

type HookResponse struct {
	Title        string
	Color        string
	ResponseType string
	Body         string
}

func indexHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Write([]byte("This is the ops tool server."))
}

func healthHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	err := json.NewEncoder(w).Encode(healthResponse{Info: version.Full()})
	if err != nil {
		// LogError(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type Server struct {
	server   *http.Server
	commands []slashcommand.SlashCommand
	Config   *config.Config

	DialogStore store.DialogStore
}

func New() *Server {
	return &Server{}
}

func (s *Server) Start(ctx context.Context) error {
	log.Println("Starting ops tool server...")

	log.Println("Loading config...")
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		// LogError(err.Error())
		return errors.Wrap(err, "failed to load config")
	}
	s.Config = cfg

	log.Println("Loading plugins...")
	plugins, err := plugin.Load(cfg.PluginsConfig)
	if err != nil {
		return errors.Wrap(err, "failed to load plugins")
	}

	log.Println("Loading commands...")
	commands, err := slashcommand.Load(plugins, cfg.CommandConfigurations)
	if err != nil {
		return errors.Wrap(err, "failed to load commands")
	}

	s.commands = commands

	s.DialogStore = store.NewInMemoryDialogStore()

	for i := range s.commands {
		log.Printf("**/%s**", s.commands[i].Command)
		for j := range s.commands[i].Commands {
			log.Printf("\t/%s %s [Name:%s | Description: %s]", s.commands[i].Command, s.commands[i].Commands[j].Command, s.commands[i].Commands[j].Name, s.commands[i].Commands[j].Description)
		}
	}

	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/healthz", healthHandler)
	router.POST("/hook", s.hookHandler)
	router.POST("/dialog", s.dialogHandler)

	s.server = &http.Server{Addr: cfg.Listen, Handler: router}
	log.Println("starting http server")

	return s.server.ListenAndServe()
}

func (s *Server) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			panic(err) // failure/timeout shutting down the server gracefully
		}
	}
}
