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
	// IgnoreBodyFields is used to ignore the assertion of some of the body field
	// The syntax used to ignore fields can be found here: https://github.com/tidwall/sjson
	// eg: ["data.transaction_id", "id"]
	IgnoreBodyFields []string
}

// Response is used to validate if a HTTP response was returned with the correct parameters
type Response struct {
	// StatusCode expected in the HTTP response
	StatusCode int
	// Body expected in the HTTP response
	Body string
	// IgnoreBodyFields is used to ignore the assertion of some of the body field
	// The syntax used to ignore fields can be found here: https://github.com/tidwall/sjson
	// eg: ["data.transaction_id", "id"]
	IgnoreBodyFields []string
	// Header expected in the HTTP response.
	// Every header set in here will be asserted, others will be ignored.
	// eg: content-type=application/json
	Header http.Header
}
