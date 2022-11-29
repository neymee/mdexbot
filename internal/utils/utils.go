package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/telebot.v3"
)

type ctxKey string

const (
	ctxKeyTraceID ctxKey = "trace_id"
)

func TelebotCtxSetup(c telebot.Context) {
	traceID := time.Now().UnixMicro()
	c.Set(string(ctxKeyTraceID), &traceID)
}

func ReqCtx(c telebot.Context) context.Context {
	ctx := context.Background()
	traceID, ok := c.Get(string(ctxKeyTraceID)).(*int64)
	if !ok {
		return ctx
	}
	ctx = context.WithValue(ctx, ctxKeyTraceID, traceID)
	return ctx
}

func Log(ctx context.Context, prefix ...string) *zerolog.Logger {
	logCtx := log.With()
	for i, p := range prefix {
		key := "prefix"
		if i > 0 {
			key = fmt.Sprint(key, i+1)
		}
		logCtx = logCtx.Str(key, p)
	}

	traceID, ok := ctx.Value(ctxKeyTraceID).(*int64)
	if ok && traceID != nil {
		logCtx = logCtx.Int64("trace_id", *traceID)
	}

	logger := logCtx.Logger()
	return &logger
}

func IsContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
