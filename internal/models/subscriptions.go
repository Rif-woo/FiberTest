package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Plan       string    `gorm:"type:subscription_plan;default:'free';not null"`
	Status     string    `gorm:"type:subscription_Status;default:'active';not null"`
	ExpiresAt  time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
