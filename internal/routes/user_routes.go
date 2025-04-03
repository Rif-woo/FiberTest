package routes

import (
	"github.com/Azertdev/FiberTest/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App, userHandler handlers.UserHandler) {
	userGroup := app.Group("/users")
	userGroup.Post("/", userHandler.CreateUser)
	userGroup.Get("/", userHandler.GetAllUsers)
	userGroup.Get("/:id", userHandler.GetUserByID)
	userGroup.Post("/authenticate", userHandler.LoginHandler)
}