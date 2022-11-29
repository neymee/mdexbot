package app

import (
	"context"

	"github.com/neymee/mdexbot/internal/bot"
	"github.com/neymee/mdexbot/internal/config"
	"github.com/neymee/mdexbot/internal/database"
	"github.com/neymee/mdexbot/internal/repo"
	"github.com/neymee/mdexbot/internal/service"
	"github.com/neymee/mdexbot/internal/utils"
)

func Run(ctx context.Context) {
	const method = "app.Run"

	cfg, err := config.Load()
	if err != nil {
		utils.Log(ctx, method).Error().Err(err).Send()
		return
	}

	config.ConfigGlobalLogger(cfg)

	utils.Log(ctx, method).Info().
		Interface("config", cfg).
		Msg("Starting with the following config")

	db, err := database.New(ctx, cfg)
	if err != nil {
		utils.Log(ctx, method).Error().Err(err).Send()
		return
	}

	r := repo.New(db)
	s := service.New(r.MDex, r.Storage, r.Storage)

	bot.Start(ctx, cfg, s)
	utils.Log(ctx, method).Info().Msg("App started")

	<-ctx.Done()

	utils.Log(ctx, method).Info().Msg("App is stopping...")
	bot.Stop()
}
