# Integration

<a href="https://godoc.org/github.com/lucasvmiguel/integration"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>

Integration is a Golang tool to run integration tests. Currently, this library only supports an HTTP request and response model.

## Install

To use the integration library, install `Go` and run go get:

```
go get -u github.com/lucasvmiguel/integration
```

## Getting started

The simplest use case is calling an endpoint via http and checking the return of the call. To test that, use the follwing code:

```go
package tests

// a http server must be initiate to execute requests
func init() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	go http.ListenAndServe(":8080", nil)
}

// sample test case
func TestPingEndpoint(t *testing.T) {
	err := Test(TestCase{
		Description: "Testing ping endpoint",
		Request: Request{
			URL:    "http://localhost:8080/ping",
			Method: http.MethodGet,
		},
		ResponseExpected: ResponseExpected{
			StatusCode: http.StatusOK,
			Body:       "pong",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
```

Note: The http server must be started together with the tests

## How to use

See how to use different aspects of the library below.

### Request

An HTTP request will be sent to the your server depending on how it's configure the `Request` property on the `TestCase`. `Request` has many different fields to be configured, see them below:

```go
// Request represents an HTTP request
type Request struct {
	// URL that will be called on the request
	// eg: https://jsonplaceholder.typicode.com/todos
	URL string
	// Method that will be called on the request
	// eg: POST
	Method string
	// Body that will be sent with the request
	// a multiline string is valid
	// eg: { "foo": "bar" }
	Body string
	// Header will be sent with the request
	// eg: content-type=application/json
	Header http.Header
}
```

### Response

An HTTP response will be expected from your server depending on how it's configure the `Response` property on the `TestCase`. If your endpoints sends a different response, the `Test` function will return an `error`. `Response` has many different fields to be configured, see them below:

```go
// ResponseExpected represents an HTTP response
type ResponseExpected struct {
	// StatusCode is the HTTP status code of the response
	StatusCode int
	// Body is the HTTP response body
	Body string
	// IgnoreBodyFields is used to ignore the assertion of some of the body field
	// The syntax used to ignore fields can be found here: https://github.com/tidwall/sjson
	// eg: ["data.transaction_id", "id"]
	IgnoreBodyFields []string
	// Header is the HTTP response headers.
	// Every header set in here will be asserted, others will be ignored.
	// eg: content-type=application/json
	Header http.Header
}
```

### Assertions

There are few different assertion. See them below:

#### SQL

SQL assertion checks if an SQL query returns an expected result. See below how to use it.

```go
func TestHandlerAndAssertSomeRecordsAreInDatabase(t *testing.T) {
  db := connectToDatabase()

	err := Test(TestCase{
		Description: "testing an endpoint and check if records are in the database",
		Request: Request{
			URL:    "http://localhost:8080/ping",
			Method: http.MethodGet,
		},
		ResponseExpected: ResponseExpected{
			StatusCode: http.StatusOK,
			Body:       "pong",
		},
		Assertions: []assertion.Assertion{
			&SQLAssertion{
        DB: db,
        Query: `
          SELECT id, title, description, category_id FROM products
        `,
        ResultExpected: `
          [
            {"category_id":"1","description":"bar1","id":"1","title":"foo1"},
            {"category_id":"1","description":"bar2","id":"2","title":"foo2"}
          ]
        `,
	    },
		},
	})

	if err == nil {
		t.Fatal(err)
	}
}
```

See all available fields when configuring an `SQLAssertion`:

```go
// SQLAssertion asserts a SQL query
type SQLAssertion struct {
	// DB database used to query the data to assert
	DB *sql.DB
	// Query that will run in the database
	Query string
	// ResultExpected expects result in json that will be returned when the query run.
	// A multiline string is valid
	// eg:
	// [{
	// 		"description":"Bar",
	// 		"id":"2",
	// 		"title":"Fooa"
	// 	}]
	ResultExpected string
}
```

#### HTTP

HTTP assertion checks if an HTTP request was sent while your endpoint was being called. See below how to use it.

```go
func TestHandlerWithAnHTTPCAll(t *testing.T) {
	err := Test(TestCase{
		Description: "testing an endpoint that calls another endpoint",
		Request: Request{
			URL:    "http://localhost:8080/ping",
			Method: http.MethodGet,
		},
		ResponseExpected: ResponseExpected{
			StatusCode: http.StatusOK,
			Body:       "pong",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTPAssertion{
				RequestExpected: RequestExpected{
          URL:    "http://localhost:8080/another_ping",
          Method: http.MethodGet,
        },
        ResponseMock: ResponseMock{
          StatusCode: http.StatusOK,
          Body:       "pong from another endpoint",
        },
			},
		},
	})

	if err == nil {
		t.Fatal(err)
	}
}
```

See all available fields when configuring an `HTTPAssertion`:

```go
// HTTPAssertion asserts a http request
type HTTPAssertion struct {
	// RequestExpected will assert if request was made with correct parameters
	RequestExpected RequestExpected
	// ResponseMock mocks a fake response to avoid your test making real http request over the internet
	ResponseMock ResponseMock
}
```

```go
// RequestExpected struct is used to validate if a request was made with the correct parameters
type RequestExpected struct {
	// URL request url that must be called
	URL string
	// Method request method that must be called with
	Method string
	// Header request header that will be asserted with
	// Every header set in here will be asserted, others will be ignored.
	Header http.Header
	// Body request body that must be called with
	// a multiline string is valid
	// eg: { "foo": "bar" }
	Body string
	// IgnoreBodyFields is used to ignore the assertion of some of the body field
	// The syntax used to ignore fields can be found here: https://github.com/tidwall/sjson
	// eg: ["data.transaction_id", "id"]
	IgnoreBodyFields []string
}
```

```go
// ResponseMock struct is used to return a fake http response to your application
type ResponseMock struct {
	// StatusCode is the HTTP status code of the response
	StatusCode int
	// Body is the HTTP response body
	// a multiline string is valid
	// eg: { "foo": "bar" }
	Body string
}
```

## License

You can see this project's license [here](LICENSE).

It's important to mention that this project contains the following libs:

- github.com/jarcoal/httpmock
- github.com/pkg/errors
- github.com/tidwall/gjson
- github.com/tidwall/sjson
- github.com/elgs/gosqljson
