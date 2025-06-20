package services

import (
	"log" // Pour la validation des dépendances

	"github.com/Azertdev/FiberTest/internal/repositories"
)

type AllServices struct {
	UserService    UserService
	CommentService CommentService
}

func NewAllServices(
	allRepositories repositories.AllRepository,
	youtubeAdapter YouTubeAdapter,                  // <- Ajouté (Interface)
	groqAdapter    GroqAdapter,                     // <- Ajouté (Interface)
	transcriptUtil TranscriptUtil,                  // <- Ajouté (Interface)

) *AllServices {

	if allRepositories.UserRepository == nil {
		log.Fatal("ERREUR FATALE: UserRepository manquant lors de la création de AllServices")
	}
	if allRepositories.InsightRepository == nil {
		log.Fatal("ERREUR FATALE: InsightRepository manquant lors de la création de AllServices")
	}
	if youtubeAdapter == nil {
		log.Fatal("ERREUR FATALE: YouTubeAdapter manquant lors de la création de AllServices")
	}
	if groqAdapter == nil {
		log.Fatal("ERREUR FATALE: GroqAdapter manquant lors de la création de AllServices")
	}
	if transcriptUtil == nil {
		log.Fatal("ERREUR FATALE: TranscriptUtil manquant lors de la création de AllServices")
	}
	userService := NewUserService(allRepositories.UserRepository)

	commentService := NewCommentService(
		allRepositories.CommentRepository, // Passez le repo Commentaire (ou nil)
		allRepositories.InsightRepository,
		youtubeAdapter,
		groqAdapter,
		transcriptUtil,
	)

	return &AllServices{
		UserService:    userService,
		CommentService: commentService,
	}
}