package database

import (
	"context"
	"fmt"
	"time"

	"github.com/neymee/mdexbot/internal/config"
	"github.com/neymee/mdexbot/internal/log"
	"github.com/neymee/mdexbot/internal/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(ctx context.Context, cfg *config.Config) (*gorm.DB, error) {
	conn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.DB.Host,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.Port,
		cfg.DB.SLL,
	)

	var (
		db  *gorm.DB
		err error
	)

	for i := 0; i < 3; i++ {
		if utils.IsContextDone(ctx) {
			return nil, fmt.Errorf("interrupted: context is cancelled")
		}

		db, err = gorm.Open(postgres.Open(conn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			log.Log(ctx, "NewDB").Warn().Msg("Database is unavailable. Trying to reconnect...")
			time.Sleep(time.Second * 3)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("database is unavailable: %w", err)
	}

	err = db.AutoMigrate(
		&Subscription{},
		&ConversationContext{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
