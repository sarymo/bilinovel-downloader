package utils

import (
	"regexp"
	"strings"
)

func CleanDirName(input string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	cleaned := re.ReplaceAllString(input, "_")

	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}
