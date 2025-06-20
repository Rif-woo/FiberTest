// internal/services/comment_service.go
package services

import (
	"context"
	"encoding/json"
	"time"

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

func NewCommentService(
	commentRepo repositories.CommentRepository,
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
	return s.commentRepo.FindAllComment()
}

func (s *commentService) FindByID(id uint) (*models.Comment, error) {
	if s.commentRepo == nil {
		return nil, fmt.Errorf("CommentRepository non initialisé dans CommentService")
	}
	return s.commentRepo.FindCommentByID(id)
}

func (s *commentService) AnalyzeAndSaveYouTubeComments(ctx context.Context, userID uuid.UUID, videoID string) (*models.Insight, error) {

	// --- Étape 1: Récupération des commentaires ---
	log.Printf("INFO: [UserID: %s] Récupération des commentaires pour videoID: %s", userID, videoID)
	// Note: Fetching 100 comments increases the chance of needing chunking.
	maxCommentsToFetch := int64(2000) // Configurable ?
	commentsData, err := s.youtubeAdapter.GetComments(ctx, videoID, maxCommentsToFetch)
	if err != nil { return nil, fmt.Errorf("échec récupération commentaires YouTube: %w", err) }
	if len(commentsData) == 0 { return nil, fmt.Errorf("aucun commentaire trouvé pour videoID %s", videoID) }
	log.Printf("INFO: [UserID: %s] %d commentaires récupérés pour videoID: %s", userID, len(commentsData), videoID)


	// --- Étape 2: Récupération et Résumé de la Transcription (Appel IA Séparé) ---
	log.Printf("INFO: [UserID: %s] Récupération transcription brute pour videoID: %s", userID, videoID)
	rawTranscript, err := s.transcriptUtil.GetTranscript(ctx, videoID)
	transcriptSummary := "Résumé non généré (erreur récupération transcript)." // Default
	// transcriptForAnalysis := "Transcription non disponible."                   // Default context for comment analysis

	if err == nil {
		log.Printf("INFO: [UserID: %s] Transcription brute récupérée. Génération du résumé...", userID)
		// Tronquer AVANT de résumer si trop long pour l'input de SummarizeTranscript
		// transcriptToSummarize := utils.TruncateTextByWords(rawTranscript, 15000) // Exemple
		summary, summaryErr := s.groqAdapter.SummarizeTranscript(ctx, rawTranscript) // Utiliser rawTranscript (ou tronqué si besoin)
		if summaryErr != nil {
			log.Printf("WARN: [UserID: %s] Échec génération résumé transcript: %v", userID, summaryErr)
			transcriptSummary = "Résumé non généré (erreur IA)."
		} else {
			transcriptSummary = summary
			log.Printf("INFO: [UserID: %s] Résumé transcript généré.", userID)
		}

		// Préparer le contexte (tronqué) pour l'analyse des commentaires
		// maxWordsForCommentContext := 5000 // <-- AJUSTER cette valeur (très important !)
		// transcriptForAnalysis = utils.TruncateTextByWords(rawTranscript, maxWordsForCommentContext)
		// log.Printf("INFO: [UserID: %s] Transcription tronquée à %d mots pour contexte analyse commentaires.", userID, maxWordsForCommentContext)

	} else {
		log.Printf("WARN: [UserID: %s] Échec récupération transcription: %v. Analyse sans contexte transcript.", userID, err)
		// transcriptSummary et transcriptForAnalysis gardent leurs valeurs par défaut
	}


	// --- Étape 3: Chunking et Analyse des Commentaires ---
	chunkSize := 50 // <-- AJUSTER cette valeur (nombre de commentaires par appel Groq)
	var allParsedInsights []*utils.ParsedInsight // Pour stocker les résultats de chaque chunk
	totalChunks := (len(commentsData) + chunkSize - 1) / chunkSize

	log.Printf("INFO: [UserID: %s] Début de l'analyse des commentaires par lots (taille: %d, total: %d) pour videoID: %s", userID, chunkSize, totalChunks, videoID)

	for i := 0; i < len(commentsData); i += chunkSize {
		end := i + chunkSize
		if end > len(commentsData) {
			end = len(commentsData)
		}
		commentChunk := commentsData[i:end] // Le lot actuel de commentaires
		currentChunkNum := (i / chunkSize) + 1

		// Formatage des commentaires pour CE lot
		var chunkContents []string
		for _, c := range commentChunk {
			chunkContents = append(chunkContents, fmt.Sprintf(
				"Auteur: %s | Date: %s | Commentaire: \"%s\"",
				c.Author, c.Date.Format("2006-01-02"), c.Content,
			))
		}

		log.Printf("INFO: [UserID: %s] Analyse du lot %d/%d (taille %d)...", userID, currentChunkNum, totalChunks, len(chunkContents))

		// Appel à Groq pour CE LOT avec le contexte transcript (tronqué)
		markdownChunkResult, err := s.groqAdapter.AnalyzeComments(ctx, chunkContents, rawTranscript)
		if err != nil {
			// Que faire si un lot échoue ? Logguer et continuer ? Ou échouer tout ?
			// Pour l'instant, on loggue et on continue au lot suivant.
			log.Printf("WARN: [UserID: %s] Échec analyse du lot %d/%d pour videoID %s: %v. Lot ignoré.", userID, currentChunkNum, totalChunks, videoID, err)
			continue // Passe au lot suivant
		}

		// Parser le résultat Markdown du lot
		parsedChunk := utils.ParseInsightResponse(markdownChunkResult)
		if parsedChunk != nil { // Vérifier si le parsing a réussi
			allParsedInsights = append(allParsedInsights, parsedChunk)
			log.Printf("INFO: [UserID: %s] Lot %d/%d analysé et parsé avec succès.", userID, currentChunkNum, totalChunks)
		} else {
			log.Printf("WARN: [UserID: %s] Échec parsing du résultat du lot %d/%d pour videoID %s. Lot ignoré.", userID, currentChunkNum, totalChunks, videoID)
		}

		// Optionnel: Pause pour éviter de surcharger les limites TPM trop rapidement
        if totalChunks > 1 && currentChunkNum < totalChunks { // Ne pas attendre après le dernier chunk
		    time.Sleep(500 * time.Millisecond) // Attente de 500ms (configurable !)
        }
	} // Fin de la boucle des chunks


	// --- Étape 4: Fusion des Résultats Parsés ---
	if len(allParsedInsights) == 0 {
		// Aucun lot n'a réussi
		log.Printf("ERROR: [UserID: %s] Aucun lot de commentaires n'a pu être analysé avec succès pour videoID %s.", userID, videoID)
		return nil, fmt.Errorf("échec de l'analyse d'au moins un lot de commentaires")
	}

	log.Printf("INFO: [UserID: %s] Fusion des résultats de %d lots analysés pour videoID: %s", userID, len(allParsedInsights), videoID)
	finalParsedInsight := utils.MergeParsedInsights(allParsedInsights) // Appel de la fonction de fusion (à définir ci-dessous)


	// --- Étape 5: Mapping vers models.Insight (utilise finalParsedInsight) ---
	log.Printf("INFO: [UserID: %s] Mapping final vers le modèle Insight pour videoID: %s", userID, videoID)
	newInsight := &models.Insight{
		UserID:            userID,
		VideoID:           videoID,
		Sentiment:         finalParsedInsight.Sentiment, // Sentiment issu de la fusion
		Summary:           finalParsedInsight.Summary,   // Résumé issu de la fusion
		TranscriptSummary: transcriptSummary,          // Résumé de la transcription (fait séparément)
	}
	marshalToJson := func(fieldName string, data interface{}) datatypes.JSON { /* ... (helper identique) ... */
        bytes, err := json.Marshal(data)
        if err != nil { log.Printf("gvgv"); return datatypes.JSON("[]") }
        return datatypes.JSON(bytes)
    }
	newInsight.TopComments = marshalToJson("TopComments", finalParsedInsight.TopComments)
	newInsight.NegativeComments = marshalToJson("NegativeComments", finalParsedInsight.NegativeComments)
	newInsight.QuestionComments = marshalToJson("QuestionComments", finalParsedInsight.QuestionComments)
	newInsight.FeedbackComments = marshalToJson("FeedbackComments", finalParsedInsight.FeedbackComments)
	newInsight.Keywords = marshalToJson("Keywords", finalParsedInsight.Keywords)


	// --- Étape 6: Sauvegarde en base ---
	log.Printf("INFO: [UserID: %s] Sauvegarde de l'insight fusionné en base pour videoID: %s", userID, videoID)
	err = s.insightRepo.CreateInsight(ctx, newInsight)
	if err != nil { return nil, fmt.Errorf("échec sauvegarde insight fusionné en base: %w", err) }


	// --- Étape 7: Retourner l'insight ---
	log.Printf("INFO: [UserID: %s] Insight fusionné sauvegardé avec succès pour videoID %s. ID: %s", userID, videoID, newInsight.ID)
	return newInsight, nil
}
