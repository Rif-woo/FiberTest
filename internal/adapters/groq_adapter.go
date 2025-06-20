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

)

type GroqAdapter interface {
	AnalyzeComments(ctx context.Context, comments []string, videoTranscript string) (string, error)
	SummarizeTranscript(ctx context.Context, transcript string) (string, error)
}

// Structure qui implémente services.GroqAdapter
type groqAdapter struct {
	apiKey string
	client *http.Client // Garde un client HTTP réutilisable
	model  string       // Nom du modèle Groq à utiliser (ex: "llama3-70b-8192")
}

// Constructeur pour groqAdapter
// Prend la clé API et retourne l'INTERFACE services.GroqAdapter
func NewGroqAdapter(apiKey string) GroqAdapter {
	if apiKey == "" {
		// Ne pas retourner d'erreur ici, car main.go vérifie déjà.
		// Si vous voulez une double vérification, retournez (nil, error).
		log.Println("WARN: Tentative de création de GroqAdapter sans clé API.")
	}
	//ancien model :llama3-70b-8192
	//deepseek-r1-distill-llama-70b
	//llama-guard-3-8b
	return &groqAdapter{
		apiKey: apiKey,
		client: &http.Client{Timeout: 90 * time.Second}, // Timeout configurable
		model:  "deepseek-r1-distill-llama-70b",                 // Modèle par défaut, pourrait être configurable
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
Tu es un analyste expert... Ton objectif est d'extraire des informations clés et de **classer chaque commentaire fourni dans la catégorie la plus appropriée** parmi Questions, Critiques, Points Positifs, ou Feedbacks Spécifiques, même si l'appartenance n'est pas parfaite. Utilise la transcription comme contexte.


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
Liste textuellement **toutes** les questions claires posées par les utilisateurs DANS LES COMMENTAIRES. Si aucune question n'est trouvée, écris "Aucune question identifiée.". Ne liste que les questions, pas de phrases introductives.
- [Question 1 textuelle telle qu'écrite par l'utilisateur]
- [Question 2 textuelle...]

## 4. Critiques Négatives
Liste **tous** les commentaires (ou extraits les plus pertinents) exprimant des critiques claires ou un mécontentement DANS LES COMMENTAIRES. Cite le commentaire exact ou l'extrait significatif. Si aucune critique n'est trouvée, écris "Aucune critique négative significative identifiée.".
- "[Auteur:] [Extrait ou commentaire négatif 1]"
- "[Auteur:] [Extrait ou commentaire négatif 2]"

## 5. Points Positifs ou Constructifs
Liste **tous** les commentaires (ou extraits les plus pertinents) exprimant un avis positif marqué, un encouragement, ou une suggestion constructive DANS LES COMMENTAIRES. Cite le commentaire exact ou l'extrait significatif. Si aucun point positif notable n'est trouvé, écris "Aucun commentaire positif ou constructif notable identifié.".
- "[Auteur:] [Extrait ou commentaire positif/constructif 1]"
- "[Auteur:] [Extrait ou commentaire positif/constructif 2]"

## 6. Feedbacks Spécifiques ou Techniques
Liste **tous** les commentaires (ou extraits pertinents) DANS LES COMMENTAIRES contenant des retours d'expérience détaillés, des suggestions techniques précises, des corrections factuelles ou des remarques spécifiques pointues liées au contenu. Cite le commentaire exact ou l'extrait significatif. Si aucun feedback de ce type n'est trouvé, écris "Aucun feedback spécifique ou technique identifié.".
- "[Auteur:] [Extrait ou commentaire de feedback 1]"
- "[Auteur:] [Extrait ou commentaire de feedback 2]"

## 7. Mots-clés et Thèmes Fréquents
Liste les **principaux** mots-clés ou courtes expressions (1-3 mots) les plus fréquents et pertinents issus DES COMMENTAIRES, reflétant les sujets de discussion principaux. Ne liste que les mots/expressions. *(Note: Pas de limite numérique ici non plus, mais "principaux" donne une indication)*
- MotClé1
- MotClé2
- Expression Clé 3

# RÈGLES IMPORTANTES
- **Chaque commentaire fourni doit apparaître dans EXACTEMENT UNE des sections 3, 4, 5 ou 6.** Choisis la catégorie la plus pertinente même si le commentaire est neutre ou ambigu.
- Respecte SCRUPULEUSEMENT le format...
- Ne modifie JAMAIS le texte des commentaires...
- Sois objectif dans la mesure du possible pour la catégorisation forcée.
`, videoTranscript, commentsFormatted)
	// --- Fin du Prompt ---

	payload := map[string]any{
		"model": ga.model, // Utilise le modèle configuré
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.5, // Configurable si besoin
		"max_tokens":  4096, // Configurable si besoin
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



func (ga *groqAdapter) SummarizeTranscript(ctx context.Context, transcript string) (string, error) {
	if ga.apiKey == "" {
		return "", errors.New("GroqAdapter non configuré avec une clé API")
	}
	// Vérifier si la transcription est vide ou indique non disponible
	if transcript == "" || transcript == "Transcription non disponible." {
		log.Println("INFO: Adapter: Transcription vide ou non disponible, aucun résumé généré.")
		// Ce n'est pas une erreur, mais il n'y a rien à résumer.
		return "Résumé non généré (transcription indisponible).", nil
	}

	// --- Définition du Prompt pour le Résumé ---
	prompt := fmt.Sprintf(`
# RÔLE ET OBJECTIF
Tu es un assistant spécialisé dans la synthèse de transcriptions de vidéos YouTube. Ton but est d'extraire les informations clés du contenu parlé dans la vidéo ci-dessous pour fournir un résumé structuré et informatif. Ne fais PAS référence aux commentaires des utilisateurs.

# TRANSCRIPTION À ANALYSER
"""
%s
"""

# INSTRUCTIONS DE SYNTHÈSE ET FORMAT DE SORTIE OBLIGATOIRE
Analyse la transcription fournie et structure ta réponse IMPÉRATIVEMENT avec les sections Markdown suivantes :

## 1. Résumé Global (3-5 phrases)
Décris brièvement le sujet principal de la vidéo et les points essentiels abordés.

## 2. Sujets Clés Abordés
Liste les 3 à 7 thèmes ou sujets principaux discutés en détail dans la vidéo.
- Sujet 1
- Sujet 2
- ...

## 3. Personnes ou Entités Mentionnées
Liste les noms de personnes, d'entreprises, de marques, de produits ou d'autres entités spécifiques nommées dans la vidéo. Si aucune n'est mentionnée, écris "Aucune mention spécifique identifiée.".
- Nom Propre 1
- Marque X
- ...

## 4. Questions Soulevées (par le créateur dans la vidéo)
Liste les questions rhétoriques ou directes posées par le locuteur DANS LA VIDÉO pour engager l'audience ou introduire un sujet. Si aucune n'est posée, écris "Aucune question clé identifiée dans la vidéo.".
- Question 1 posée dans la vidéo ?
- Question 2 ... ?

# RÈGLES IMPORTANTES
- Base-toi EXCLUSIVEMENT sur le contenu de la transcription fournie.
- Sois objectif et concis.
- Respecte SCRUPULEUSEMENT le format Markdown demandé avec les titres exacts.
`, transcript) // Injection de la transcription brute (ou tronquée si nécessaire avant l'appel)
	// --- Fin du Prompt ---

	// Préparation du payload pour l'API Groq
	payload := map[string]any{
		// Utiliser un modèle potentiellement plus petit/rapide si suffisant ?
		"model": ga.model, // ou "llama3-8b-8192" ?
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"temperature": 0.3, // Plus factuel pour le résumé
		// Ajuster max_tokens pour la longueur attendue du résumé ET pour éviter l'erreur 400
		"max_tokens":  768, // Exemple, plus petit que pour l'analyse des commentaires
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("erreur marshalling payload résumé Groq: %w", err)
	}

	// Création et exécution de la requête HTTP
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("erreur création requête résumé Groq: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ga.apiKey)

	log.Printf("INFO: Adapter: Appel API Groq pour résumer la transcription...")
	resp, err := ga.client.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) { return "", errors.New("résumé Groq annulé par contexte") }
		if errors.Is(err, context.DeadlineExceeded) { return "", errors.New("timeout résumé Groq") }
		return "", fmt.Errorf("erreur appel API résumé Groq: %w", err)
	}
	defer resp.Body.Close()

	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erreur lecture réponse résumé Groq: %w", err)
	}

	// Gestion de la réponse et des erreurs API
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Réponse brute erreur API résumé Groq (%d): %s", resp.StatusCode, string(respBodyBytes))
		// Tenter de parser l'erreur Groq pour un message plus clair
		var errorResponse struct { Error struct { Message string `json:"message"`; Type string `json:"type"`; Code string `json:"code"` }}
		if json.Unmarshal(respBodyBytes, &errorResponse) == nil && errorResponse.Error.Message != "" {
            return "", fmt.Errorf("erreur API résumé Groq (%d - %s): %s", resp.StatusCode, errorResponse.Error.Type, errorResponse.Error.Message)
        }
		// Sinon, erreur générique
		return "", fmt.Errorf("erreur API résumé Groq (%d)", resp.StatusCode)
	}

	// Décodage de la réponse JSON de succès
	var groqResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		// Usage non traité ici, mais pourrait l'être
	}

	if err := json.Unmarshal(respBodyBytes, &groqResponse); err != nil {
		log.Printf("ERROR: Impossible de décoder la réponse Groq pour le résumé. Body: %s", string(respBodyBytes))
		return "", fmt.Errorf("erreur décodage réponse succès résumé Groq: %w", err)
	}

	// Vérifier si la réponse contient bien du contenu
	if len(groqResponse.Choices) == 0 || groqResponse.Choices[0].Message.Content == "" {
		log.Printf("WARN: Réponse Groq reçue pour le résumé mais sans contenu. Body: %s", string(respBodyBytes))
		return "", errors.New("aucune réponse ('content') reçue de Groq pour le résumé")
	}

	log.Printf("INFO: Adapter: Résumé de transcription généré avec succès.")
	// Retourne le contenu du résumé généré par Groq
	return groqResponse.Choices[0].Message.Content, nil
} // --- FIN NOUVELLE FONCTION ---