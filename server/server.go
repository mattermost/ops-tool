package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/julienschmidt/httprouter"
	"github.com/mattermost/ops-tool/config"
	"github.com/mattermost/ops-tool/log"
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
	log := log.FromContext(ctx)

	log.Info("Starting ops tool server...")

	log.Info("Loading config...")
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.WithError(err).Error("Failed to load config")
		return errors.Wrap(err, "failed to load config")
	}
	s.Config = cfg

	log.Debug("Loading plugins...")
	plugins, err := plugin.Load(cfg.PluginsConfig)
	if err != nil {
		return errors.Wrap(err, "failed to load plugins")
	}

	log.Debug("Loading commands...")
	commands, err := slashcommand.Load(plugins, cfg.CommandConfigurations)
	if err != nil {
		return errors.Wrap(err, "failed to load commands")
	}
	s.commands = commands

	log.Debug("Loading scheduled commands...")
	scheduler := gocron.NewScheduler(time.UTC)
	for _, scheduledCommand := range cfg.ScheduledCommands {
		log.Debug("Scheduling command '" + scheduledCommand.Name + "' for " + scheduledCommand.Cron)
		scheduler.Cron(scheduledCommand.Cron).DoWithJobDetails(s.scheduledCommandHandler, scheduledCommand)
	}
	scheduler.StartAsync()

	s.DialogStore = store.NewInMemoryDialogStore()

	// Prints all registered commands
	for i := range s.commands {
		log.Debugf("**/%s**", s.commands[i].Command)
		for j := range s.commands[i].Commands {
			log.Debugf("\t/%s %s [Name:%s | Description: %s]", s.commands[i].Command, s.commands[i].Commands[j].Command, s.commands[i].Commands[j].Name, s.commands[i].Commands[j].Description)
		}
	}

	router := httprouter.New()
	router.GET("/", indexHandler)
	router.GET("/healthz", healthHandler)
	router.POST("/hook", enhancedHandler(log, s.hookHandler))
	router.POST("/dialog", enhancedHandler(log, s.dialogHandler))

	s.server = &http.Server{Addr: cfg.Listen, Handler: router}
	log.Info("starting http server")

	return s.server.ListenAndServe()
}

type ReqIdKey struct{}

func enhancedHandler(logger *log.Logger, fn func(context.Context, http.ResponseWriter, *http.Request, httprouter.Params)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		requestID := randStr(16)
		ctx := log.WithRqId(r.Context(), requestID)

		fn(ctx, w, r, ps)
	}
}

// randStr generates a random string of a certain length
func randStr(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func (s *Server) findCommand(rootCommand string) *slashcommand.SlashCommand {
	for _, cmd := range s.commands {
		if strings.EqualFold(rootCommand, cmd.Command) {
			return &cmd
		}
	}
	return nil
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
