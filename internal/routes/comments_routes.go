package routes

import (
	"github.com/Azertdev/FiberTest/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupCommentsRoutes(app *fiber.App, commentsHandler handlers.CommentHandler) {
	commentGroup := app.Group("/comments")
	commentGroup.Get("/", commentsHandler.GetComments)
	// userGroup.Get("/", userHandler.GetAllUsers)
	// userGroup.Get("/:id", userHandler.GetUserByID)
}