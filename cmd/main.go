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

	// Initialize services
	allServices := services.NewAllServices(userRepo)
	
	userHandler := handlers.NewUserHandler(allServices.UserService)

	app := fiber.New()
	routes.SetupUserRoutes(app, userHandler)

	log.Fatal(app.Listen(":3001"))
}
