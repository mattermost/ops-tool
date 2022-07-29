package log

import (
	"context"
	"os"

	"go.uber.org/zap"
)

type correlationIdType int

const (
	requestIdKey correlationIdType = iota
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

func WithRqId(ctx context.Context, rqId string) context.Context {
	return context.WithValue(ctx, requestIdKey, rqId)
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

// Logger returns a zap logger with as much context as possible
func FromContext(ctx context.Context) *Logger {
	newLogger := logger.SugaredLogger
	if ctx != nil {
		if ctxRqId, ok := ctx.Value(requestIdKey).(string); ok {
			newLogger = newLogger.With(zap.String("req_id", ctxRqId))
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
