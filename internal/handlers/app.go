package handlers

import (
	"github.com/zafchiel/image-service/internal/config"
	"github.com/zafchiel/image-service/internal/storage"
	"gorm.io/gorm"
)

type App struct {
	DB      *gorm.DB
	Config  *config.Config
	Storage storage.Storage
}
