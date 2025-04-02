// internal/adapters/youtube_adapter.go
package adapters

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Azertdev/FiberTest/internal/models"
	"github.com/Azertdev/FiberTest/internal/services"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	// Import googleapi si vous utilisez des CallOption spécifiques, sinon pas nécessaire ici.
	// "google.golang.org/api/googleapi"
)

// ... (Structure youtubeAdapter et constructeur NewYouTubeAdapter restent les mêmes) ...
type youtubeAdapter struct {
	apiKey      string
	ytService   *youtube.Service
}

func NewYouTubeAdapter(apiKey string) (services.YouTubeAdapter, error) {
    // ... (code du constructeur identique) ...
	if apiKey == "" {
		return nil, errors.New("clé API YouTube manquante pour l'adapter")
	}
	ctx := context.Background()
	ytService, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("échec de la création du service YouTube: %w", err)
	}
	log.Println("Service YouTube initialisé avec succès dans l'adapter.")
	return &youtubeAdapter{
		apiKey:    apiKey,
		ytService: ytService,
	}, nil
}


// Implémentation de la méthode GetComments de l'interface
func (a *youtubeAdapter) GetComments(ctx context.Context, videoID string, maxResults int64) ([]models.Comment, error) {
	if a.ytService == nil {
		return nil, errors.New("service YouTube non initialisé dans l'adapter")
	}

	log.Printf("Adapter: Récupération des commentaires pour videoID: %s (max: %d)", videoID, maxResults)

	call := a.ytService.CommentThreads.List([]string{"snippet"}).
		VideoId(videoID).
		TextFormat("plainText").
		MaxResults(maxResults)

	// --- CORRECTION ICI ---
	// Supprimer .WithContext(ctx) - la méthode .Do() est appelée directement sur 'call'.
	// Le contexte 'ctx' sera vérifié *après* l'erreur si nécessaire.
	// Vous pouvez passer des googleapi.CallOption ici si besoin, mais pas le contexte directement.
	response, err := call.Do()
	// --- Fin de la Correction ---

	if err != nil {
		// Il est TOUJOURS pertinent de vérifier si l'erreur provient du contexte
		// car l'annulation/timeout du contexte peut causer une erreur dans la couche transport HTTP.
		select {
		case <-ctx.Done(): // Vérifie si le contexte a été annulé ou a dépassé son délai
			log.Printf("WARN: Le contexte a été annulé ou a expiré pendant/après l'appel à YouTube API pour videoID %s: %v", videoID, ctx.Err())
			// Retourner l'erreur du contexte est souvent plus informatif dans ce cas
			return nil, fmt.Errorf("échec de la récupération des commentaires YouTube (contexte terminé: %w): %w", ctx.Err(), err)
		default:
			// Le contexte n'était pas terminé, l'erreur vient probablement de l'API elle-même.
			return nil, fmt.Errorf("erreur lors de l'appel à l'API YouTube CommentThreads pour videoID %s: %w", videoID, err)
		}
	}

	// Traitement de la réponse (reste identique)
	var comments []models.Comment
	log.Printf("Adapter: Traitement de %d threads de commentaires reçus pour videoID: %s", len(response.Items), videoID)
	for _, item := range response.Items {
		// ... (vérifications nil et parsing) ...
		if item.Snippet == nil || item.Snippet.TopLevelComment == nil || item.Snippet.TopLevelComment.Snippet == nil {
			log.Printf("WARN: Structure de commentaire inattendue reçue de l'API YouTube pour videoID %s, élément ignoré.", videoID)
			continue
		}
		snippet := item.Snippet.TopLevelComment.Snippet
		parsedTime, parseErr := time.Parse(time.RFC3339, snippet.PublishedAt)
		if parseErr != nil {
			log.Printf("WARN: Erreur de parsing de la date '%s' pour un commentaire sur videoID %s: %v. Utilisation de l'heure actuelle.", snippet.PublishedAt, videoID, parseErr)
			parsedTime = time.Now()
		}
		comment := models.Comment{
			VideoID: videoID,
			Content: snippet.TextDisplay,
			Author:  snippet.AuthorDisplayName,
			Date:    parsedTime,
		}
		comments = append(comments, comment)
	}

	log.Printf("Adapter: %d commentaires formatés retournés pour videoID: %s", len(comments), videoID)
	return comments, nil
}