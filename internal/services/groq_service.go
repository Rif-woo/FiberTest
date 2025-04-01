package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	// "io/ioutil"
	"net/http"

	// "os"
	"strings"
)

type GroqService struct {
	apiKey string
}

func NewGroqService(apiKey string) *GroqService {
	return &GroqService{apiKey: apiKey}
}

func (gs *GroqService) AnalyzeComments(comments []string, videoTranscript string) (string, error) {
	if len(comments) == 0 {
		return "", fmt.Errorf("aucun commentaire fourni")
	}

	// Prompt
	prompt := fmt.Sprintf(`
Tu es un analyste expert chargé d'examiner les commentaires d'une vidéo YouTube.

Voici la transcription de la vidéo :
"""
%s
"""

Voici les commentaires (avec auteur, et texte exact, sans modification) :
- %s

Ta tâche :
1. Donne un **sentiment global**
2. Rédige un **résumé général**
3. Liste les **questions posées**
4. Dresse une liste des **critiques négatives** envers la vidéo, le créateur ou la communauté
5. Montre les **commentaires positifs ou constructifs**
6. Note les **feedbacks intéressants ou techniques**
7. Extrais les **mots-clés fréquents**

Utilise les commentaires tels qu’ils sont écrits. Ne modifie pas le ton ou le style des auteurs.
`, videoTranscript, strings.Join(comments, "\n- "))

	// Corps de la requête
	payload := map[string]any{
		"model": "llama3-70b-8192",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	body, _ := json.Marshal(payload)

	// Requête HTTP vers Groq
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+gs.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq API error: %s", data)
	}

	var res struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	return res.Choices[0].Message.Content, nil
}
