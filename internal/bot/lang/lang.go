package lang

import (
	"fmt"
	"html"
	"strings"

	"github.com/neymee/mdexbot/internal/domain"
)

const (
	errInternalError = "Error occured. Please try again."
	errWrongContext  = "You have command <b><i>%s</i></b> in progress. Please finish it or cancel with /cancel to start another command."

	start = "Use command /subscribe to subscribe on manga updates."

	cancelNoCommands = "There is no command to cancel."
	cancelSuccessful = "The command <b><i>%s</i></b> has been canceled."

	list       = "Titles you follow:"
	listNoSubs = "You don't have any active subscriptions. You can add subscription with /subscribe command."

	subscribeInit              = "Send me a link on a manga you want to track."
	subscribeChooseLanguage    = "<b><i>%s</i></b>\n\nChoose the language you want to track:"
	subscribeConfirmed          = "Great! You will receive a message when a new chapter of [%s] <b><i>%s</i></b> is published."
	subscribeAllreadyFollowing = "You're already following [%s] <b><i>%s</i></b>."

	subscribeErrInvalidLink = "Link \"%s\" is not recognized. Please send a valid link to a manga page on mangadex.org.\n\nFor example: https://mangadex.org/title/d8a959f7-648e-4c8d-8f23-f1f3f8e129f3/one-punch-man"

	unsubscribeNoSubs      = "You don't have any active subscriptions."
	unsubscribeChooseSub   = "Choose subscription you want to delete:"
	unsubscribeConfirmed   = "OK, you will not be longer notified about [%s] <b><i>%s</i></b> updates."
	unsubscribeNotFollowed = "You're already unsubscribed from [%s] <b><i>%s</i></b>. Call /unsubscribe again to get actual follows list."

	newChapterSingle = "[%s] <b><i>%s</i></b>\n\nNew chapter published: <b>%s</b>"
	newChapterMulti  = "[%s] <b><i>%s</i></b>\n\n%d new chapters published!"
)

func ErrInternalError() string {
	return errInternalError
}

func ErrWrongContext(cmd domain.Command) string {
	cmdEscaped := html.EscapeString(string(cmd))
	return fmt.Sprintf(errWrongContext, cmdEscaped)
}

func Start() string {
	return start
}

func CancelNoCommands() string {
	return cancelNoCommands
}
func CancelSuccessful(cmd domain.Command) string {
	cmdEscaped := html.EscapeString(string(cmd))
	return fmt.Sprintf(cancelSuccessful, cmdEscaped)
}

func List() string {
	return list
}

func ListNoSubs() string {
	return listNoSubs
}

func SubscribeInit() string {
	return subscribeInit
}

func SubscribeChooseLanguage(title string) string {
	titleEscaped := html.EscapeString(title)
	return fmt.Sprintf(subscribeChooseLanguage, titleEscaped)
}

func SubscribeConfirmed(title string, lang string) string {
	titleEscaped := html.EscapeString(title)
	langEscaped := html.EscapeString(lang)
	return fmt.Sprintf(subscribeConfirmed, langEscaped, titleEscaped)
}

func SubscribeAllreadyFollowing(title string, lang string) string {
	titleEscaped := html.EscapeString(title)
	langEscaped := html.EscapeString(lang)
	return fmt.Sprintf(subscribeAllreadyFollowing, langEscaped, titleEscaped)
}

func SubscribeErrInvalidLink(link string) string {
	linkEscaped := html.EscapeString(link)
	return fmt.Sprintf(subscribeErrInvalidLink, linkEscaped)
}

func UnsubscribeNoSubs() string {
	return unsubscribeNoSubs
}

func UnsubscribeChooseSub() string {
	return unsubscribeChooseSub
}

func UnsubscribeConfirmed(title string, lang string) string {
	titleEscaped := html.EscapeString(title)
	langEscaped := html.EscapeString(lang)
	return fmt.Sprintf(unsubscribeConfirmed, langEscaped, titleEscaped)
}

func UnsubscribeNotFollowed(title string, lang string) string {
	titleEscaped := html.EscapeString(title)
	langEscaped := html.EscapeString(lang)
	return fmt.Sprintf(unsubscribeNotFollowed, langEscaped, titleEscaped)
}

func NewChapterSingle(title, lang, chapterNum, chapterTitle, volumeNum string) string {
	chBuilder := strings.Builder{}
	if volumeNum != "" {
		chBuilder.WriteString(fmt.Sprintf("Vol. %s", volumeNum))
	}
	if chapterNum != "" {
		if chBuilder.Len() > 0 {
			chBuilder.WriteString(", ")
		}
		chBuilder.WriteString(fmt.Sprintf("Ch. %s", chapterNum))
	}
	if chapterTitle != "" {
		if chBuilder.Len() > 0 {
			chBuilder.WriteString(" - ")
		}
		chBuilder.WriteString(chapterTitle)
	}

	return fmt.Sprintf(
		newChapterSingle,
		html.EscapeString(lang),
		html.EscapeString(title),
		html.EscapeString(chBuilder.String()),
	)
}

func NewChapterMulti(title, lang string, chapterCount int) string {
	return fmt.Sprintf(
		newChapterMulti,
		html.EscapeString(lang),
		html.EscapeString(title),
		chapterCount,
	)
}
