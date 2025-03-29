package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Type      string    `gorm:"type:enum('analysis', 'payment', 'alert');not null"`
	Message   string    `gorm:"type:text;not null"`
	IsRead    bool      `gorm:"type:boolean;default:false"`
	CreatedAt time.Time
}
