package conversation

import (
	"context"

	"github.com/neymee/mdexbot/internal/domain"
)

type ConversationRepo interface {
	ConversationContext(ctx context.Context, recipient domain.Recipient) (string, error)
	SetConversationContext(ctx context.Context, recipient domain.Recipient, cmd string) error
	DeleteConversationContext(ctx context.Context, recipient domain.Recipient) error
}
