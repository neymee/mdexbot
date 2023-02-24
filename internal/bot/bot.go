package bot

import (
	"context"
	"time"

	"github.com/neymee/mdexbot/internal/config"
	"github.com/neymee/mdexbot/internal/service"
	"gopkg.in/telebot.v3"
)

var (
	bot *telebot.Bot
)

func Start(
	ctx context.Context,
	cfg *config.Config,
	services *service.Services,
) error {
	var err error
	bot, err = telebot.NewBot(
		telebot.Settings{
			Token:  cfg.Bot.Token,
			Poller: &telebot.LongPoller{Timeout: 30 * time.Second},
		},
	)
	if err != nil {
		return err
	}

	initHandlers(bot, services)

	go bot.Start()
	go runUpdatesChecker(ctx, cfg, services)

	return nil
}

func Stop() {
	if bot != nil {
		bot.Stop()
	}
}
