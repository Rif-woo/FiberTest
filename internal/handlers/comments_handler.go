package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Azertdev/FiberTest/internal/services"
)

type CommentHandler struct {
	commentService services.CommentService
}

func NewCommentHandler(commentService services.CommentService) CommentHandler {
	return CommentHandler{commentService}
}

func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
	videoID := c.Query("video_id")
	if videoID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Video ID is required"})
	}
	comments, err := h.commentService.GetYouTubeComments(videoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err})
	}
	return c.Status(200).JSON(comments)
}


