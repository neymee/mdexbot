package bot

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/neymee/mdexbot/internal/bot/lang"
	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/log"
	"gopkg.in/telebot.v3"
)

const (
	MangaDexDomain string = "mangadex.org"
	MangaDexURL    string = "https://" + MangaDexDomain
)

var ErrInvalidLink = errors.New("invalid link")

func setupReqCtx(c telebot.Context) {
	traceID := time.Now().UnixMicro()
	c.Set(string(log.KeyTraceID), &traceID)
}

func reqCtx(c telebot.Context) context.Context {
	ctx := context.Background()
	traceID, ok := c.Get(string(log.KeyTraceID)).(*int64)
	if !ok {
		return ctx
	}
	ctx = context.WithValue(ctx, log.KeyTraceID, traceID)
	return ctx
}

// mangaIDFromURL extracts manga id from url
func mangaIDFromURL(link string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(link))
	if err != nil {
		return "", ErrInvalidLink
	}

	if u.Host != MangaDexDomain {
		return "", ErrInvalidLink
	}

	splitted := strings.Split(u.Path, "/")
	if len(splitted) < 3 || splitted[1] != "title" {
		return "", ErrInvalidLink
	}

	id := splitted[2]
	isUUID, err := regexp.Match("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$", []byte(id))
	if err != nil {
		return "", ErrInvalidLink
	}
	if !isUUID {
		return "", ErrInvalidLink
	}

	return id, nil
}

func buildLanguageButtons(manga domain.Manga) [][]telebot.InlineButton {
	langButtons := [][]telebot.InlineButton{
		{
			{
				Text:   "Any",
				Data:   formatButtonData(manga.ID, "any"),
				Unique: CmdSubscribeBtn.String(),
			},
		},
	}
	langBtnsRow := []telebot.InlineButton{}
	var langIdx int
	for _, translationLang := range manga.TranslationLanguages {
		if translationLang == "" {
			continue
		}

		text := translationLang

		flag, ok := lang.GetFlag(translationLang)
		if ok {
			text += " " + flag
		}

		langBtnsRow = append(
			langBtnsRow,
			telebot.InlineButton{
				Text:   text,
				Data:   formatButtonData(manga.ID, translationLang),
				Unique: CmdSubscribeBtn.String(),
			},
		)

		if langIdx += 1; langIdx%3 == 0 || langIdx == len(manga.TranslationLanguages) {
			langButtons = append(langButtons, langBtnsRow)
			langBtnsRow = []telebot.InlineButton{}
		}
	}

	return langButtons
}

func formatButtonData(mangaID string, lang string) string {
	return fmt.Sprintf("%s/%s", mangaID, lang)
}

func parseButtonData(data string) (string, string, error) {
	splitted := strings.SplitN(data, "/", 2)
	if len(splitted) != 2 || splitted[0] == "" || splitted[1] == "" {
		return "", "", fmt.Errorf("invalid button data: \"%s\"", data)
	}
	return splitted[0], splitted[1], nil
}
