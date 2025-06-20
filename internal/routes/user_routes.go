package routes

import (
	"github.com/Azertdev/FiberTest/internal/handlers"
	"github.com/Azertdev/FiberTest/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(app *fiber.App, userHandler handlers.UserHandler) {
	userGroup := app.Group("/users")
	userGroup.Post("/", userHandler.CreateUser)
	userGroup.Get("/", middleware.JWTMiddleware, userHandler.GetAllUsers)
	userGroup.Get("/:id", middleware.JWTMiddleware, userHandler.GetUserByID)
	userGroup.Post("/authenticate", userHandler.LoginHandler)
}