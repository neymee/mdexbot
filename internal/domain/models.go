package domain

import (
	"fmt"
	"strconv"
	"time"
)

type Manga struct {
	ID                   string
	Title                map[string]string // lang : title
	TranslationLanguages []string
}

func (m *Manga) GetTitle() string {
	title, ok := m.Title["en"]
	if !ok {
		for _, v := range m.Title {
			title = v
			break
		}
	}
	return title
}

type Chapter struct {
	ID          string
	Title       string
	Volume      string
	Chapter     string
	ExternalUrl string
	PublishedAt time.Time
}

type Recipient string

func RecipientFromInt64(r int64) Recipient {
	return Recipient(fmt.Sprint(r))
}

func (r Recipient) Recipient() string {
	return string(r)
}

func (r Recipient) AsInt64() int64 {
	i, err := strconv.ParseInt(string(r), 10, 64)
	if err != nil {
		i = -1
	}
	return i
}

type Update struct {
	MangaTitle  string
	MangaID     string
	Language    string
	Recipients  []Recipient
	NewChapters []Chapter
}

type Subscription struct {
	MangaID    string
	MangaTitle string
	Language   string
}

type SubscriptionExtended struct {
	Subscription
	UpdatedAt  time.Time
	Recipients []Recipient
}
