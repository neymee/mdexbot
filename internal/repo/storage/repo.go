package storage

import (
	"github.com/neymee/mdexbot/internal/service/conversation"
	"github.com/neymee/mdexbot/internal/service/subscription"
	"gorm.io/gorm"
)

type Repo struct {
	db *gorm.DB
}

var _ subscription.SubscriptionRepo = (*Repo)(nil)
var _ conversation.ConversationRepo = (*Repo)(nil)

func New(
	db *gorm.DB,
) *Repo {
	return &Repo{
		db: db,
	}
}
