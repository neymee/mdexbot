package bot

import (
	"errors"
	"fmt"
	"time"

	"github.com/neymee/mdexbot/internal/bot/lang"
	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/log"
	"github.com/neymee/mdexbot/internal/service"
	"github.com/neymee/mdexbot/internal/service/subscription"
	"gopkg.in/telebot.v3"
)

type InternalError error

func initHandlers(bot *telebot.Bot, s *service.Services) {
	bot.Handle(CmdStart.Endpoint(), onStart(s), middlewares(CmdStart)...)

	bot.Handle(CmdText.Endpoint(), onText(s), middlewares(CmdText)...)
	bot.Handle(CmdSubscribe.Endpoint(), onSubscribe(s), middlewares(CmdSubscribe)...)
	bot.Handle(CmdSubscribeBtn.Endpoint(), onSubscribeBtn(s), middlewares(CmdSubscribeBtn)...)

	bot.Handle(CmdUnsubscribe.Endpoint(), onUnsubscribe(s), middlewares(CmdUnsubscribe)...)
	bot.Handle(CmdUnsubscribeBtn.Endpoint(), onUnsubscribeBtn(s), middlewares(CmdUnsubscribeBtn)...)

	bot.Handle(CmdList.Endpoint(), onList(s), middlewares(CmdList)...)
	bot.Handle(CmdCancel.Endpoint(), onCancel(s), middlewares(CmdCancel)...)

	bot.OnError = func(err error, c telebot.Context) {
		log.Error(reqCtx(c), "bot.OnError", err).
			Int64("recipient", c.Chat().ID).
			Int("message_id", c.Message().ID).
			Msg("Error during processing request")

		if _, ok := err.(InternalError); ok {
			send(reqCtx(c), domain.RecipientFromInt64(c.Chat().ID), lang.ErrInternalError())
		}
	}
}

func middlewares(method Command) []telebot.MiddlewareFunc {
	return []telebot.MiddlewareFunc{
		func(next telebot.HandlerFunc) telebot.HandlerFunc {
			// setup context
			return func(c telebot.Context) error {
				setupReqCtx(c)
				return next(c)
			}
		},
		func(next telebot.HandlerFunc) telebot.HandlerFunc {
			// recover panic
			return func(c telebot.Context) error {
				defer func() {
					if err := recover(); err != nil {
						log.Log(reqCtx(c), method.String()).Error().
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
				log.Log(reqCtx(c), method.String()).Trace().
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
					log.Log(reqCtx(c), method.String()).Trace().
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
			reqCtx(c),
			domain.RecipientFromInt64(c.Chat().ID),
			lang.Start(),
		)
	}
}

func onText(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := reqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		cmd, err := s.Conversation.ConversationContext(ctx, rec)
		if err != nil {
			return InternalError(err)
		}

		if cmd != CmdSubscribe.String() {
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
		ctx := reqCtx(c)
		rec := domain.RecipientFromInt64(c.Chat().ID)

		err := s.Conversation.SetConversationContext(ctx, rec, CmdSubscribe.String())
		if err != nil {
			return InternalError(err)
		}

		return send(ctx, rec, lang.SubscribeInit())
	}
}

func onSubscribeBtn(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := reqCtx(c)
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
		ctx := reqCtx(c)
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
					Unique: CmdUnsubscribeBtn.String(),
				},
			})
		}

		return send(ctx, rec, lang.UnsubscribeChooseSub(), withKeyboard(keyboard))
	}
}

func onUnsubscribeBtn(s *service.Services) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		ctx := reqCtx(c)
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
		ctx := reqCtx(c)
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
		ctx := reqCtx(c)
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
