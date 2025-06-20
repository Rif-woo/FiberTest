package repositories

import (
	"github.com/Azertdev/FiberTest/internal/models"
	"gorm.io/gorm"
)

type CommentRepository interface {
	FindAllComment() ([]models.Comment, error)
	FindCommentByID(id uint) (*models.Comment, error)
	SaveYouTubeComments(comments []models.Comment) error
}

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &CommentRepo{db}
}

func (r *CommentRepo) FindAllComment() ([]models.Comment, error) {
	var comments []models.Comment
	err := r.db.Find(&comments).Error
	return comments, err
}

func (r *CommentRepo) FindCommentByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.First(&comment, id).Error
	return &comment, err
}

func (r *CommentRepo) SaveYouTubeComments(comments []models.Comment) error {
	return r.db.Create(&comments).Error
}
