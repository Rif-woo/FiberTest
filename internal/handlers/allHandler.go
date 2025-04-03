package handlers

import "github.com/Azertdev/FiberTest/internal/services"

type AllHandlers struct{
	UserHandler UserHandler
	CommentHandler CommentHandler
}

func NewAllHandlers(UserHandler services.UserService, CommentHandler services.CommentService) AllHandlers{
	return AllHandlers{
		UserHandler: NewUserHandler(UserHandler),
		CommentHandler: NewCommentHandler(CommentHandler),
	}
}