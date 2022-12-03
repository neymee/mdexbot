package storage

import (
	"context"
	"time"

	"github.com/neymee/mdexbot/internal/database"
	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/log"
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

	topic := database.Topic{
		MangaID: sub.MangaID,
		Lang:    sub.Language,
		Title:   sub.MangaTitle,
	}

	err := r.db.FirstOrCreate(&topic, &topic).Error
	if err != nil {
		return err
	}

	topicSub := database.TopicSubscription{
		TopicID:   topic.ID,
		Recipient: user.Recipient(),
	}

	return r.db.Create(&topicSub).Error
}

func (r *Repo) SetSubscriptionLastUpdate(
	ctx context.Context,
	sub domain.Subscription,
	updatedAt time.Time,
	chapters ...domain.Chapter,
) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.SetSubscriptionLastUpdate").Trace().
			Dur("duration", time.Since(t)).
			Interface("subscription", sub).
			Time("updated_at", updatedAt).
			Interface("chapters", chapters).
			Send()
	}(time.Now())

	topic := database.Topic{}
	err := r.db.Model(&database.Topic{}).
		Find(&topic, "manga_id = ? AND lang = ?", sub.MangaID, sub.Language).
		Error
	if err != nil {
		return err
	}

	for _, c := range chapters {
		err := r.db.Create(&database.NotifiedChapter{
			TopicID: topic.ID,
			Chapter: c.Chapter,
			Volume:  c.Volume,
		}).Error
		if err != nil {
			return err
		}
	}

	return r.db.Model(&database.Topic{}).
		Where("id = ?", topic.ID).
		Update("updated_at", updatedAt).Error
}

func (r *Repo) IsChapterNotified(ctx context.Context, sub domain.Subscription, chapter domain.Chapter) (bool, error) {
	defer func(t time.Time) {
		log.Log(ctx, "storage.IsChapterNotified").Trace().
			Dur("duration", time.Since(t)).
			Interface("subscription", sub).
			Str("chapter", chapter.Chapter).
			Str("volume", chapter.Volume).
			Send()
	}(time.Now())

	topic := database.Topic{}
	err := r.db.Model(&database.Topic{}).
		Find(&topic, "manga_id = ? AND lang = ?", sub.MangaID, sub.Language).
		Error
	if err != nil {
		return false, err
	}

	res := r.db.First(
		&database.NotifiedChapter{},
		"topic_id = ? AND chapter = ? AND volume = ?",
		topic.ID, chapter.Chapter, chapter.Volume,
	)
	if res.Error != nil {
		return false, err
	}

	return res.RowsAffected > 0, nil
}

func (r *Repo) UserSubscriptions(ctx context.Context, recipient domain.Recipient) ([]domain.Subscription, error) {
	defer func(t time.Time) {
		log.Log(ctx, "storage.UserSubscriptions").Trace().
			Dur("duration", time.Since(t)).
			Str("recipient", recipient.Recipient()).
			Send()
	}(time.Now())

	var topics []database.Topic

	err := r.db.Joins(
		`JOIN topic_subscriptions ON topic_subscriptions.topic_id = topics.id
			AND topic_subscriptions.recipient = ?
			AND topic_subscriptions.deleted_at IS NULL`,
		recipient.Recipient(),
	).Find(&topics).Error
	if err != nil {
		return nil, err
	}

	subs := make([]domain.Subscription, 0, len(topics))
	for _, s := range topics {
		subs = append(subs, domain.Subscription{
			MangaID:    s.MangaID,
			MangaTitle: s.Title,
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

	topic := database.Topic{}
	err := r.db.Model(&database.Topic{}).
		Preload("Subscriptions").
		Find(&topic, "manga_id = ? AND lang = ?", mangaID, lang).Error
	if err != nil {
		return err
	}

	err = r.db.Delete(
		&database.TopicSubscription{},
		"recipient = ? AND topic_id = ?",
		recipient.Recipient(),
		topic.ID,
	).Error
	if err != nil {
		return err
	}

	if len(topic.Subscriptions) == 1 && topic.Subscriptions[0].Recipient == recipient.Recipient() {
		return r.db.Delete(&topic).Error
	}

	return nil
}

func (r *Repo) DeleteAllSubscriptions(ctx context.Context, recipient domain.Recipient) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.DeleteAllSubscriptions").Trace().
			Dur("duration", time.Since(t)).
			Str("recipient", recipient.Recipient()).
			Send()
	}(time.Now())

	return r.db.Delete(
		&database.TopicSubscription{},
		"recipient = ?",
		recipient.Recipient(),
	).Error
}

func (r *Repo) AllSubscriptions(ctx context.Context) ([]domain.SubscriptionExtended, error) {
	defer func(t time.Time) {
		log.Log(ctx, "storage.AllSubscriptions").Trace().
			Dur("duration", time.Since(t)).
			Send()
	}(time.Now())

	var topics []database.Topic
	err := r.db.Preload("Subscriptions").Find(&topics).Error
	if err != nil {
		return nil, err
	}

	result := make([]domain.SubscriptionExtended, 0, len(topics))
	for _, t := range topics {
		recs := make([]domain.Recipient, 0, len(t.Subscriptions))
		for _, s := range t.Subscriptions {
			recs = append(recs, domain.Recipient(s.Recipient))
		}

		result = append(result, domain.SubscriptionExtended{
			Subscription: domain.Subscription{
				MangaID:    t.MangaID,
				MangaTitle: t.Title,
				Language:   t.Lang,
			},
			UpdatedAt:  t.UpdatedAt,
			Recipients: recs,
		})
	}

	return result, nil
}
