package utils

import (
	"regexp"
	"strings"
)

type ParsedInsight struct {
	Sentiment        string
	Summary          string
	TopComments      []string
	NegativeComments []string
	QuestionComments []string
	FeedbackComments []string
	Keywords         []string
}

func ParseInsightResponse(raw string) *ParsedInsight {
	lines := strings.Split(raw, "\n")
	parsed := &ParsedInsight{}

	var currentBlock string
	var buffer []string

	flush := func() {
		switch currentBlock {
		case "sentiment":
			parsed.Sentiment = strings.Join(buffer, " ")
		case "summary":
			parsed.Summary = strings.Join(buffer, " ")
		case "questions":
			parsed.QuestionComments = cleanBulletList(buffer)
		case "positives":
			parsed.TopComments = cleanBulletList(buffer)
		case "negatives":
			parsed.NegativeComments = cleanBulletList(buffer)
		case "feedback":
			parsed.FeedbackComments = cleanBulletList(buffer)
		case "keywords":
			parsed.Keywords = cleanBulletList(buffer)
		}
		buffer = []string{}
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "**Sentiment global"):
			flush()
			currentBlock = "sentiment"
		case strings.HasPrefix(line, "**Résumé"):
			flush()
			currentBlock = "summary"
		case strings.HasPrefix(line, "**Principales questions"):
			flush()
			currentBlock = "questions"
		case strings.HasPrefix(line, "**Commentaires négatifs"):
			flush()
			currentBlock = "negatives"
		case strings.HasPrefix(line, "**Commentaires positifs"):
			flush()
			currentBlock = "positives"
		case strings.HasPrefix(line, "**Autres feedbacks"):
			flush()
			currentBlock = "feedback"
		case strings.HasPrefix(line, "**Mots-clés"):
			flush()
			currentBlock = "keywords"
		case strings.HasPrefix(line, "* "):
			buffer = append(buffer, strings.TrimPrefix(line, "* "))
		default:
			buffer = append(buffer, line)
		}
	}

	flush()
	return parsed
}

func cleanBulletList(lines []string) []string {
	var cleaned []string
	re := regexp.MustCompile(`\s*\(.*?\)`) // supprime (X mentions)
	for _, l := range lines {
		cleaned = append(cleaned, re.ReplaceAllString(l, ""))
	}
	return cleaned
}
