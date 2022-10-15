package mock

// Response is used to return a mocked response
type Response struct {
	// StatusCode that will be returned in the mocked HTTP response
	// default value is 200
	StatusCode int
	// Body that will be returned in the mocked HTTP response.
	// A multiline string is valid.
	// eg: { "foo": "bar" }
	Body string
}
