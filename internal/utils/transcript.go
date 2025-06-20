// internal/utils/transcript.go
package utils

import (
	"context" // <- Importer context
	"errors"
	"fmt"
	"os/exec"
	"strings"
	// Importer l'interface depuis le package services pour s'assurer de la conformité
	// (Optionnel mais bonne pratique, évite les erreurs de frappe dans la signature)
	// "github.com/Azertdev/FiberTest/internal/services"
)

// Structure (vide pour l'instant) qui implémentera l'interface TranscriptUtil
type transcriptUtil struct{}

// Constructeur qui retourne l'interface services.TranscriptUtil
// C'est cette fonction que main.go appelle.
func NewTranscriptUtil() *transcriptUtil { // Retourne le type concret, qui implémente l'interface
	return &transcriptUtil{}
}

// Méthode TruncateTextByWords (reste une fonction utilitaire simple, pas besoin d'être une méthode)
func TruncateTextByWords(text string, maxWords int) string {
	words := strings.Fields(text)
	if len(words) > maxWords {
		words = words[:maxWords]
	}
	return strings.Join(words, " ")
}

// GetTranscript est maintenant une MÉTHODE de *transcriptUtil
// Elle DOIT respecter la signature de l'interface services.TranscriptUtil
func (tu *transcriptUtil) GetTranscript(ctx context.Context, videoID string) (string, error) {
	// Utiliser exec.CommandContext pour pouvoir potentiellement annuler/timeout la commande via le contexte
	// Note: Le script python lui-même doit aussi être conçu pour gérer l'annulation si nécessaire.
	cmd := exec.CommandContext(ctx, ".venv/bin/python", "scripts/get_transcript.py", videoID)

	// CombinedOutput attend que la commande se termine. Le contexte peut l'interrompre.
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Vérifier si l'erreur est due à l'annulation du contexte
		if ctx.Err() == context.Canceled {
			return "", errors.New("récupération de la transcription annulée")
		}
		if ctx.Err() == context.DeadlineExceeded {
			return "", errors.New("timeout lors de la récupération de la transcription")
		}
		// Autre erreur d'exécution
		return "", fmt.Errorf("erreur lors de l'exécution du script python (%s): %w, sortie: %s", videoID, err, string(output))
	}

	transcript := string(output)

	// Vérifier si le script python a retourné une erreur applicative
	if strings.HasPrefix(transcript, "ERROR:") {
		// Nettoyer le message d'erreur potentiel
		errMsg := strings.TrimSpace(strings.TrimPrefix(transcript, "ERROR:"))
		return "", errors.New(errMsg)
	}

	// Tronquer la transcription avant de la retourner
	truncatedTranscript := TruncateTextByWords(transcript, 50) // Utilise la fonction utilitaire

	return truncatedTranscript, nil
}

// Assurez-vous que *transcriptUtil implémente bien services.TranscriptUtil
// var _ services.TranscriptUtil = (*transcriptUtil)(nil) // Décommentez si vous importez l'interface
