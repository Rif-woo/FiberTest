package services

import (
	"context"
	// "time"
	"github.com/Azertdev/FiberTest/internal/models" // Adapt path if needed
)

// YouTubeAdapter defines the contract for fetching YouTube data.
type YouTubeAdapter interface {
	GetComments(ctx context.Context, videoID string, maxResults int64) ([]models.Comment, error) // Return models.Comment for simplicity now
}

// GroqAdapter defines the contract for interacting with the Groq API.
type GroqAdapter interface {
	AnalyzeComments(ctx context.Context, comments []string, videoTranscript string) (string, error)
}

// TranscriptUtil defines the contract for fetching video transcripts.
type TranscriptUtil interface {
	GetTranscript(ctx context.Context, videoID string) (string, error)
}
