package adapters

// internal/adapters/groq_adapter.go

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Azertdev/FiberTest/internal/services" // Pour l'interface
)

// Structure qui implémente services.GroqAdapter
type groqAdapter struct {
	apiKey string
	client *http.Client // Garde un client HTTP réutilisable
	model  string       // Nom du modèle Groq à utiliser (ex: "llama3-70b-8192")
}

// Constructeur pour groqAdapter
// Prend la clé API et retourne l'INTERFACE services.GroqAdapter
func NewGroqAdapter(apiKey string) services.GroqAdapter {
	if apiKey == "" {
		// Ne pas retourner d'erreur ici, car main.go vérifie déjà.
		// Si vous voulez une double vérification, retournez (nil, error).
		log.Println("WARN: Tentative de création de GroqAdapter sans clé API.")
	}
	return &groqAdapter{
		apiKey: apiKey,
		client: &http.Client{Timeout: 90 * time.Second}, // Timeout configurable
		model:  "llama3-70b-8192",                 // Modèle par défaut, pourrait être configurable
	}
}

// Implémentation de la méthode AnalyzeComments de l'interface
func (ga *groqAdapter) AnalyzeComments(ctx context.Context, comments []string, videoTranscript string) (string, error) {
	if ga.apiKey == "" {
		return "", errors.New("GroqAdapter non configuré avec une clé API")
	}
	if len(comments) == 0 {
		return "", errors.New("aucun commentaire fourni pour l'analyse Groq")
	}

	// --- Construction du Prompt (copié/adapté depuis l'ancienne logique) ---
	commentsFormatted := strings.Join(comments, "\n- ")
	// Tronquer la transcription si nécessaire (la logique de troncature pourrait être dans le service avant l'appel à l'adapter)
	// if len(videoTranscript) > 4000 { videoTranscript = videoTranscript[:4000] + "..." }

	prompt := fmt.Sprintf(`
# RÔLE ET OBJECTIF
Tu es un analyste expert spécialisé dans l'analyse des retours de communauté sur YouTube. Ton objectif est d'extraire des informations clés et structurées à partir des commentaires d'une vidéo, en utilisant la transcription fournie comme contexte général.

# CONTEXTE : TRANSCRIPTION DE LA VIDÉO
"""
%s
"""

# DONNÉES À ANALYSER : COMMENTAIRES UTILISATEURS
(Chaque ligne est un commentaire incluant l'auteur et le texte exact)
- %s

# INSTRUCTIONS D'ANALYSE ET FORMAT DE SORTIE OBLIGATOIRE
Analyse **UNIQUEMENT LES COMMENTAIRES** fournis ci-dessus pour répondre aux sections 3 à 7. Utilise la transcription seulement pour comprendre le contexte général.
Structure IMPÉRATIVEMENT ta réponse en utilisant le format Markdown suivant, avec exactement ces titres de section :

## 1. Sentiment Général
Décris en une phrase concise le sentiment dominant qui se dégage des commentaires (ex: Majoritairement Positif, Négatif, Neutre, Partagé avec des points spécifiques, Enthousiaste mais avec des questions techniques).

## 2. Résumé Général des Commentaires
Rédige un court paragraphe (3-5 phrases maximum) résumant les thèmes principaux, les points de discussion récurrents et les réactions générales observées DANS LES COMMENTAIRES.

## 3. Questions Posées
Liste textuellement les questions claires posées par les utilisateurs DANS LES COMMENTAIRES. Si aucune question n'est trouvée, écris "Aucune question identifiée.". Ne liste que les questions, pas de phrases introductives.
- [Question 1 textuelle telle qu'écrite par l'utilisateur]
- [Question 2 textuelle...]

## 4. Critiques Négatives
Liste jusqu'à 5 commentaires (ou extraits les plus pertinents) exprimant des critiques claires ou un mécontentement DANS LES COMMENTAIRES. Cite le commentaire exact ou l'extrait significatif. Si aucune critique n'est trouvée, écris "Aucune critique négative significative identifiée.".
- "[Auteur:] [Extrait ou commentaire négatif 1]"
- "[Auteur:] [Extrait ou commentaire négatif 2]"

## 5. Points Positifs ou Constructifs
Liste jusqu'à 5 commentaires (ou extraits les plus pertinents) exprimant un avis positif marqué, un encouragement, ou une suggestion constructive DANS LES COMMENTAIRES. Cite le commentaire exact ou l'extrait significatif. Si aucun point positif notable n'est trouvé, écris "Aucun commentaire positif ou constructif notable identifié.".
- "[Auteur:] [Extrait ou commentaire positif/constructif 1]"
- "[Auteur:] [Extrait ou commentaire positif/constructif 2]"

## 6. Feedbacks Spécifiques ou Techniques
Liste jusqu'à 5 commentaires (ou extraits pertinents) DANS LES COMMENTAIRES contenant des retours d'expérience détaillés, des suggestions techniques précises, des corrections factuelles ou des remarques spécifiques pointues liées au contenu. Cite le commentaire exact ou l'extrait significatif. Si aucun feedback de ce type n'est trouvé, écris "Aucun feedback spécifique ou technique identifié.".
- "[Auteur:] [Extrait ou commentaire de feedback 1]"
- "[Auteur:] [Extrait ou commentaire de feedback 2]"

## 7. Mots-clés et Thèmes Fréquents
Liste les 5 à 10 mots-clés ou courtes expressions (1-3 mots) les plus fréquents et pertinents issus DES COMMENTAIRES, reflétant les sujets de discussion principaux. Ne liste que les mots/expressions.
- MotClé1
- MotClé2
- Expression Clé 3

# RÈGLES IMPORTANTES
- Respecte SCRUPULEUSEMENT le format Markdown demandé avec les titres exacts.
- Ne modifie JAMAIS le texte des commentaires cités dans les sections 3, 4, 5, 6.
- Base les sections 3 à 7 EXCLUSIVEMENT sur le contenu des commentaires fournis.
- Sois objectif et concis.
`, videoTranscript, commentsFormatted) 
	// --- Fin du Prompt ---

	payload := map[string]any{
		"model": ga.model, // Utilise le modèle configuré
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.5, // Configurable si besoin
		"max_tokens":  2048, // Configurable si besoin
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("erreur lors du marshalling du payload Groq: %w", err)
	}

	// Création de la requête avec le contexte
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("erreur lors de la création de la requête Groq: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ga.apiKey) // Utilise la clé API stockée

	// Exécution de la requête via le client HTTP stocké
	resp, err := ga.client.Do(req)
	if err != nil {
		// Vérifier si l'erreur est due à l'annulation du contexte
		if errors.Is(err, context.Canceled) {
			return "", errors.New("appel à l'API Groq annulé")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return "", errors.New("timeout lors de l'appel à l'API Groq")
		}
		return "", fmt.Errorf("erreur lors de l'appel API Groq: %w", err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la lecture de la réponse Groq: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur API Groq (%d): %s", resp.StatusCode, string(respBodyBytes))
	}

	var groqResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(respBodyBytes, &groqResponse); err != nil {
		// Logguer la réponse brute peut aider au débogage si le JSON est invalide
		// log.Printf("DEBUG: Réponse Groq invalide: %s", string(respBodyBytes))
		return "", fmt.Errorf("erreur lors du décodage de la réponse Groq: %w", err)
	}

	if len(groqResponse.Choices) == 0 || groqResponse.Choices[0].Message.Content == "" {
		// log.Printf("DEBUG: Réponse Groq vide reçue: %+v", groqResponse)
		return "", errors.New("aucune réponse ('content') reçue de Groq")
	}

	// Logguer l'usage peut être utile pour le suivi des coûts/limites
	// log.Printf("DEBUG: Usage Groq: %+v", groqResponse.Usage)

	return groqResponse.Choices[0].Message.Content, nil
}