package services

import (
	"github.com/Azertdev/FiberTest/internal/repositories"
)

type AllServices struct {
	UserService    UserService
	CommentService CommentService
	// D'autres services ici
}

func NewAllServices(userRepo repositories.UserRepository, CommentRepo repositories.CommentRepository) *AllServices {
	return &AllServices{
		UserService:    NewUserService(userRepo),
		CommentService: NewCommentService(CommentRepo),
		// Initialiser d'autres services ici
	}
}
