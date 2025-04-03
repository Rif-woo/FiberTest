// internal/handlers/comments_handler.go
package handlers

import (
	// "fmt" // Importer fmt pour formater les erreurs
	"log" // Importer log pour le logging

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid" // <-- NÉCESSAIRE pour générer et utiliser des UUIDs

	// Importer models si vous voulez vérifier le type de retour (optionnel)
	// "github.com/Azertdev/FiberTest/internal/models"
	"github.com/Azertdev/FiberTest/internal/services"
)

type CommentHandler struct {
	commentService services.CommentService
}

func NewCommentHandler(commentService services.CommentService) CommentHandler {
	return CommentHandler{commentService}
}

// Renommer la fonction est une bonne pratique pour refléter l'action (Analyse)
// mais gardons GetComments si vous préférez pour l'instant.
// func (h *CommentHandler) HandleAnalyzeCommentsRequest(c *fiber.Ctx) error {
func (h *CommentHandler) GetComments(c *fiber.Ctx) error { // Gardons votre nom actuel pour l'instant
	videoID := c.Query("video_id")
	if videoID == "" {
		// Réponse d'erreur structurée
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Paramètre 'video_id' manquant dans la query string",
		})
	}

	// --- CORRECTION PRINCIPALE ---
	// Supprimer : userID := "testUser"
	// Logique pour obtenir ou générer un UUID valide
	var finalUserID uuid.UUID // Variable pour l'UUID à utiliser

	// Tenter de récupérer depuis les locals (pour le futur quand l'auth sera là)
	userIDClaim := c.Locals("userID")
	if userIDClaim != nil {
		var ok bool
		finalUserID, ok = userIDClaim.(uuid.UUID)
		if !ok {
			log.Printf("ERROR: userID trouvé dans c.Locals mais type invalide: %T", userIDClaim)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Erreur interne (type userID invalide)",
			})
		}
		log.Printf("INFO: Utilisation de userID %s depuis c.Locals", finalUserID)
	} else {
		// Cas actuel : Pas d'authentification, générer un UUID temporaire
		finalUserID = uuid.New() // Génère un UUID v4 aléatoire
		log.Printf("WARN: Aucun userID dans c.Locals. Génération d'un UUID temporaire: %s", finalUserID)
		// Note : Quand l'auth sera obligatoire, retourner 401 ici au lieu de générer.
	}
	// --- FIN CORRECTION ---


	// Appel du service avec le finalUserID (qui est maintenant un uuid.UUID)
	log.Printf("INFO: Début analyse pour videoID: %s, userID: %s", videoID, finalUserID)
	insight, err := h.commentService.AnalyzeAndSaveYouTubeComments(c.Context(), finalUserID, videoID)
	if err != nil {
		log.Printf("ERROR: Échec AnalyzeAndSaveYouTubeComments pour videoID %s, userID %s: %v", videoID, finalUserID, err)
		// Réponse d'erreur structurée et plus générique pour le client
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			// Masquer les détails de l'erreur interne au client
			"message": "Échec lors de l'analyse des commentaires.",
			// Vous pouvez ajouter un code d'erreur interne si nécessaire
			// "code": "ANALYSIS_FAILED",
		})
	}

	// Réponse de succès structurée
	log.Printf("INFO: Analyse réussie pour videoID: %s, userID %s. Insight ID: %s", videoID, finalUserID, insight.ID)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{ // Utiliser 201 Created est sémantiquement mieux ici
		"status":  "success",
		"message": "Analyse terminée et sauvegardée.",
		// Renommer "insights" en "data" est plus standard et utiliser l'objet insight directement
		"data": insight,
	})
}