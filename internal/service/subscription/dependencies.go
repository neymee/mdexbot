package subscription

import (
	"context"
	"time"

	"github.com/neymee/mdexbot/internal/domain"
)

type MangaDexAPI interface {
	Manga(ctx context.Context, id string) (domain.Manga, error)
	LastChapters(
		ctx context.Context,
		mangaID string,
		lang *string,
		publishedSince *time.Time,
	) ([]domain.Chapter, error)
}

type SubscriptionRepo interface {
	UserSubscriptions(ctx context.Context, recipient domain.Recipient) ([]domain.Subscription, error)
	SetUserSubscription(
		ctx context.Context,
		recipient domain.Recipient,
		sub domain.Subscription,
	) error
	DeleteUserSubscription(
		ctx context.Context,
		recipient domain.Recipient,
		mangaID string,
		lang string,
	) error
	AllSubscriptions(ctx context.Context) ([]domain.SubscriptionExtended, error)
	SetSubscriptionLastUpdate(
		ctx context.Context,
		sub domain.Subscription,
		updatedAt time.Time,
	) error
	DeleteAllSubscriptions(context.Context, domain.Recipient) error
}
