package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/neymee/mdexbot/internal/database"
	"github.com/neymee/mdexbot/internal/domain"
	"github.com/neymee/mdexbot/internal/errors"
	"github.com/neymee/mdexbot/internal/log"
	"gorm.io/gorm/clause"
)

func (r *Repo) ConversationContext(ctx context.Context, recipient domain.Recipient) (string, error) {
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
		return "", fmt.Errorf("%w: %w", errors.DatabaseError, res.Error)
	}
	return convCtx.Command, nil
}

func (r *Repo) SetConversationContext(ctx context.Context, recipient domain.Recipient, cmd string) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.SetConversationContext").Trace().
			Dur("duration", time.Since(t)).
			Interface("conversation_context", cmd).
			Send()
	}(time.Now())

	err := r.db.Clauses(
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
		Command:   cmd,
	}).Error

	if err != nil {
		return fmt.Errorf("%w: %w", errors.DatabaseError, err)
	}
	return nil
}

func (r *Repo) DeleteConversationContext(ctx context.Context, recipient domain.Recipient) error {
	defer func(t time.Time) {
		log.Log(ctx, "storage.DeleteConversationContext").Trace().
			Dur("duration", time.Since(t)).
			Str("recipient", recipient.Recipient()).
			Send()
	}(time.Now())

	err := r.db.Delete(&database.ConversationContext{}, "recipient = ?", recipient).Error
	if err != nil {
		return fmt.Errorf("%w: %w", errors.DatabaseError, err)
	}
	return nil
}
