# Integration

![Coverage](https://img.shields.io/badge/Coverage-87.1%25-brightgreen)
[![run tests](https://github.com/lucasvmiguel/integration/actions/workflows/test.yml/badge.svg)](https://github.com/lucasvmiguel/integration/actions/workflows/test.yml)
<a href="https://godoc.org/github.com/lucasvmiguel/integration"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>

Integration is a Golang tool to run integration tests. Currently, this library supports regular `HTTP` and `GRPC`.

## Install

To use the integration library, install `Go` and run go get:

```
go get -u github.com/lucasvmiguel/integration
```

## Getting started

The simplest use case is calling an endpoint via http and checking the return of the call. To test that, use the follwing code:

```go
package test

import (
	"net/http"
	"testing"

	"github.com/lucasvmiguel/integration"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
)

func init() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	go http.ListenAndServe(":8080", nil)
}

func TestPingEndpoint(t *testing.T) {
	err := integration.Test(&integration.HTTPTestCase{
		Description: "Testing ping endpoint",
		Request: call.Request{
			URL:    "http://localhost:8080/ping",
			Method: http.MethodGet,
		},
		Response: expect.Response{
			StatusCode: http.StatusOK,
			Body:       "pong",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
```

### Examples

You can check more examples below:
- [HTTP](http_test.go)
- [GRPC](grpc_test.go)


**Note: The http/grpc server must be started together with the tests**

## How to use

See how to use different aspects of the library below.

### HTTP

#### Request

An HTTP request will be sent to the your server depending on how it's configure the `Request` property on the `HTTPTestCase`. `Request` has many different fields to be configured, see them below:

```go
// Request sets up how a HTTP request will be called
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

#### Response

An HTTP response will be expected from your server depending on how it's configure the `Response` property on the `HTTPTestCase`. If your endpoint sends a different response, the `Test` function will return an `error`. `Response` has many different fields to be configured, see them below:

```go
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
```

You can also ignore request body field assertion adding the annotation `<<PRESENSE>>`. Check the example below:

```go
&integration.HTTPTestCase{
	Description: "Test Ignored field",
	Request: call.Request{
		URL:    "http://localhost:8080/test",
		Method: goHTTP.MethodPost
	},
	Response: expect.Response{
		StatusCode: goHTTP.StatusCreated,
		Body: `{
			"title": "some title",
			"code": "<<PRESENCE>>"
		}`,
	},
}
```

Reference: https://github.com/kinbiko/jsonassert

### GRPC

#### Call

An GRPC call will be sent to the your server depending on how it's configure the `Call` property on the `GRPCTestCase`. `Call` has many different fields to be configured, see them below:

```go
// Call sets up how a GRPC request will be called
type Call struct {
	// GRPC service client used to call the server
	// eg: ChatServiceClient
	ServiceClient interface{}

	// Function that will be called on the request
	// eg: SayHello
	Function string

	// Message that will be sent with the request
	// Eg: &chat.Message{Id: 1, Body: "Hello From the Server!"}
	Message interface{}
}
```

#### Output

An GRPC output will be expected from your server depending on how it's configure the `Output` property on the `GRPCTestCase`. If your endpoints send a different response, the `Test` function will return an `error`. `Output` has different fields to be configured, see them below:

```go
// Output is used to validate if a GRPC response was returned with the correct parameters
type Output struct {
	// Message expected in the GRPC response
	// Eg: &chat.Message{Id: 1, Body: "Hello From the Server!", Comment: "<<PRESENCE>>"}
	Message interface{}

	// Error expected in the GRPC response
	// Eg: status.New(codes.Unavailable, "error message"),
	Err *status.Status
}
```

You can also ignore request body field assertion adding the annotation `<<PRESENSE>>`.

### Assertions

There are few different assertion that can be made. Assertions work for both regular `HTTP` and `GRPC`.
See them below:

#### SQL

SQL assertion checks if an SQL query returns an expected result. See below how to use it.

```go
func TestEndpoint(t *testing.T) {
	db, _ := connectToDatabase()
	
	err := integration.Test(&integration.HTTPTestCase{
		Description: "Test Endpoint",
		Request: call.Request{
			URL:    "http://localhost:8080/test",
			Method: http.MethodGet,
		},
		Response: expect.Response{
			StatusCode: http.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.SQL{
				DB: db,
				Query: call.Query{
					Statement: `
					SELECT id, title, description, category_id FROM products
					`,
				},
				Result: expect.Result{
					{"id": 1, "title": "foo1", "description": "bar1", "category_id": 1},
					{"id": 2, "title": "foo2", "description": "bar2", "category_id": 1},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
```

See all available fields when configuring an `SQLAssertion`:

```go
// SQL asserts a SQL query
type SQL struct {
	// DB database used to query the data to assert
	DB *sql.DB
	// Query that will run in the database
	Query call.Query
	// ResultExpected expects result in json that will be returned when the query run.
	Result expect.Result
}
```

```go
// Query sets up how a SQL query will be called
type Query struct {
	// Statement that will be queried.
	// eg: SELECT * FROM products
	Statement string

	// Params that can be passed to the SQL query
	Params []any
}
```

```go
// Result is used to validate if a SQL query was returned with the correct items and fields
type Result []map[string]any
```

#### HTTP

HTTP assertion checks if an HTTP request was sent while your endpoint was being called.
The test will fail if you don't call the endpoints configured on the HTTP assertion. However, if you call multiple times an endpoint and you just have one HTTP assertion configured, the test will pass.

See below how to use it:

```go
func TestEndpoint(t *testing.T) {
	err := integration.Test(&integration.HTTPTestCase{
		Description: "Test Endpoint",
		Request: call.Request{
			URL:    "http://localhost:8080/test",
			Method: http.MethodGet,
		},
		Response: expect.Response{
			StatusCode: http.StatusOK,
			Body:       "hello",
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
					URL:    "https://jsonplaceholder.typicode.com/posts",
					Method: http.MethodPost,
					Body: `{
						"title": "foo",
						"body": "bar",
						"userId": "<<PRESENCE>>"
					}`,
				},
				Response: mock.Response{
					StatusCode: http.StatusOK,
					Body: `{
						"id": 1,
						"title": "foo bar"
					}`,
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
```

See all available fields when configuring an `HTTPAssertion`:

```go
// HTTP asserts a http request
type HTTP struct {
	// Request will assert if request was made with correct parameters
	Request expect.Request
	// Response mocks a fake response to avoid your test making real http request over the internet
	Response mock.Response
}
```

```go
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
}
```

You can also ignore request body field assertion adding the annotation `<<PRESENSE>>`

```go
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
```

## Roadmap

Feel free to create issues for features or fixes.

https://github.com/lucasvmiguel/integration/issues

## License

You can see this project's license [here](LICENSE).

It's important to mention that this project contains the following libs:

- github.com/jarcoal/httpmock
- github.com/pkg/errors
- github.com/kinbiko/jsonassert
- google.golang.org/grpc
