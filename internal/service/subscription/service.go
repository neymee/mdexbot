package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/neymee/mdexbot/internal/domain"
)

type Service interface {
	Manga(ctx context.Context, mangaID string) (domain.Manga, error)
	List(ctx context.Context, rec domain.Recipient) ([]domain.Subscription, error)
	Subscribe(ctx context.Context, user domain.Recipient, mangaID string, lang string) (domain.Subscription, error)
	Unsubscribe(ctx context.Context, user domain.Recipient, mangaID string, lang string) (domain.Subscription, error)
	UnsubscribeAll(ctx context.Context, user domain.Recipient) error
	Updates(ctx context.Context) ([]domain.Update, error)
}

type service struct {
	mdex    MangaDexAPI
	storage SubscriptionRepo
}

func New(
	mdex MangaDexAPI,
	storage SubscriptionRepo,
) Service {
	return &service{
		mdex:    mdex,
		storage: storage,
	}
}

func (s *service) Manga(ctx context.Context, mangaID string) (domain.Manga, error) {
	return s.mdex.Manga(ctx, mangaID)
}

func (s *service) List(ctx context.Context, rec domain.Recipient) ([]domain.Subscription, error) {
	return s.storage.UserSubscriptions(ctx, rec)
}

func (s *service) Subscribe(ctx context.Context, user domain.Recipient, mangaID string, lang string) (domain.Subscription, error) {
	allSubs, err := s.storage.UserSubscriptions(ctx, user)
	if err != nil {
		return domain.Subscription{}, err
	}

	// find all subs to this manga, check if already subscribed
	var mangaSubs []domain.Subscription
	for _, sub := range allSubs {
		if sub.MangaID == mangaID {
			if sub.Language == lang {
				return domain.Subscription{}, &AlreadySubscribedError{Manga: sub.MangaTitle, Lang: sub.Language}
			}
			mangaSubs = append(mangaSubs, sub)
		}
	}

	// remove sub to "any" if lang != "any" OR remove all subs if lang == "any"
	if lang == "any" && len(mangaSubs) > 0 || len(mangaSubs) == 1 && mangaSubs[0].Language == "any" {
		for _, sub := range mangaSubs {
			if sub.MangaID == mangaID {
				err := s.storage.DeleteUserSubscription(ctx, user, sub.MangaID, sub.Language)
				if err != nil {
					return domain.Subscription{}, err
				}
			}
		}
	}

	manga, err := s.mdex.Manga(ctx, mangaID)
	if err != nil {
		return domain.Subscription{}, err
	}

	sub := domain.Subscription{
		MangaID:    mangaID,
		MangaTitle: manga.GetTitle(),
		Language:   lang,
	}

	err = s.storage.SetUserSubscription(ctx, user, sub)
	if err != nil {
		return domain.Subscription{}, err
	}

	return sub, nil
}

func (s *service) Unsubscribe(ctx context.Context, user domain.Recipient, mangaID string, lang string) (domain.Subscription, error) {
	currentSubs, err := s.storage.UserSubscriptions(ctx, user)
	if err != nil {
		return domain.Subscription{}, err
	}

	var deletedSub *domain.Subscription
	for _, sub := range currentSubs {
		if sub.MangaID == mangaID && sub.Language == lang {
			deletedSub = &sub
			break
		}
	}

	if deletedSub == nil {
		return domain.Subscription{}, ErrNoSuchSubscription
	}

	err = s.storage.DeleteUserSubscription(ctx, user, mangaID, lang)
	if err != nil {
		return domain.Subscription{}, err
	}

	return *deletedSub, nil
}

func (s *service) UnsubscribeAll(ctx context.Context, user domain.Recipient) error {
	return s.storage.DeleteAllSubscriptions(ctx, user)
}

func (s *service) Updates(ctx context.Context) ([]domain.Update, error) {
	subs, err := s.storage.AllSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	var updates []domain.Update
	for _, sub := range subs {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("interrupted: context if cancelled")
		default:
		}

		chapters, err := s.newChapters(ctx, sub)
		if err != nil {
			return nil, err
		}

		err = s.storage.SetSubscriptionLastUpdate(ctx, sub.Subscription, time.Now().UTC(), chapters...)
		if err != nil {
			return nil, err
		}

		chaptersLen := len(chapters)
		if chaptersLen == 0 {
			continue
		}

		updates = append(
			updates,
			domain.Update{
				MangaTitle:  sub.MangaTitle,
				MangaID:     sub.MangaID,
				Language:    sub.Language,
				NewChapters: chapters,
				Recipients:  sub.Recipients,
			},
		)
	}

	return updates, nil
}

// newChapters returns chapters published since last subscription update.
// Chapters that have already been notified will be filtered.
func (s *service) newChapters(ctx context.Context, sub domain.SubscriptionExtended) ([]domain.Chapter, error) {
	var lang *string
	if sub.Language != "any" {
		lang = &sub.Language
	}

	lastChapters, err := s.mdex.LastChapters(ctx, sub.MangaID, lang, &sub.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// filter chapters that have already been notified
	type key struct{ vol, ch string }
	chaptersAdded := map[key]struct{}{}
	chapters := []domain.Chapter{}
	for _, ch := range lastChapters {
		isNotified, err := s.storage.IsChapterNotified(ctx, sub.Subscription, ch)
		if err != nil {
			return nil, err
		}

		key := key{vol: ch.Volume, ch: ch.Chapter}
		_, isAdded := chaptersAdded[key]

		if !isNotified && !isAdded {
			chapters = append(chapters, ch)
			chaptersAdded[key] = struct{}{}
		}
	}
	return chapters, nil
}
