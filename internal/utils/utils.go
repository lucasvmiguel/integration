package utils

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// Error struct to carry json errors
// Reference: "github.com/kinbiko/jsonassert" (Printer)
type JsonError struct {
	Err error
}

// Sets a formatted error
func (e *JsonError) Errorf(msg string, args ...interface{}) {
	e.Err = fmt.Errorf(msg, args...)
}

// Trim remove all empty spaces from string
func Trim(str string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(str, "")
}

// Checks if a string is in JSON format
func IsJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
