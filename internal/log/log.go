package log

import (
	"context"
	"io"
	"os"

	"github.com/neymee/mdexbot/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ctxKey string

const (
	KeyTraceID ctxKey = "trace_id"
)

func Configure(cfg *config.Config) {
	writers := []io.Writer{}
	for _, o := range cfg.Log.Output {
		switch o {
		case "file":
			lumberjack := &cfg.Log.Lumberjack
			writers = append(writers, lumberjack)
			lumberjack.Rotate()
		case "stdout":
			if os.Getenv("PRETTY_LOGGING") == "true" {
				writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout})
			} else {
				writers = append(writers, os.Stdout)
			}
		}
	}
	log.Logger = log.Output(zerolog.MultiLevelWriter(writers...))

	lvl := map[string]zerolog.Level{
		"trace":    zerolog.TraceLevel,
		"debug":    zerolog.DebugLevel,
		"info":     zerolog.InfoLevel,
		"warn":     zerolog.WarnLevel,
		"error":    zerolog.ErrorLevel,
		"":         zerolog.ErrorLevel, // default
		"fatal":    zerolog.FatalLevel,
		"panic":    zerolog.PanicLevel,
		"no":       zerolog.NoLevel,
		"disabled": zerolog.Disabled,
	}[cfg.Log.Level]
	log.Logger = log.Level(lvl)
}

func Log(ctx context.Context, prefix string) *zerolog.Logger {
	logCtx := log.With().Str("prefix", prefix)

	traceID, ok := ctx.Value(KeyTraceID).(*int64)
	if ok && traceID != nil {
		logCtx = logCtx.Int64("trace_id", *traceID)
	}

	logger := logCtx.Logger()
	return &logger
}

// Error is shorthand for Log(...).Error().Err(err)
func Error(ctx context.Context, prefix string, err error) *zerolog.Event {
	return Log(ctx, prefix).Error().Err(err)
}
