package services

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type AIService struct {
	client *openai.Client
}

func NewAIService(apiKey string) *AIService {
	client := openai.NewClient(apiKey)
	return &AIService{client: client}
}

// AnalyzeComments prend une liste de commentaires et retourne un résumé IA (sentiment, questions, mots-clés)
func (ai *AIService) AnalyzeComments(comments []string) (string, error) {
	if len(comments) == 0 {
		return "", fmt.Errorf("aucun commentaire à analyser")
	}

	// Concatène les commentaires pour le prompt
	commentBlock := strings.Join(comments, "\n- ")

	// Prompt personnalisé
	prompt := fmt.Sprintf(`
Voici une liste de commentaires d'une vidéo YouTube. Analyse-les et retourne :
1. Le sentiment global (positif, négatif ou neutre)
2. Les principales questions posées par les viewers
3. Les mots-clés ou thèmes les plus fréquents

Commentaires :
- %s
`, commentBlock)

	// Appel à GPT-4
	resp, err := ai.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-3.5-turbo",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", err
	}

	// Retour du texte généré
	return resp.Choices[0].Message.Content, nil
}
