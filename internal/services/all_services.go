package services

import (
	"github.com/Azertdev/FiberTest/internal/repositories"
)

type AllServices struct {
	UserService    UserService
	// D'autres services ici
}

func NewAllServices(userRepo repositories.UserRepository) *AllServices {
	return &AllServices{
		UserService:    NewUserService(userRepo),
		// Initialiser d'autres services ici
	}
}
