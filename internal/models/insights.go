package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Insight struct {
	ID                uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID            uuid.UUID      `gorm:"type:uuid;not null;index"`
	VideoID           string         `gorm:"not null;index"` // pour retrouver les insights par vidéo
	Sentiment         string         // Ex: "Négatif/Neutre"
	Summary           string         // Résumé du ton général
	TopComments       datatypes.JSON // []string
	NegativeComments  datatypes.JSON // []string
	QuestionComments  datatypes.JSON // []string
	FeedbackComments  datatypes.JSON // []string ou autres remarques
	Keywords          datatypes.JSON // []string
	CreatedAt         time.Time
}
