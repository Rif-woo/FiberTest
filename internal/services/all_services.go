package services

import (
	"github.com/Azertdev/FiberTest/internal/repositories"
)

type AllServices struct {
	UserService    *userService
	CommentService *CommentService
	// D'autres services ici
}

func NewAllServices(allRepositories *repositories.AllRepository) *AllServices {
	return &AllServices{
		UserService:    NewUserService(allRepositories.UserRepository),
		CommentService: NewCommentService(allRepositories.CommentRepository),
		// Initialiser d'autres services ici
	}
}
