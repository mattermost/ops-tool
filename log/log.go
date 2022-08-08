package log

import (
	"context"
	"os"

	"github.com/mattermost/ops-tool/version"
	"go.uber.org/zap"
)

type correlationIDType int

const (
	requestIDKey correlationIDType = iota
	pluginKey
	slashCommandKey
)

type Logger struct {
	*zap.SugaredLogger
}

var logger *Logger

func init() {
	l, _ := zap.NewDevelopment(
		zap.Fields(
			zap.Int("pid", os.Getpid()),
		),
	)
	logger = &Logger{l.Sugar()}
}

func WithReqID(ctx context.Context, rqID string) context.Context {
	return context.WithValue(ctx, requestIDKey, rqID)
}

func WithPlugin(ctx context.Context, plugin string) context.Context {
	return context.WithValue(ctx, pluginKey, plugin)
}

func WithSlashCommand(ctx context.Context, slashcmd string) context.Context {
	return context.WithValue(ctx, slashCommandKey, slashcmd)
}

func Default() *Logger {
	return logger
}

func AttachVersion(v *version.Info) *Logger {
	logger = &Logger{logger.With(
		zap.String("app_name", v.Name),
		zap.String("version", v.Version),
		zap.String("build_hash", v.Hash),
		zap.String("build_date", v.Date),
	)}
	return logger
}

// Logger returns a zap logger with as much context as possible
func FromContext(ctx context.Context) *Logger {
	newLogger := logger.SugaredLogger
	if ctx != nil {
		if ctxRqID, ok := ctx.Value(requestIDKey).(string); ok {
			newLogger = newLogger.With(zap.String("req_id", ctxRqID))
		}
		if ctxPlugin, ok := ctx.Value(pluginKey).(string); ok {
			newLogger = newLogger.With(zap.String("plugin", ctxPlugin))
		}
		if ctxSlashCmd, ok := ctx.Value(slashCommandKey).(string); ok {
			newLogger = newLogger.With(zap.String("slash_command", ctxSlashCmd))
		}
	}

	return &Logger{newLogger}
}

func (l *Logger) Println(args ...interface{}) {
	l.SugaredLogger.Debug(args...)
}

func (l *Logger) Printf(format string, args ...interface{}) {
	l.SugaredLogger.Debugf(format, args...)
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.With(zap.Error(err))}
}
