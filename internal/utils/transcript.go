package utils

import (
	"errors"
	"os/exec"
	"strings"
)

func TruncateTextByWords(text string, maxWords int) string {
	words := strings.Fields(text)
	if len(words) > maxWords {
		words = words[:maxWords]
	}
	return strings.Join(words, " ")
}

func GetTranscript(videoID string) (string, error) {
	cmd := exec.Command(".venv/bin/python", "scripts/get_transcript.py", videoID)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	transcript := string(output)
	if strings.HasPrefix(transcript, "ERROR:") {
		return "", errors.New(transcript)
	}

	return TruncateTextByWords(transcript, 1000), nil
}
