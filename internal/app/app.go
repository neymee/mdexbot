package app

import (
	"context"

	"github.com/neymee/mdexbot/internal/bot"
	"github.com/neymee/mdexbot/internal/config"
	"github.com/neymee/mdexbot/internal/database"
	"github.com/neymee/mdexbot/internal/log"
	"github.com/neymee/mdexbot/internal/repo"
	"github.com/neymee/mdexbot/internal/service"
)

func Run(ctx context.Context) {
	const method = "app.Run"

	cfg, err := config.Load()
	if err != nil {
		log.Error(ctx, method, err).Send()
		return
	}

	log.Configure(cfg)

	log.Log(ctx, method).Info().
		Interface("config", cfg).
		Msg("Starting with the following config")

	db, err := database.New(ctx, cfg)
	if err != nil {
		log.Error(ctx, method, err).Send()
		return
	}

	r := repo.New(db)
	s := service.New(r.MDex, r.Storage, r.Storage)

	err = bot.Start(ctx, cfg, s)
	if err != nil {
		log.Error(ctx, method, err).Send()
	}

	log.Log(ctx, method).Info().Msg("App started")

	<-ctx.Done()

	log.Log(ctx, method).Info().Msg("App is stopping...")
	bot.Stop()
}
