package repositories

import "gorm.io/gorm"

type AllRepository struct{
	UserRepository UserRepository
	CommentRepository CommentRepository
	InsightRepository InsightRepository
}

func NewAllRepository(db *gorm.DB) AllRepository{
	return AllRepository{
		UserRepository: NewUserRepository(db),
		CommentRepository: NewCommentRepository(db),
		InsightRepository: NewInsightRepository(db),
	}
}