package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	VideoID    string    `gorm:"type:varchar(255);not null"`
	Platform   string    `gorm:"type:user_platform;default:'youtube';not null"`
	Content    string    `gorm:"type:text;not null"`
	Author     string    `gorm:"type:varchar(255);not null"`
	Date       time.Time `gorm:"type:timestamp;not null"`
	CreatedAt  time.Time
}
