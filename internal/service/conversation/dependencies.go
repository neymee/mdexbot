package conversation

import (
	"context"

	"github.com/neymee/mdexbot/internal/domain"
)

type ConversationRepo interface {
	ConversationContext(ctx context.Context, recipient domain.Recipient) (domain.Command, error)
	SetConversationContext(ctx context.Context, recipient domain.Recipient, cmd domain.Command) error
	DeleteConversationContext(ctx context.Context, recipient domain.Recipient) error
}
