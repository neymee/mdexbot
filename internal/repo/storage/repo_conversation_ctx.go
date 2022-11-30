package storage

import (
	"context"
	"time"

	"github.com/neymee/mdexbot/internal/database"
	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/log"
	"gorm.io/gorm/clause"
)

func (r *Repo) ConversationContext(ctx context.Context, recipient domain.Recipient) (domain.Command, error) {
	defer func(t time.Time) {
		log.Log(ctx, "storage.ConversationContext").Trace().
			Dur("duration", time.Since(t)).
			Str("recipient", recipient.Recipient()).
			Send()
	}(time.Now())

	var convCtx *database.ConversationContext
	res := r.db.Limit(1).Find(&convCtx, "recipient = ?", recipient)
	if res.RowsAffected == 0 {
		return "", nil
	} else if res.Error != nil {
		return "", res.Error
	}
	return domain.Command(convCtx.Command), nil
}

func (r *Repo) SetConversationContext(ctx context.Context, recipient domain.Recipient, cmd domain.Command) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.SetConversationContext").Trace().
			Dur("duration", time.Since(t)).
			Interface("conversation_context", cmd).
			Send()
	}(time.Now())

	db := r.db.Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "recipient"}},
			DoUpdates: clause.Assignments(
				map[string]interface{}{
					"command": cmd,
				},
			),
		},
	).Create(&database.ConversationContext{
		Recipient: recipient.Recipient(),
		Command:   cmd.String(),
	})

	return db.Error
}

func (r *Repo) DeleteConversationContext(ctx context.Context, recipient domain.Recipient) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.DeleteConversationContext").Trace().
			Dur("duration", time.Since(t)).
			Str("recipient", recipient.Recipient()).
			Send()
	}(time.Now())

	return r.db.Delete(&database.ConversationContext{}, "recipient = ?", recipient).Error
}
