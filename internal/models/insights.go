package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Insight struct {
	ID               uuid.UUID       `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID           uuid.UUID       `gorm:"type:uuid;not null;index"`
	TopComments      datatypes.JSON  `gorm:"type:jsonb"`
	NegativeComments datatypes.JSON	 `gorm:"type:jsonb"`
	QuestionComments datatypes.JSON	 `gorm:"type:jsonb"`
	FeedbackComments datatypes.JSON	 `gorm:"type:jsonb"`
	CreatedAt        time.Time
}
