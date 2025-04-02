// internal/repositories/insight_repository.go
package repositories

import (
	"context" // Bonne pratique d'utiliser le contexte
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

		"github.com/Azertdev/FiberTest/internal/models"
 // Adaptez le chemin d'import
)

// InsightRepository définit les opérations pour les Insights
type InsightRepository interface {
	CreateInsight(ctx context.Context, insight *models.Insight) error
	GetInsightByVideoID(ctx context.Context, userID uuid.UUID, videoID string) (*models.Insight, error)
	// Ajoutez d'autres méthodes si nécessaire (Update, Delete, List...)
}

type insightRepository struct {
	db *gorm.DB
}

// NewInsightRepository crée une nouvelle instance de InsightRepository
func NewInsightRepository(db *gorm.DB) InsightRepository {
	return &insightRepository{db: db}
}

// CreateInsight enregistre un nouvel Insight en base de données
func (r *insightRepository) CreateInsight(ctx context.Context, insight *models.Insight) error {
	// GORM gère automatiquement les champs ID (avec default) et CreatedAt
	result := r.db.WithContext(ctx).Create(insight)
	if result.Error != nil {
		// TODO: Gérer les erreurs spécifiques (ex: violation de contrainte unique si vous en ajoutez)
		return fmt.Errorf("échec de la création de l'insight: %w", result.Error)
	}
	return nil
}

// GetInsightByVideoID récupère un Insight par UserID et VideoID
func (r *insightRepository) GetInsightByVideoID(ctx context.Context, userID uuid.UUID, videoID string) (*models.Insight, error) {
	var insight models.Insight
	result := r.db.WithContext(ctx).Where("user_id = ? AND video_id = ?", userID, videoID).First(&insight)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Retourne nil, nil si non trouvé (choix de conception, on pourrait retourner une erreur spécifique)
		}
		return nil, fmt.Errorf("échec de la récupération de l'insight: %w", result.Error)
	}
	return &insight, nil
}