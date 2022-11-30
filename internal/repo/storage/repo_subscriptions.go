package storage

import (
	"context"
	"time"

	"github.com/neymee/mdexbot/internal/database"
	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/log"
	"gorm.io/gorm/clause"
)

func (r *Repo) SetUserSubscription(
	ctx context.Context,
	user domain.Recipient,
	sub domain.Subscription,
) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.SetUserSubscription").Trace().
			Dur("duration", time.Since(t)).
			Interface("subscription", sub).
			Send()
	}(time.Now())

	dbSub := &database.Subscription{
		MangaID:    sub.MangaID,
		MangaTitle: sub.MangaTitle,
		Lang:       sub.Language,
		Recipient:  user.Recipient(),
	}

	db := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(dbSub)
	return db.Error
}

func (r *Repo) SetSubscriptionLastUpdate(
	ctx context.Context,
	sub domain.SubscriptionExtended,
	updatedAt time.Time,
) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.SetSubscriptionLastUpdate").Trace().
			Dur("duration", time.Since(t)).
			Interface("subscription", sub).
			Time("updated_at", updatedAt).
			Send()
	}(time.Now())

	for _, rec := range sub.Recipients {
		dbSub := &database.Subscription{
			MangaID:    sub.MangaID,
			MangaTitle: sub.MangaTitle,
			Lang:       sub.Language,
			Recipient:  rec.Recipient(),
		}

		err := r.db.Model(dbSub).Update("updated_at", updatedAt).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repo) UserSubscriptions(ctx context.Context, recipient domain.Recipient) ([]domain.Subscription, error) {
	defer func(t time.Time) {
		log.Log(ctx, "storage.UserSubscriptions").Trace().
			Dur("duration", time.Since(t)).
			Str("recipient", recipient.Recipient()).
			Send()
	}(time.Now())

	var dbSubs []database.Subscription
	err := r.db.Find(&dbSubs, "recipient = ?", recipient).Error
	if err != nil {
		return nil, err
	}

	subs := make([]domain.Subscription, 0, len(dbSubs))
	for _, s := range dbSubs {
		subs = append(subs, domain.Subscription{
			MangaID:    s.MangaID,
			MangaTitle: s.MangaTitle,
			Language:   s.Lang,
		})
	}

	return subs, nil
}

func (r *Repo) DeleteUserSubscription(
	ctx context.Context,
	recipient domain.Recipient,
	mangaID string,
	lang string,
) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.DeleteUserSubscription").Trace().
			Dur("duration", time.Since(t)).
			Str("recipient", recipient.Recipient()).
			Str("lang", lang).
			Send()
	}(time.Now())

	db := r.db.Delete(
		&database.Subscription{},
		"recipient = ? and manga_id = ? and lang = ?",
		recipient,
		mangaID,
		lang,
	)
	return db.Error
}

func (r *Repo) DeleteAllSubscriptions(ctx context.Context, recipient domain.Recipient) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.DeleteAllSubscriptions").Trace().
			Dur("duration", time.Since(t)).
			Str("recipient", recipient.Recipient()).
			Send()
	}(time.Now())

	db := r.db.Delete(
		&database.Subscription{},
		"recipient = ?",
		recipient,
	)

	return db.Error
}

func (r *Repo) AllSubscriptions(ctx context.Context) ([]domain.SubscriptionExtended, error) {
	defer func(t time.Time) {
		log.Log(ctx, "storage.AllSubscriptions").Trace().
			Dur("duration", time.Since(t)).
			Send()
	}(time.Now())

	var dbSubs []database.Subscription
	err := r.db.Find(&dbSubs).Error
	if err != nil {
		return nil, err
	}

	// grouping by id+lang is temporary solution until DB refactored
	type subKey struct {
		id, lang string
	}

	subsByKey := make(map[subKey]*domain.SubscriptionExtended)

	for _, s := range dbSubs {
		key := subKey{s.MangaID, s.Lang}
		sub := subsByKey[key]
		if sub == nil {
			sub = &domain.SubscriptionExtended{
				Subscription: domain.Subscription{
					MangaID:    s.MangaID,
					MangaTitle: s.MangaTitle,
					Language:   s.Lang,
				},
				UpdatedAt: s.UpdatedAt,
			}
			subsByKey[key] = sub
		}

		sub.Recipients = append(sub.Recipients, domain.Recipient(s.Recipient))
	}

	result := make([]domain.SubscriptionExtended, 0, len(subsByKey))
	for _, sub := range subsByKey {
		result = append(result, *sub)
	}

	return result, nil
}
