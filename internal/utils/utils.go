package utils

import (
	"regexp"
)

// Trim remove all empty spaces from string
func Trim(str string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(str, "")
}
