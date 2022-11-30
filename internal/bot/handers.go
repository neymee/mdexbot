package bot

import (
	"errors"
	"fmt"
	"time"

	"github.com/neymee/mdexbot/internal/bot/lang"
	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/service"
	"github.com/neymee/mdexbot/internal/service/subscription"
	"github.com/neymee/mdexbot/internal/utils"
	"gopkg.in/telebot.v3"
)

type InternalError error

func initHandlers(bot *telebot.Bot, s *service.Services) {
	bot.Handle(domain.CmdStart.Endpoint(), onStart(s), middlewares(domain.CmdStart)...)

	bot.Handle(domain.CmdText.Endpoint(), onText(s), middlewares(domain.CmdText)...)
	bot.Handle(domain.CmdSubscribe.Endpoint(), onSubscribe(s), middlewares(domain.CmdSubscribe)...)
	bot.Handle(domain.CmdSubscribeBtn.Endpoint(), onSubscribeBtn(s), middlewares(domain.CmdSubscribeBtn)...)

	bot.Handle(domain.CmdUnsubscribe.Endpoint(), onUnsubscribe(s), middlewares(domain.CmdUnsubscribe)...)
	bot.Handle(domain.CmdUnsubscribeBtn.Endpoint(), onUnsubscribeBtn(s), middlewares(domain.CmdUnsubscribeBtn)...)

	bot.Handle(domain.CmdList.Endpoint(), onList(s), middlewares(domain.CmdList)...)
	bot.Handle(domain.CmdCancel.Endpoint(), onCancel(s), middlewares(domain.CmdCancel)...)

	bot.OnError = func(err error, c telebot.Context) {
		utils.Log(utils.ReqCtx(c), "bot.onError").Error().
			Err(err).
			Int64("recipient", c.Chat().ID).
			Int("message_id", c.Message().ID).
			Msg("Error during processing request")

		if _, ok := err.(InternalError); ok {
			send(utils.ReqCtx(c), domain.RecipientFromInt64(c.Chat().ID), lang.ErrInternalError())
		}
	}
}

func middlewares(method domain.Command) []telebot.MiddlewareFunc {
	return []telebot.MiddlewareFunc{
		func(next telebot.HandlerFunc) telebot.HandlerFunc {
			// setup context
			return func(c telebot.Context) error {
				utils.TelebotCtxSetup(c)
				return next(c)
			}
		},
		func(next telebot.HandlerFunc) telebot.HandlerFunc {
			// recover panic
			return func(c telebot.Context) error {
				defer func() {
					if err := recover(); err != nil {
						utils.Log(utils.ReqCtx(c), method.String()).Error().
							Interface("panic", err).
							Msg("Panic recovered")
					}
				}()
				return next(c)
			}
		},
		func(next telebot.HandlerFunc) telebot.HandlerFunc {
			// log request
			return func(c telebot.Context) error {
				utils.Log(utils.ReqCtx(c), method.String()).Trace().
					Int64("chat_id", c.Chat().ID).
					Int("message_id", c.Message().ID).
					Str("text", c.Text()).
					Str("data", c.Data()).
					Msg("Request received")
				return next(c)
			}
		},
		func(next telebot.HandlerFunc) telebot.HandlerFunc {
			// duration
			return func(c telebot.Context) error {
				defer func(start time.Time) {
					utils.Log(utils.ReqCtx(c), method.String()).Trace().
						Dur("duration", time.Since(start)).
						Int64("chat_id", c.Chat().ID).
						Msg("Request processed")
				}(time.Now())
				return next(c)
			}
		},
		func(next telebot.HandlerFunc) telebot.HandlerFunc {
			// automatically respond to callback if any
			return func(c telebot.Context) error {
				if c.Callback() != nil {
					defer c.Respond()
				}
				return next(c)
			}
		},
	}
}

func onStart(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		return send(
			utils.ReqCtx(c),
			domain.RecipientFromInt64(c.Chat().ID),
			lang.Start(),
		)
	}
}

func onText(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := utils.ReqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		cmd, err := s.Conversation.ConversationContext(ctx, rec)
		if err != nil {
			return InternalError(err)
		}

		if cmd != domain.CmdSubscribe {
			return send(ctx, rec, lang.Start())
		}

		mangaID, err := mangaIDFromURL(c.Text())
		if err != nil {
			return send(ctx, rec, lang.SubscribeErrInvalidLink(c.Text()))
		}

		manga, err := s.Subscription.Manga(ctx, mangaID)
		if err != nil {
			return InternalError(err)
		}

		err = s.Conversation.DeleteConversationContext(ctx, rec)
		if err != nil {
			return InternalError(err)
		}

		keyboard := buildLanguageButtons(manga)

		return send(
			ctx,
			rec,
			lang.SubscribeChooseLanguage(manga.GetTitle()),
			withKeyboard(keyboard),
		)
	}
}

func onSubscribe(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := utils.ReqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		err := s.Conversation.SetConversationContext(ctx, rec, domain.CmdSubscribe)
		if err != nil {
			return InternalError(err)
		}

		return send(ctx, rec, lang.SubscribeInit())
	}
}

func onSubscribeBtn(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := utils.ReqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		mangaID, mangaLang, err := parseButtonData(c.Callback().Data)
		if err != nil {
			return InternalError(err)
		}

		sub, err := s.Subscription.Subscribe(ctx, rec, mangaID, mangaLang)
		if alsErr := new(subscription.AlreadySubscribedError); errors.As(err, &alsErr) {
			return send(ctx, rec, lang.SubscribeAllreadyFollowing(alsErr.Manga, alsErr.Lang))
		} else if err != nil {
			return InternalError(err)
		}

		return send(
			ctx,
			rec,
			lang.SubscribeConfirmed(sub.MangaTitle, lang.GetFlagOrLang(sub.Language)),
		)
	}
}

func onUnsubscribe(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := utils.ReqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		cmd, err := s.Conversation.ConversationContext(ctx, rec)
		if err != nil {
			return InternalError(err)
		} else if cmd != "" {
			return send(ctx, rec, lang.ErrWrongContext(cmd))
		}

		subs, err := s.Subscription.List(ctx, rec)
		if err != nil {
			return InternalError(err)
		} else if len(subs) == 0 {
			return send(ctx, rec, lang.UnsubscribeNoSubs())
		}

		keyboard := [][]telebot.InlineButton{}

		for _, sub := range subs {
			keyboard = append(keyboard, []telebot.InlineButton{
				{
					Text:   fmt.Sprintf("[%s] %s", lang.GetFlagOrLang(sub.Language), sub.MangaTitle),
					Data:   formatButtonData(sub.MangaID, sub.Language),
					Unique: domain.CmdUnsubscribeBtn.String(),
				},
			})
		}

		return send(ctx, rec, lang.UnsubscribeChooseSub(), withKeyboard(keyboard))
	}
}

func onUnsubscribeBtn(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := utils.ReqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		mangaID, mangaLang, err := parseButtonData(c.Callback().Data)
		if err != nil {
			return InternalError(err)
		}

		sub, err := s.Subscription.Unsubscribe(ctx, rec, mangaID, mangaLang)
		if err != nil {
			return InternalError(err)
		}

		return send(
			ctx,
			rec,
			lang.UnsubscribeConfirmed(sub.MangaTitle, lang.GetFlagOrLang(sub.Language)),
		)
	}
}

func onList(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := utils.ReqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		cmd, err := s.Conversation.ConversationContext(ctx, rec)
		if err != nil {
			return InternalError(err)
		} else if cmd != "" {
			return send(ctx, rec, lang.ErrWrongContext(cmd))
		}

		subs, err := s.Subscription.List(ctx, rec)
		if err != nil {
			return InternalError(err)
		} else if len(subs) == 0 {
			return send(ctx, rec, lang.ListNoSubs())
		}

		keyboard := [][]telebot.InlineButton{}
		for _, s := range subs {
			keyboard = append(keyboard, []telebot.InlineButton{
				{
					Text: fmt.Sprintf("[%s] %s", lang.GetFlagOrLang(s.Language), s.MangaTitle),
					URL:  fmt.Sprintf("%s/title/%s", MangaDexURL, s.MangaID),
				},
			})
		}

		return send(ctx, rec, lang.List(), withKeyboard(keyboard))
	}
}

func onCancel(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := utils.ReqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		cmd, err := s.Conversation.ConversationContext(ctx, rec)
		if err != nil {
			return InternalError(err)
		} else if cmd == "" {
			return send(ctx, rec, lang.CancelNoCommands())
		}

		err = s.Conversation.DeleteConversationContext(ctx, rec)
		if err != nil {
			return InternalError(err)
		}

		return send(ctx, rec, lang.CancelSuccessful(cmd))
	}
}
