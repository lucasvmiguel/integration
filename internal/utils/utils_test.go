package utils

import (
	"testing"
)

func TestTrim(t *testing.T) {
	str := `
		foo
		bar Test
	
	`

	expected := "foobar Test"
	result := Trim(str)
	if result != expected {
		t.Fatalf("result should be '%s', it got '%s'", expected, result)
	}
}

func TestTrim_JSON(t *testing.T) {
	str := `{
		"message": "hello   world",
"id": 123, 
			"foo": "bar"
	}`

	expected := `{"message": "hello   world","id": 123, "foo": "bar"}`
	result := Trim(str)
	if result != expected {
		t.Fatalf("result should be '%s', it got '%s'", expected, result)
	}
}
