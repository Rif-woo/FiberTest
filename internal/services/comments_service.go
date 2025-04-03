// internal/services/comment_service.go
package services

import (
	"context"
	"encoding/json"
	// "errors" // errors n'est plus utilisé directement ici pour les API Keys
	"fmt"
	"log"
	// "time" // Garder l'import time

	"github.com/google/uuid" // Import pour UUID
	"gorm.io/datatypes"      // Import pour datatypes.JSON

	"github.com/Azertdev/FiberTest/internal/models"
	"github.com/Azertdev/FiberTest/internal/repositories"
	"github.com/Azertdev/FiberTest/internal/utils"
	
)

type CommentService interface {
	FindAll() ([]models.Comment, error)
	FindByID(id uint) (*models.Comment, error)
	AnalyzeAndSaveYouTubeComments(ctx context.Context, userID uuid.UUID, videoID string) (*models.Insight, error)
}

type commentService struct {
	commentRepo    repositories.CommentRepository // Conservé si FindAll/FindByID sont utilisés
	insightRepo    repositories.InsightRepository  // Injection du repo Insight
	youtubeAdapter YouTubeAdapter                  // Injection de l'adapter YouTube
	groqAdapter    GroqAdapter                     // Injection de l'adapter Groq
	transcriptUtil TranscriptUtil                  // Injection de l'utilitaire de transcription
}

// NewCommentService est le constructeur pour commentService.
// Il prend toutes les dépendances nécessaires.
func NewCommentService(
	commentRepo repositories.CommentRepository, // Peut être nil si FindAll/FindByID non utilisés
	insightRepo repositories.InsightRepository,
	youtubeAdapter YouTubeAdapter,
	groqAdapter GroqAdapter,
	transcriptUtil TranscriptUtil,
) CommentService { // Retourne l'interface
	// Validation rapide des dépendances critiques
	if insightRepo == nil || youtubeAdapter == nil || groqAdapter == nil || transcriptUtil == nil {
		log.Fatal("ERREUR FATALE: Dépendances manquantes lors de la création de CommentService")
	}
	return &commentService{
		commentRepo:    commentRepo,
		insightRepo:    insightRepo,
		youtubeAdapter: youtubeAdapter,
		groqAdapter:    groqAdapter,
		transcriptUtil: transcriptUtil,
	}
}

// --- Méthodes existantes (FindAll, FindByID) ---

func (s *commentService) FindAll() ([]models.Comment, error) {
	if s.commentRepo == nil {
		return nil, fmt.Errorf("CommentRepository non initialisé dans CommentService")
	}
	return s.commentRepo.FindAll()
}

func (s *commentService) FindByID(id uint) (*models.Comment, error) {
	if s.commentRepo == nil {
		return nil, fmt.Errorf("CommentRepository non initialisé dans CommentService")
	}
	return s.commentRepo.FindByID(id)
}

// --- Méthode GetYouTubeComments originale SUPPRIMÉE ---
// La logique d'appel à l'API YouTube est maintenant dans l'implémentation de YouTubeAdapter.

// --- Méthode AnalyzeYouTubeComments originale (retournant string) SUPPRIMÉE ---
// Remplacée par AnalyzeAndSaveYouTubeComments ci-dessous.

// --- NOUVELLE MÉTHODE : AnalyzeAndSaveYouTubeComments ---

// AnalyzeAndSaveYouTubeComments orchestre le processus complet :
// récupération des commentaires, transcription, analyse IA, parsing, et sauvegarde.
func (s *commentService) AnalyzeAndSaveYouTubeComments(ctx context.Context, userID uuid.UUID, videoID string) (*models.Insight, error) {
	// 1. Vérifier si un insight existe déjà (Optionnel - décommentez et adaptez si nécessaire)
	// existingInsight, err := s.insightRepo.GetInsightByVideoID(ctx, userID, videoID)
	// if err != nil {
	//     // Ne pas bloquer si la vérification échoue, juste logguer peut-être ? Ou retourner l'erreur.
	//     log.Printf("WARN: [UserID: %s] Erreur lors de la vérification de l'insight existant pour videoID %s: %v", userID, videoID, err)
	// }
	// if existingInsight != nil {
	//     log.Printf("INFO: [UserID: %s] Insight déjà existant pour videoID %s. Retour de l'existant.", userID, videoID)
	//     return existingInsight, nil // Retourner l'insight existant si trouvé
	// }

	// 2. Récupérer les commentaires via l'adapter YouTube
	log.Printf("INFO: [UserID: %s] Récupération des commentaires pour videoID: %s", userID, videoID)
	maxComments := int64(50) // Rendre configurable si besoin
	commentsData, err := s.youtubeAdapter.GetComments(ctx, videoID, maxComments)
	if err != nil {
		// Erreur critique, on ne peut pas continuer sans commentaires
		return nil, fmt.Errorf("échec de la récupération des commentaires YouTube pour videoID %s: %w", videoID, err)
	}
	if len(commentsData) == 0 {
		// Gérer le cas sans commentaires : retourner une erreur claire ? ou un insight "vide" ?
		// Pour l'instant, retourne une erreur.
		log.Printf("INFO: [UserID: %s] Aucun commentaire trouvé pour videoID %s.", userID, videoID)
		return nil, fmt.Errorf("aucun commentaire trouvé pour la vidéo %s", videoID) // Erreur spécifique
	}
	log.Printf("INFO: [UserID: %s] %d commentaires récupérés pour videoID: %s", userID, len(commentsData), videoID)

	// Formatage des commentaires pour le prompt Groq
	var commentContents []string
	for _, c := range commentsData {
		commentContents = append(commentContents, fmt.Sprintf(
			"Auteur: %s | Date: %s | Commentaire: \"%s\"", // Format cohérent avec le prompt
			c.Author,
			c.Date.Format("2006-01-02"), // Utiliser le champ Date du modèle
			c.Content,                  // Utiliser le champ Content du modèle
		))
	}

	// 3. Récupérer la transcription via l'utilitaire de transcription
	log.Printf("INFO: [UserID: %s] Récupération de la transcription pour videoID: %s", userID, videoID)
	videoTranscript, err := s.transcriptUtil.GetTranscript(ctx, videoID)
	if err != nil {
		// Non critique : on logue un avertissement mais on continue sans transcription
		log.Printf("WARN: [UserID: %s] Échec de la récupération de la transcription pour videoID %s: %v. Analyse sans transcription.", userID, videoID, err)
		videoTranscript = "Transcription non disponible." // Fournir une valeur par défaut
	} else {
		log.Printf("INFO: [UserID: %s] Transcription récupérée pour videoID: %s", userID, videoID)
		// TODO: Ajouter ici la logique pour tronquer videoTranscript si elle est trop longue
		// maxLength := 4000 // Exemple (en caractères ou tokens, selon Groq)
		// if len(videoTranscript) > maxLength {
		//     videoTranscript = videoTranscript[:maxLength] + "..." // Troncature simple
		//     log.Printf("INFO: [UserID: %s] Transcription tronquée pour videoID: %s", userID, videoID)
		// }
	}

	// 4. Appeler Groq pour l'analyse via l'adapter Groq
	log.Printf("INFO: [UserID: %s] Appel de l'analyse Groq pour videoID: %s", userID, videoID)
	markdownResult, err := s.groqAdapter.AnalyzeComments(ctx, commentContents, videoTranscript)
	if err != nil {
		// Erreur critique de l'analyse IA
		return nil, fmt.Errorf("échec de l'analyse Groq pour videoID %s: %w", videoID, err)
	}
	// Décommenter pour déboguer la sortie brute de Groq
	// log.Printf("DEBUG: [UserID: %s] Markdown brut reçu de Groq pour videoID %s:\n---\n%s\n---\n", userID, videoID, markdownResult)

	// 5. Parser la réponse Markdown en utilisant l'utilitaire
	log.Printf("INFO: [UserID: %s] Parsing de la réponse Groq pour videoID: %s", userID, videoID)
	parsedInsightData := utils.ParseInsightResponse(markdownResult) // Utilise la fonction de parsing existante
	// Décommenter pour déboguer la structure parsée
	// log.Printf("DEBUG: [UserID: %s] Insight parsé pour videoID %s: %+v", userID, videoID, parsedInsightData)

	// 6. Mapper ParsedInsight vers models.Insight et Marshal les champs JSON
	log.Printf("INFO: [UserID: %s] Mapping vers le modèle Insight pour videoID: %s", userID, videoID)
	newInsight := &models.Insight{
		UserID:    userID, // Utilisation de l'userID fourni
		VideoID:   videoID,
		Sentiment: parsedInsightData.Sentiment,
		Summary:   parsedInsightData.Summary,
		// ID et CreatedAt sont gérés automatiquement par GORM/DB
	}

	// Fonction utilitaire interne pour marshaler en JSON ou retourner un JSON vide/null
	marshalToJson := func(fieldName string, data interface{}) datatypes.JSON {
		bytes, err := json.Marshal(data)
		if err != nil {
			// Log l'erreur mais ne bloque pas, retourne un JSON vide '[]' (ou 'null' si vous préférez)
			log.Printf("WARN: [UserID: %s] Échec du marshalling JSON pour le champ '%s' (videoID: %s): %v. Utilisation de '[]'.", userID, fieldName, videoID, err)
			// Vous pourriez vouloir retourner datatypes.JSON("null") selon la gestion en base/front
			return datatypes.JSON("[]")
		}
		return datatypes.JSON(bytes)
	}

	newInsight.TopComments = marshalToJson("TopComments", parsedInsightData.TopComments)
	newInsight.NegativeComments = marshalToJson("NegativeComments", parsedInsightData.NegativeComments)
	newInsight.QuestionComments = marshalToJson("QuestionComments", parsedInsightData.QuestionComments)
	newInsight.FeedbackComments = marshalToJson("FeedbackComments", parsedInsightData.FeedbackComments)
	newInsight.Keywords = marshalToJson("Keywords", parsedInsightData.Keywords)

	// 7. Sauvegarder l'Insight en base de données via le repository Insight
	log.Printf("INFO: [UserID: %s] Sauvegarde de l'insight en base pour videoID: %s", userID, videoID)
	err = s.insightRepo.CreateInsight(ctx, newInsight)
	if err != nil {
		// Erreur lors de la sauvegarde en base
		// TODO: Gérer les erreurs spécifiques, ex: violation de contrainte unique si l'insight existe déjà
		return nil, fmt.Errorf("échec de la sauvegarde de l'insight en base pour videoID %s: %w", videoID, err)
	}

	log.Printf("INFO: [UserID: %s] Insight sauvegardé avec succès pour videoID %s. Nouvel ID: %s", userID, videoID, newInsight.ID)

	// 8. Retourner l'insight qui vient d'être créé (avec son ID et CreatedAt remplis par GORM)
	return newInsight, nil
}