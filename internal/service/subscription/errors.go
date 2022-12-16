package subscription

import "fmt"

var (
	ErrNoSuchSubscription = fmt.Errorf("no such subscription")
	ErrMangaNotFound      = fmt.Errorf("manga not found")
)

type AlreadySubscribedError struct {
	Manga string
	Lang  string
}

func (e *AlreadySubscribedError) Error() string {
	return fmt.Sprintf("already subscribed to [%s] %s", e.Lang, e.Manga)
}
