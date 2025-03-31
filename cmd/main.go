package main

import (
	"log"
	"github.com/Azertdev/FiberTest/config"
	"github.com/Azertdev/FiberTest/internal/handlers"
	"github.com/Azertdev/FiberTest/internal/repositories"
	"github.com/Azertdev/FiberTest/internal/routes"
	"github.com/Azertdev/FiberTest/internal/services"

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.InitDB()

	userRepo := repositories.NewUserRepository(config.DB)
	commentRepo := repositories.NewCommentRepository(config.DB)
	// Initialize services
	allServices := services.NewAllServices(userRepo, commentRepo)
	
	userHandler := handlers.NewUserHandler(allServices.UserService)
	commentHandler := handlers.NewCommentHandler(allServices.CommentService)

	app := fiber.New()
	
	routes.SetupUserRoutes(app, userHandler)
	routes.SetupCommentsRoutes(app, commentHandler)

	log.Fatal(app.Listen(":3001"))
}
