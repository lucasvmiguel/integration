package expect

import "net/http"

// Request struct is used to validate if a HTTP request was made with the correct parameters
type Request struct {
	// URL expected in the HTTP request
	URL string
	// Method expected in the HTTP request
	Method string
	// Header expected in the HTTP request
	// Every header set in here will be asserted, others will be ignored.
	Header http.Header
	// Body expected in the HTTP request.
	// A multiline string is valid.
	// eg: { "foo": "bar" }
	Body string
	// How many times the request is expected to be called
	// default: 1
	Times int
}

// Response is used to validate if a HTTP response was returned with the correct parameters
type Response struct {
	// StatusCode expected in the HTTP response
	StatusCode int
	// Body expected in the HTTP response
	Body string
	// Header expected in the HTTP response.
	// Every header set in here will be asserted, others will be ignored.
	// eg: content-type=application/json
	Header http.Header
}
