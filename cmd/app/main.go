package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/neymee/mdexbot/internal/app"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	app.Run(ctx)
}
