package mdex

import (
	"fmt"
	"time"
)

type apiResponse[T any] struct {
	Result   string     `json:"result"`
	Response string     `json:"response"`
	Data     *T         `json:"data,omitempty"`
	Errors   []apiError `json:"errors,omitempty"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
	Total    int        `json:"total"`
}

func (r *apiResponse[T]) Validate() error {
	if len(r.Errors) > 0 {
		return &r.Errors[0]
	}
	if r.Data == nil {
		return fmt.Errorf("request failed with empty data")
	}
	return nil
}

type apiError struct {
	ID      string `json:"id"`
	Status  int    `json:"status"`
	Title   string `json:"title"`
	Details string `json:"details"`
}

func (e *apiError) Error() string {
	return fmt.Sprintf(
		"%s - failed with status %d [%s]: %s",
		e.ID,
		e.Status,
		e.Title,
		e.Details,
	)
}

// Manga
type apiManga struct {
	ID         string        `json:"id"`
	Attributes apiMangaAttrs `json:"attributes"`
}

type apiMangaAttrs struct {
	Title              map[string]string `json:"title"`
	AvailableLanguages []string          `json:"availableTranslatedLanguages"`
}

// Feed
type apiMangaFeedItem struct {
	ID         string                `json:"id"`
	Type       string                `json:"type"`
	Attributes apiMangeFeedItemAttrs `json:"attributes"`
}

type apiMangeFeedItemAttrs struct {
	Volume      string    `json:"volume"`
	Chapter     string    `json:"chapter"`
	Title       string    `json:"title"`
	Language    string    `json:"translatedLanguage"`
	ExternalUrl string    `json:"externalUrl"`
	PublishedAt time.Time `json:"publishAt"`
}
