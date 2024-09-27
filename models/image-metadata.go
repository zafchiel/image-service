package models

import "gorm.io/gorm"

type ImageMetadata struct {
	gorm.Model
	Filename string `gorm:"unique;uniqueIndex;not null"`
	Format   string `gorm:"not null"`
	Size     int64  `gorm:"not null"`
	UserID   uint
}
