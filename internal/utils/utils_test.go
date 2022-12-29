package utils

import (
	"errors"
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

func TestIsJSON_JSON(t *testing.T) {
	str := `{
		"message": "hello   world",
"id": 123, 
			"foo": "bar"
	}`

	expected := true
	result := IsJSON(str)
	if result != expected {
		t.Fatalf("result should be '%v', it got '%v'", expected, result)
	}
}

func TestIsJSON_String(t *testing.T) {
	str := `
		hello
			world
	`

	expected := false
	result := IsJSON(str)
	if result != expected {
		t.Fatalf("result should be '%v', it got '%v'", expected, result)
	}
}

func TestJsonError(t *testing.T) {
	expected := errors.New("foo")
	e := JsonError{}
	e.Errorf("foo")

	if e.Err.Error() != expected.Error() {
		t.Fatalf("result should be '%v', it got '%v'", expected.Error(), e.Err.Error())
	}
}
