package utils

import (
	// "regexp"
	"log"
	"sort"
	"strings"
	// "log" // Ajout pour le débogage
)

type ParsedInsight struct {
	Sentiment        string   `json:"Sentiment"` // Ajouter des tags JSON est une bonne pratique
	Summary          string   `json:"Summary"`
	QuestionComments []string `json:"QuestionComments"`
	NegativeComments []string `json:"NegativeComments"`
	TopComments      []string `json:"TopComments"`       // Renommé pour correspondre à l'usage précédent ? Ou à vérifier. Prompt = "Positifs ou Constructifs"
	FeedbackComments []string `json:"FeedbackComments"`
	Keywords         []string `json:"Keywords"`
}

func ParseInsightResponse(raw string) *ParsedInsight {
	// Log d'entrée pour voir ce que l'on reçoit VRAIMENT de Groq
	// log.Printf("DEBUG: ParseInsightResponse: Raw input:\n---\n%s\n---\n", raw)

	lines := strings.Split(raw, "\n")
	parsed := &ParsedInsight{
		// Initialiser les slices pour éviter les `null` en JSON si vides
		QuestionComments: []string{},
		NegativeComments: []string{},
		TopComments:      []string{},
		FeedbackComments: []string{},
		Keywords:         []string{},
	}

	var currentBlock string
	var buffer []string

	flush := func() {
		// Log pour voir ce qui est flushé et où
		// log.Printf("DEBUG: Flushing block '%s' with buffer: %v", currentBlock, buffer)

		content := strings.TrimSpace(strings.Join(buffer, "\n")) // Join avec \n pour les paragraphes, puis trim
		listItems := cleanBulletList(buffer) // Nettoie les items pour les listes

		switch currentBlock {
		case "sentiment":
			// Gérer le cas où le texte est sur la même ligne que le titre (peu probable avec ##) ou la suivante
			if strings.HasPrefix(content, "## 1. Sentiment Général") {
				content = strings.TrimPrefix(content, "## 1. Sentiment Général")
			}
			parsed.Sentiment = strings.TrimSpace(content)
		case "summary":
			if strings.HasPrefix(content, "## 2. Résumé Général des Commentaires") {
				content = strings.TrimPrefix(content, "## 2. Résumé Général des Commentaires")
			}
			parsed.Summary = strings.TrimSpace(content)
		case "questions":
            // Si Groq répond "Aucune question identifiée.", la liste sera vide.
            if len(listItems) > 0 && listItems[0] != "Aucune question identifiée." {
			    parsed.QuestionComments = listItems
            }
		case "negatives":
             if len(listItems) > 0 && listItems[0] != "Aucune critique négative significative identifiée." {
			    parsed.NegativeComments = listItems
             }
		case "positives": // Doit correspondre au nom du champ struct: TopComments ou PositiveComments? J'utilise TopComments basé sur l'ancienne structure JSON
             if len(listItems) > 0 && listItems[0] != "Aucun commentaire positif ou constructif notable identifié." {
			    parsed.TopComments = listItems
             }
		case "feedback":
             if len(listItems) > 0 && listItems[0] != "Aucun feedback spécifique ou technique identifié." {
			    parsed.FeedbackComments = listItems
             }
		case "keywords":
			// Les mots-clés sont souvent une liste simple
			parsed.Keywords = listItems
		}
		buffer = []string{} // Réinitialiser le buffer après flush
	}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}

		newBlockDetected := false
		switch {
		// Vérifier les en-têtes EXACTS du prompt
		case strings.HasPrefix(trimmedLine, "## 1. Sentiment Général"):
			flush()
			currentBlock = "sentiment"
			// Ajouter la ligne au buffer si le contenu n'est pas sur la même ligne (style Markdown)
            buffer = append(buffer, strings.TrimPrefix(trimmedLine, "## 1. Sentiment Général"))
			newBlockDetected = true
		case strings.HasPrefix(trimmedLine, "## 2. Résumé Général des Commentaires"):
			flush()
			currentBlock = "summary"
            buffer = append(buffer, strings.TrimPrefix(trimmedLine, "## 2. Résumé Général des Commentaires"))
			newBlockDetected = true
		case strings.HasPrefix(trimmedLine, "## 3. Questions Posées"):
			flush()
			currentBlock = "questions"
			newBlockDetected = true
             // Ne pas ajouter le titre lui-même au buffer des listes
		case strings.HasPrefix(trimmedLine, "## 4. Critiques Négatives"):
			flush()
			currentBlock = "negatives"
			newBlockDetected = true
		case strings.HasPrefix(trimmedLine, "## 5. Points Positifs ou Constructifs"): // Le nom du bloc doit correspondre à la clé dans le switch `flush`
			flush()
			currentBlock = "positives" // --> Mappe vers TopComments dans le switch
			newBlockDetected = true
		case strings.HasPrefix(trimmedLine, "## 6. Feedbacks Spécifiques ou Techniques"):
			flush()
			currentBlock = "feedback"
			newBlockDetected = true
		case strings.HasPrefix(trimmedLine, "## 7. Mots-clés et Thèmes Fréquents"):
			flush()
			currentBlock = "keywords"
			newBlockDetected = true
		}

		// Si ce n'est pas un nouveau bloc, ajouter la ligne au buffer
		// Sauf si c'est le titre lui-même qu'on vient d'ajouter
		if !newBlockDetected {
             // Vérifier si c'est un item de liste (commence par "- ")
			 if strings.HasPrefix(trimmedLine, "- ") {
                 // Ajouter seulement le contenu de l'item, sans le "- "
				 buffer = append(buffer, strings.TrimSpace(strings.TrimPrefix(trimmedLine, "- ")))
			 } else if currentBlock == "sentiment" || currentBlock == "summary" {
                 // Pour les blocs texte (sentiment, summary), ajouter la ligne entière
                 buffer = append(buffer, trimmedLine)
             }
		}
	}

	flush() // Appeler flush une dernière fois pour la dernière section

	// Log de sortie pour voir le résultat du parsing
	// log.Printf("DEBUG: ParseInsightResponse: Parsed result: %+v", parsed)

	return parsed
}

// Modifié pour gérer les cas "Aucune..." et nettoyer le préfixe "-"
func cleanBulletList(lines []string) []string {
	var cleaned []string
	// re := regexp.MustCompile(`\s*\(.*?\)`) // Regex pour supprimer "(X mentions)" - peut-être pas nécessaire/utile

	// Premier élément peut être le reste du titre ou une phrase comme "Aucune..."
	if len(lines) == 0 {
		return cleaned
	}

    // Supprimer les lignes vides potentielles ajoutées au buffer
    var nonEmptyLines []string
    for _, l := range lines {
        if strings.TrimSpace(l) != "" {
            nonEmptyLines = append(nonEmptyLines, strings.TrimSpace(l))
        }
    }
    lines = nonEmptyLines


	// Gérer explicitement les cas "Aucune..."
	if len(lines) == 1 {
		lineLower := strings.ToLower(lines[0])
		if strings.HasPrefix(lineLower, "aucune") || strings.HasPrefix(lineLower, "aucun") {
            // Retourner la phrase telle quelle ou une liste vide ?
            // Ici on retourne la phrase pour info, mais le code appelant gère déjà ça
			// return []string{lines[0]}
            return []string{} // Préférable: une liste vide si "Aucun..."
		}
	}


	for _, l := range lines {
        // Nettoyage plus simple : juste supprimer les espaces superflus
		cleanedLine := strings.TrimSpace(l)
		// cleanedLine = re.ReplaceAllString(cleanedLine, "") // Appliquer regex si besoin
		if cleanedLine != "" {
			cleaned = append(cleaned, cleanedLine)
		}
	}
	return cleaned
}

type keywordFrequency struct {
	Keyword string
	Count   int
}

// mergeParsedInsights combine les résultats de plusieurs analyses partielles de manière plus intelligente.
func MergeParsedInsights(partials []*ParsedInsight) *ParsedInsight {
	totalPartials := len(partials)
	if totalPartials == 0 {
		log.Println("WARN: mergeParsedInsights appelé avec une liste vide de partiels.")
		return &ParsedInsight{} // Retourne une structure vide
	}

	log.Printf("INFO: Début de la fusion de %d analyses partielles.", totalPartials)

	// --- Agrégation des données brutes ---
	sentimentCounts := make(map[string]int)
	allSummaries := []string{}
	allTopComments := []string{}
	allNegativeComments := []string{}
	allQuestionComments := []string{}
	allFeedbackComments := []string{}
	keywordsCount := make(map[string]int) // map[lowercase_keyword]count

	for i, p := range partials {
		if p == nil {
            log.Printf("WARN: Partiel #%d est nil, ignoré.", i+1)
            continue // Ignore les partiels nil
        }

		// Compter les sentiments (ignorer si vide)
        trimmedSentiment := strings.TrimSpace(p.Sentiment)
		if trimmedSentiment != "" {
			sentimentCounts[trimmedSentiment]++
		}

		// Collecter les résumés (ignorer si vide)
        trimmedSummary := strings.TrimSpace(p.Summary)
		if trimmedSummary != "" {
			allSummaries = append(allSummaries, trimmedSummary)
		}

		// Collecter tous les éléments des listes (ignorer les chaînes vides)
		for _, item := range p.TopComments { if strings.TrimSpace(item) != "" { allTopComments = append(allTopComments, item) } }
		for _, item := range p.NegativeComments { if strings.TrimSpace(item) != "" { allNegativeComments = append(allNegativeComments, item) } }
		for _, item := range p.QuestionComments { if strings.TrimSpace(item) != "" { allQuestionComments = append(allQuestionComments, item) } }
		for _, item := range p.FeedbackComments { if strings.TrimSpace(item) != "" { allFeedbackComments = append(allFeedbackComments, item) } }

		// Compter les mots-clés (en minuscule, ignorer si vide)
		for _, kw := range p.Keywords {
            trimmedKw := strings.ToLower(strings.TrimSpace(kw))
			if trimmedKw != "" {
				keywordsCount[trimmedKw]++
			}
		}
	}

    // --- Traitement et Synthèse ---

	// 1. Déterminer le Sentiment Global (Majorité)
	finalSentiment := "Indéterminé" // Valeur par défaut
	maxCount := 0
	if len(sentimentCounts) > 0 {
		for sentiment, count := range sentimentCounts {
			if count > maxCount {
				maxCount = count
				finalSentiment = sentiment
			} else if count == maxCount {
				// Gérer les égalités (ex: si Positif et Négatif ont le même compte max)
                // On pourrait choisir "Partagé" ou une autre logique. Pour l'instant, on garde le premier trouvé.
                // Ou on peut explicitement définir une priorité ou un état mixte.
                finalSentiment = "Partagé / Mixte" // Exemple pour gérer une égalité simple
			}
		}
        log.Printf("INFO: Fusion: Sentiment final déterminé: %s (basé sur %v)", finalSentiment, sentimentCounts)
	} else {
        log.Println("WARN: Fusion: Aucun sentiment trouvé dans les partiels.")
    }


	// 2. Déterminer le Résumé Global (Stratégie simple: prendre le premier non vide)
	// TODO: Amélioration possible -> Méta-résumé par IA ou concaténation intelligente.
	finalSummary := "Impossible de générer un résumé global." // Défaut
	if len(allSummaries) > 0 {
		finalSummary = allSummaries[0] // Prend le résumé du premier chunk qui en avait un
        log.Printf("INFO: Fusion: Résumé final basé sur le premier partiel non vide.")
	} else {
        log.Println("WARN: Fusion: Aucun résumé trouvé dans les partiels.")
    }


    // 3. Dédoublonner et Limiter les Listes d'Exemples
    limitPerList := 10 // Limite max d'éléments à retourner par liste (configurable !)
    deduplicateAndLimit := func(items []string, limit int) []string {
        seen := make(map[string]bool)
        result := []string{}
        for _, item := range items {
            if !seen[item] {
                seen[item] = true
                result = append(result, item)
                if len(result) >= limit {
                    break
                }
            }
        }
        return result
    }

    finalTopComments := deduplicateAndLimit(allTopComments, limitPerList)
    finalNegativeComments := deduplicateAndLimit(allNegativeComments, limitPerList)
    finalQuestionComments := deduplicateAndLimit(allQuestionComments, limitPerList)
    finalFeedbackComments := deduplicateAndLimit(allFeedbackComments, limitPerList)
    log.Printf("INFO: Fusion: Listes dédoublonnées et limitées à %d éléments.", limitPerList)


    // 4. Classer les Mots-clés par Fréquence et Limiter
    rankedKeywords := []keywordFrequency{}
    for kw, count := range keywordsCount {
        rankedKeywords = append(rankedKeywords, keywordFrequency{Keyword: kw, Count: count})
    }
    // Trier par fréquence (décroissant)
    sort.Slice(rankedKeywords, func(i, j int) bool {
        return rankedKeywords[i].Count > rankedKeywords[j].Count
    })

    finalKeywords := []string{}
    limitKeywords := 15 // Limite max de mots-clés à retourner (configurable !)
    for i, kwf := range rankedKeywords {
        if i >= limitKeywords {
            break
        }
        finalKeywords = append(finalKeywords, kwf.Keyword) // Prend juste le mot, pas le compte
    }
    log.Printf("INFO: Fusion: %d mots-clés finaux sélectionnés (Top %d par fréquence).", len(finalKeywords), limitKeywords)


	// --- Construction de la Structure Fusionnée Finale ---
	merged := &ParsedInsight{
		Sentiment:        finalSentiment,
		Summary:          finalSummary,
		TopComments:      finalTopComments,
		NegativeComments: finalNegativeComments,
		QuestionComments: finalQuestionComments,
		FeedbackComments: finalFeedbackComments,
		Keywords:         finalKeywords,
	}

	log.Printf("INFO: Fusion terminée.")
	return merged
}