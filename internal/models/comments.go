package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Platform   string    `gorm:"type:enum('instagram', 'twitter', 'youtube');not null"`
	Content    string    `gorm:"type:text;not null"`
	Sentiment  string    `gorm:"type:enum('positive', 'neutral', 'negative');not null"`
	IsQuestion bool      `gorm:"type:boolean;default:false"`
	CreatedAt  time.Time
}
