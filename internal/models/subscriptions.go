package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Plan       string    `gorm:"type:enum('free', 'pro', 'business');default:'free';not null"`
	Status     string    `gorm:"type:enum('active', 'cancelled');default:'active';not null"`
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
