package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/neymee/mdexbot/internal/bot/lang"
	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/service"
	"github.com/neymee/mdexbot/internal/utils"
	"gopkg.in/telebot.v3"
)

func runUpdatesChecker(ctx context.Context, s *service.Services) {
	checkUpdates(ctx, s)

	t := time.NewTicker(time.Hour)
	for {
		select {
		case <-t.C:
			checkUpdates(ctx, s)
		case <-ctx.Done():
			return
		}
	}
}

func checkUpdates(ctx context.Context, s *service.Services) {
	const method = "checkUpdates"

	defer func() {
		if err := recover(); err != nil {
			utils.Log(ctx, method).Error().Interface("panic", err).Msg("Panic recovered")
		}
	}()

	updates, err := s.Subscription.Updates(ctx)
	if err != nil {
		utils.Log(ctx, method).Error().Err(err).Msg("fetching updates error")
		return
	}

	for _, upd := range updates {
		text, keyboard := buildUpdateMessage(upd.MangaTitle, upd.Language, upd.NewChapters)

		for _, rec := range upd.Recipients {
			err := send(ctx, rec, text, withKeyboard(keyboard))

			if tbErr := new(telebot.Error); errors.As(err, &tbErr) && tbErr.Code == 403 {
				// user banned the bot, delete all their subscriptions
				err := s.Subscription.UnsubscribeAll(ctx, rec)
				if err != nil {
					utils.Log(ctx, method).Error().
						Err(err).
						Int64("recipient", rec.AsInt64()).
						Msg("UnsubscribeAll error")
				}

				err = s.Conversation.DeleteConversationContext(ctx, rec)
				if err != nil {
					utils.Log(ctx, method).Error().
						Err(err).
						Int64("recipient", rec.AsInt64()).
						Msg("DeleteConversationContext error")
				}

				utils.Log(ctx, "Send").Warn().
					Int64("recipient", rec.AsInt64()).
					Msg("The recipient has banned the bot and theirs subscriptions have been removed")

			} else if err != nil {
				utils.Log(ctx, method).Error().
					Err(err).
					Int64("recipient", rec.AsInt64()).
					Msg("error during sending message")
			}
		}
	}
}

func buildUpdateMessage(
	mangaTitle string,
	mangaLang string,
	chapters []domain.Chapter,
) (text string, keyboard [][]telebot.InlineButton) {
	first := chapters[0]
	if len(chapters) == 1 {
		text = lang.NewChapterSingle(mangaTitle, lang.GetFlagOrLang(mangaLang), first.Chapter, first.Title, first.Volume)
	} else {
		text = lang.NewChapterMulti(mangaTitle, lang.GetFlagOrLang(mangaLang), len(chapters))
	}

	var link string
	if first.ExternalUrl != "" {
		link = first.ExternalUrl
	} else {
		link = fmt.Sprintf("%s/chapter/%s", MangaDexURL, first.ID)
	}

	keyboard = [][]telebot.InlineButton{
		{
			{
				Text: "Read",
				URL:  link,
			},
		},
	}

	return
}
