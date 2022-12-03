package database

import (
	"time"

	"gorm.io/gorm"
)

type ConversationContext struct {
	Recipient string `gorm:"primarykey"`
	Command   string
	CreatedAt time.Time
}

type Topic struct {
	gorm.Model
	MangaID          string `gorm:"uniqueIndex:idx_topic_manga_id_lang,where:deleted_at IS NULL"`
	Lang             string `gorm:"uniqueIndex:idx_topic_manga_id_lang,where:deleted_at IS NULL"`
	Title            string
	Subscriptions    []TopicSubscription
	NotifiedChapters []NotifiedChapter
}

type TopicSubscription struct {
	gorm.Model
	TopicID   uint   `gorm:"uniqueIndex:idx_topic_subscription_topic_id_recipient,where:deleted_at IS NULL"`
	Recipient string `gorm:"uniqueIndex:idx_topic_subscription_topic_id_recipient,where:deleted_at IS NULL"`
}

type NotifiedChapter struct {
	gorm.Model
	TopicID uint   `gorm:"uniqueIndex:idx_notified_chapters_composite,where:deleted_at IS NULL"`
	Chapter string `gorm:"uniqueIndex:idx_notified_chapters_composite,where:deleted_at IS NULL"`
	Volume  string `gorm:"uniqueIndex:idx_notified_chapters_composite,where:deleted_at IS NULL"`
}
