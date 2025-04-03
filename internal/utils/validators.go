package utils

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func NotBlankInt(value int) bool {
	return value != 0
}

func CheckAge(value int, n int) bool {
	if value > 15 && value < n {
		return true
	}
	return false
}

func NotNull(tab []int) bool {
	return len(tab) != 0
}

func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func CheckValue(value int, n int) bool {
	if value > n || value < 1 {
		return false
	}
	return true
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func IsImage(contentType string) bool {
	imageMIMETypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
	}

	for _, mimeType := range imageMIMETypes {
		if contentType == mimeType {
			return true
		}
	}
	return false
}

func VerifyImg(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Println(err)
		return false
	}

	contentType := http.DetectContentType(buffer)
	if IsImage(contentType) {
		return true
	} else {
		return false
	}
}

func CheckFileSize(filePath string) bool {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	fileSize := fileInfo.Size()

	return fileSize < (20 << 20)
}
