package conversation

import (
	"context"

	"github.com/neymee/mdexbot/internal/domain"
)

type Service interface {
	ConversationContext(ctx context.Context, recipient domain.Recipient) (string, error)
	SetConversationContext(ctx context.Context, recipient domain.Recipient, cmd string) error
	DeleteConversationContext(ctx context.Context, recipient domain.Recipient) error
}

type service struct {
	repo ConversationRepo
}

func New(r ConversationRepo) Service {
	return &service{repo: r}
}

func (s *service) ConversationContext(ctx context.Context, recipient domain.Recipient) (string, error) {
	return s.repo.ConversationContext(ctx, recipient)
}

func (s *service) SetConversationContext(ctx context.Context, recipient domain.Recipient, cmd string) error {
	return s.repo.SetConversationContext(ctx, recipient, cmd)
}

func (s *service) DeleteConversationContext(ctx context.Context, recipient domain.Recipient) error {
	return s.repo.DeleteConversationContext(ctx, recipient)
}
