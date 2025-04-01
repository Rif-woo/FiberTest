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

	allRepositories := repositories.NewAllRepository(config.DB)
	allServices := services.NewAllServices(allRepositories)
	allHandlers := handlers.NewAllHandlers(allServices.UserService, *allServices.CommentService)

	app := fiber.New()
	
	routes.SetupUserRoutes(app, &allHandlers.UserHandler)
	routes.SetupCommentsRoutes(app, &allHandlers.CommentHandler)

	log.Fatal(app.Listen(":3001"))
}
