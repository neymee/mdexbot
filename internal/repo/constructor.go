package repo

import (
	"github.com/neymee/mdexbot/internal/repo/mdex"
	"github.com/neymee/mdexbot/internal/repo/storage"
	"gorm.io/gorm"
)

type Repos struct {
	MDex    *mdex.Repo
	Storage *storage.Repo
}

func New(
	db *gorm.DB,
) *Repos {
	return &Repos{
		MDex:    mdex.New(),
		Storage: storage.New(db),
	}
}
