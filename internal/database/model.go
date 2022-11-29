package database

import "time"

type ConversationContext struct {
	Recipient string `gorm:"primarykey"`
	Command   string
	CreatedAt time.Time
}

type Subscription struct {
	Lang       string `gorm:"primarykey"`
	Recipient  string `gorm:"primarykey"`
	MangaID    string `gorm:"primarykey"`
	MangaTitle string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
