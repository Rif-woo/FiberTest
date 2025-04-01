package services

import (
	"os"

	"github.com/Azertdev/FiberTest/internal/repositories"
)

type AllServices struct {
	UserService    UserService
	CommentService CommentService
	GroqService GroqService
	// D'autres services ici
}

func NewAllServices(userRepo repositories.UserRepository, CommentRepo repositories.CommentRepository) *AllServices {
	return &AllServices{
		UserService:    NewUserService(userRepo),
		CommentService: NewCommentService(CommentRepo),
		GroqService: *NewGroqService(os.Getenv("GROQ_API_KEY")),
		// Initialiser d'autres services ici
	}
}
