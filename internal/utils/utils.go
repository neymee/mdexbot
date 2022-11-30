package utils

import (
	"context"
	"time"

	"github.com/neymee/mdexbot/internal/log"
	"gopkg.in/telebot.v3"
)

func TelebotCtxSetup(c telebot.Context) {
	traceID := time.Now().UnixMicro()
	c.Set(string(log.KeyTraceID), &traceID)
}

func ReqCtx(c telebot.Context) context.Context {
	ctx := context.Background()
	traceID, ok := c.Get(string(log.KeyTraceID)).(*int64)
	if !ok {
		return ctx
	}
	ctx = context.WithValue(ctx, log.KeyTraceID, traceID)
	return ctx
}

func IsContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
