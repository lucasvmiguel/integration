package utils

import (
	"testing"
)

func TestTrim(t *testing.T) {
	str := `
		foo
		bar Test
	
	`

	expected := "foobarTest"
	result := Trim(str)
	if result != expected {
		t.Fatalf("result should be '%s', it got '%s'", expected, result)
	}
}
