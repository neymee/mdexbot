package service

import (
	"github.com/neymee/mdexbot/internal/service/conversation"
	"github.com/neymee/mdexbot/internal/service/subscription"
)

type Services struct {
	Subscription subscription.Service
	Conversation conversation.Service
}

func New(
	mdexAPI subscription.MangaDexAPI,
	subRepo subscription.SubscriptionRepo,
	convRepo conversation.ConversationRepo,
) *Services {
	return &Services{
		Subscription: subscription.New(mdexAPI, subRepo),
		Conversation: conversation.New(convRepo),
	}
}
