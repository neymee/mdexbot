package mdex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/errors"
	"github.com/neymee/mdexbot/internal/log"
	"github.com/neymee/mdexbot/internal/metrics"
	"github.com/neymee/mdexbot/internal/service/subscription"
)

const (
	urlBase         = "https://api.mangadex.org"
	apiGetManga     = "/manga/%s"
	apiGetMangaFeed = "/manga/%s/feed"
)

func urlGetManga(id string) string {
	return urlBase + fmt.Sprintf(apiGetManga, id)
}

func urlGetMangaFeed(id string) string {
	return urlBase + fmt.Sprintf(apiGetMangaFeed, id)
}

type Repo struct{}

var _ subscription.MangaDexAPI = (*Repo)(nil)

func New() *Repo {
	return &Repo{}
}

func (r *Repo) Manga(ctx context.Context, id string) (domain.Manga, error) {
	defer func(start time.Time) {
		duration := time.Since(start)
		metrics.HTTPDuration(fmt.Sprintf(apiGetManga, "*")).Observe(duration.Seconds())
		log.Log(ctx, "mdex.Manga").Trace().
			Dur("duration", duration).
			Str("id", id).
			Send()
	}(time.Now())

	var result domain.Manga

	u := urlGetManga(id)
	resp, err := http.Get(u)
	if err != nil {
		return result, errors.FailedHTTPReqError{Err: err}
	}

	if resp.StatusCode == 404 {
		return result, subscription.ErrMangaNotFound
	} else if resp.StatusCode != 200 {
		return result, errors.FailedHTTPReqError{Err: fmt.Errorf("request failed with status %d", resp.StatusCode)}
	}

	var manga *apiResponse[apiManga]
	if err := json.NewDecoder(resp.Body).Decode(&manga); err != nil {
		return result, err
	}

	if err := manga.Validate(); err != nil {
		return result, err
	}

	result.ID = manga.Data.ID
	result.Title = manga.Data.Attributes.Title
	result.TranslationLanguages = manga.Data.Attributes.AvailableLanguages

	return result, nil
}

func (r *Repo) LastChapters(
	ctx context.Context,
	mangaID string,
	lang *string,
	publishedSince *time.Time,
) ([]domain.Chapter, error) {
	defer func(start time.Time) {
		duration := time.Since(start)
		metrics.HTTPDuration(fmt.Sprintf(apiGetMangaFeed, "*")).Observe(duration.Seconds())
		log.Log(ctx, "mdex.LastChapters").Trace().
			Dur("duration", duration).
			Str("manga_id", mangaID).
			Interface("lang", lang).
			Interface("published_since", publishedSince).
			Send()
	}(time.Now())

	u, err := url.Parse(urlGetMangaFeed(mangaID))
	if err != nil {
		return nil, err
	}

	qry := url.Values{
		"limit":                []string{"20"},
		"contentRating[]":      []string{"safe", "suggestive", "erotica", "pornographic"},
		"includeFutureUpdates": []string{"1"},
		"order[publishAt]":     []string{"asc"},
	}
	if lang != nil {
		qry.Add("translatedLanguage[]", *lang)
	}
	if publishedSince != nil && (*publishedSince != time.Time{}) {
		qry.Add("publishAtSince", publishedSince.UTC().Format("2006-01-02T15:04:05"))
	}
	u.RawQuery = qry.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, errors.FailedHTTPReqError{Err: err}
	}

	if resp.StatusCode == 404 {
		return nil, subscription.ErrMangaNotFound
	} else if resp.StatusCode != 200 {
		return nil, errors.FailedHTTPReqError{Err: fmt.Errorf("request failed with status %d", resp.StatusCode)}
	}

	var feeds *apiResponse[[]apiMangaFeedItem]

	if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
		return nil, err
	}

	if err := feeds.Validate(); err != nil {
		return nil, err
	}

	chapters := []apiMangaFeedItem{}
	for _, f := range *feeds.Data {
		if f.Type == "chapter" {
			chapters = append(chapters, f)
		}
	}

	if len(chapters) == 0 {
		// TODO: and total > limit + offset -> request the rest?
		return nil, nil
	}

	result := make([]domain.Chapter, 0, len(chapters))
	for _, ch := range chapters {
		r := domain.Chapter{
			ID:          ch.ID,
			Title:       ch.Attributes.Title,
			Volume:      ch.Attributes.Volume,
			Chapter:     ch.Attributes.Chapter,
			ExternalUrl: ch.Attributes.ExternalUrl,
			PublishedAt: ch.Attributes.PublishedAt,
		}
		result = append(result, r)
	}

	return result, nil
}
