package repositories

import (
	"github.com/Azertdev/FiberTest/internal/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	FindAll() ([]models.Comment, error)
	FindByID(id uint) (*models.Comment, error)
	SaveYouTubeComments(comments []models.Comment) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db}
}

func (r *commentRepository) FindAll() ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Find(&comments).Error
	return comments, err
}

func (r *commentRepository) FindByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, id).Error
	return &comment, err
}

func (r *commentRepository) SaveYouTubeComments(comments []models.Comment) error {
	return r.db.Create(&comments).Error
}
