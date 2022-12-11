package bot

import (
	"context"
	"time"

	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/log"
	"github.com/neymee/mdexbot/internal/metrics"
	"gopkg.in/telebot.v3"
)

type sendOptionFunc func(*telebot.SendOptions)

func send(ctx context.Context, to domain.Recipient, text string, options ...sendOptionFunc) error {
	defer func(start time.Time) {
		metrics.MessageCounter.Inc()
		log.Log(ctx, "bot.send").Trace().
			Dur("duration", time.Since(start)).
			Int64("recipient", to.AsInt64()).
			Send()
	}(time.Now())

	opt := &telebot.SendOptions{
		ParseMode:             telebot.ModeHTML,
		DisableWebPagePreview: true,
		DisableNotification:   true,
	}

	for _, o := range options {
		o(opt)
	}

	_, err := bot.Send(to, text, opt)
	if err != nil {
		return err
	}

	return nil
}

func withKeyboard(keyboard [][]telebot.InlineButton) sendOptionFunc {
	return func(opt *telebot.SendOptions) {
		opt.ReplyMarkup = &telebot.ReplyMarkup{
			InlineKeyboard: keyboard,
		}
	}
}
