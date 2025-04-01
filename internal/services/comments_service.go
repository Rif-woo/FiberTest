package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Azertdev/FiberTest/internal/models"
	"github.com/Azertdev/FiberTest/internal/repositories"
	"github.com/Azertdev/FiberTest/internal/utils"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type CommentService interface {
	FindAll() ([]models.Comment, error)
	FindByID(id uint) (*models.Comment, error)
	GetYouTubeComments(videoID string) ([]models.Comment, error)
	AnalyzeYouTubeComments(videoID string) (string, error)
}

type commentService struct {
	commentRepo repositories.CommentRepository
}

func NewCommentService(commentRepo repositories.CommentRepository) CommentService {
	return &commentService{commentRepo}
}

func (s *commentService) FindAll() ([]models.Comment, error) {
	return s.commentRepo.FindAll()
}

func (s *commentService) FindByID(id uint) (*models.Comment, error) {
	return s.commentRepo.FindByID(id)
}

func (s *commentService) GetYouTubeComments(videoID string) ([]models.Comment, error) {
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


func (s *commentService) AnalyzeYouTubeComments(videoID string) (string, error) {
	comments, err := s.GetYouTubeComments(videoID)
	if err != nil {
		return "", err
	}

	var contents []string
	for _, c := range comments {
		contents = append(contents, fmt.Sprintf(
		"Auteur : %s | Date : %s | Commentaire : \"%s\"",
		c.Author,
		c.Date.Format("2006-01-02"),
		c.Content,
	))
	}

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return "", errors.New("GROQ_API_KEY is not set")
	}

	// Appel de OpenAI ici
	ai := NewGroqService(apiKey) // ou injecté dans all_services.go
	videoTranscript, err := utils.GetTranscript(videoID)
	if err != nil {
		return "", err
	}
	result, err := ai.AnalyzeComments(contents, videoTranscript)
	if err != nil {
		return "", err
	}

	parsed := utils.ParseInsightResponse(result)
	jsonBytes, err := json.Marshal(parsed)
	if err != nil {
		return "", err
	}


	return string(jsonBytes), nil
}
