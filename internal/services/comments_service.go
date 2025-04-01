package services

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/Azertdev/FiberTest/internal/models"
	"github.com/Azertdev/FiberTest/internal/repositories"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)


type CommentService struct {
	commentRepo repositories.CommentRepository
}

func NewCommentService(commentRepo repositories.CommentRepository) *CommentService {
	return &CommentService{commentRepo}
}

func (s *CommentService) FindAll() ([]models.Comment, error) {
	return s.commentRepo.FindAllComment()
}

func (s *CommentService) FindByID(id uint) (*models.Comment, error) {
	return s.commentRepo.FindCommentByID(id)
}

func (s *CommentService) GetYouTubeComments(videoID string) ([]models.Comment, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return nil, errors.New("YOUTUBE_API_KEY is not set")
	}

	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil,err
	}

	call := youtubeService.CommentThreads.List([]string{"snippet"}).
		VideoId(videoID).
		TextFormat("plainText").
		MaxResults(50) // Récupérer au max 50 commentaires

	response, err := call.Do()
	if err != nil {
		return nil,err
	}

	var comments []models.Comment
	for _, item := range response.Items {
		parsedTime, err := time.Parse(time.RFC3339, item.Snippet.TopLevelComment.Snippet.PublishedAt)
		if err != nil {
			log.Fatalf("Erreur de parsing de la date : %v", err) // Arrête le programme si la date est invalide
		}
		comment := models.Comment{
			VideoID: videoID,
			Content:    item.Snippet.TopLevelComment.Snippet.TextDisplay,
			Author:  item.Snippet.TopLevelComment.Snippet.AuthorDisplayName,
			Date:    parsedTime,
		}
		comments = append(comments, comment)
	}

	// store comments in database
	// s.commentRepo.SaveYouTubeComments(comments)
	
	return comments, nil
}
