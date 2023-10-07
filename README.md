# Integration

![Coverage](https://img.shields.io/badge/Coverage-82.8%25-brightgreen)
[![run tests](https://github.com/lucasvmiguel/integration/actions/workflows/test.yml/badge.svg)](https://github.com/lucasvmiguel/integration/actions/workflows/test.yml)
<a href="https://godoc.org/github.com/lucasvmiguel/integration"><img src="https://img.shields.io/badge/api-reference-blue.svg?style=flat-square" alt="GoDoc"></a>

Integration is a Golang tool to run integration tests.

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

## How to use

See how to use all the available types of test cases below.

### HTTP

A HTTP request can be tested using the `HTTPTestCase` struct. See below how to use it:

#### Example

```go
integration.HTTPTestCase{
	Description: "Example",
	Request: call.Request{
		URL:    "http://localhost:8080/posts/1",
	},
	Response: expect.Response{
		StatusCode: http.StatusOK,
		Body: `{
			"title": "some title",
			"code": "<<PRESENCE>>"
		}`,
	},
	Assertions: []assertion.Assertion{
		&assertion.HTTP{
			Request: expect.Request{
				URL:    "https://jsonplaceholder.typicode.com/posts/1",
				Method: http.MethodGet,
			},
		},
	},
}
```

#### Fields

Fields to configure for `HTTPTestCase` struct

| Field       | Description                                                                            | Example                 | Required? | Default |
| ----------- | -------------------------------------------------------------------------------------- | ----------------------- | --------- | ------- |
| Description | Description describes a test case                                                      | My test                 | false     | -       |
| Request     | Request is what the test case will try to call                                         | call.Request{}          | true      | -       |
| Response    | Response is going to be used to assert if the HTTP endpoint returned what was expected | expect.Response{}       | true      | -       |
| Assertions  | Assertions that will run in test case                                                  | []assertion.Assertion{} | false     | -       |

#### Request

A HTTP request will be sent to the your server depending on how it's configured the `Request` property on the `HTTPTestCase`. `Request` has many different fields to be configured, see them below:

| Field  | Description                                                        | Example                                    | Required? | Default |
| ------ | ------------------------------------------------------------------ | ------------------------------------------ | --------- | ------- |
| URL    | URL that will be called on the request                             | https://jsonplaceholder.typicode.com/todos | true      | -       |
| Method | Method that will be called on the request                          | POST                                       | false     | GET     |
| Body   | Body that will be sent with the request. Multiline string is valid | { "foo": "bar" }                           | false     | -       |
| Header | Header will be sent with the request                               | content-type=application/json              | false     | -       |

#### Response

A HTTP response will be expected from your server depending on how it's configured the `Response` property on the `HTTPTestCase`. If your endpoint sends a different response, the `Test` function will return an `error`. `Response` has many different fields to be configured, see them below:

| Field      | Description                                                                                             | Example                       | Required? | Default |
| ---------- | ------------------------------------------------------------------------------------------------------- | ----------------------------- | --------- | ------- |
| StatusCode | StatusCode expected in the HTTP response                                                                | 200                           | true      | -       |
| Body       | Body expected in the HTTP response                                                                      | hello                         | false     | -       |
| Header     | Header expected in the HTTP response. Every header set in here will be asserted, others will be ignored | content-type=application/json | false     | -       |

You can also ignore a JSON response body field assertion adding the annotation `<<PRESENSE>>`. More info [here](https://github.com/kinbiko/jsonassert)

### GRPC

A GRPC call can be tested using the `GRPCTestCase` struct. See below how to use it:

#### Example

```go
integration.GRPCTestCase{
	Description: "Example",
	Call: call.Call{
		ServiceClient: c,
		Function:      "SayHello",
		Message: &chat.Message{
			Id:      1,
			Body:    "Hello From Client!",
			Comment: "Whatever",
		},
	},
	Output: expect.Output{
		Message: &chat.Message{
			Id:      1,
			Body:    "Hello From the Server!",
			Comment: "<<PRESENCE>>",
		},
	},
	Assertions: []assertion.Assertion{
		&assertion.HTTP{
			Request: expect.Request{
				URL:    "https://jsonplaceholder.typicode.com/posts/1",
			},
		},
	},
}
```

#### Fields

Fields to configure for `GRPCTestCase` struct

| Field       | Description                                                                          | Example                 | Required? | Default |
| ----------- | ------------------------------------------------------------------------------------ | ----------------------- | --------- | ------- |
| Description | Description describes a test case                                                    | My test                 | false     | -       |
| Call        | Call is what the test case will try to call                                          | call.Call{}             | true      | -       |
| Output      | Output is going to be used to assert if the GRPC response returned what was expected | expect.Output{}         | true      | -       |
| Assertions  | Assertions that will run in test case                                                | []assertion.Assertion{} | false     | -       |

#### Call

A GRPC call will be sent to the your GRPC server depending on how it's configured the `Call` property on the `GRPCTestCase`. `Call` has many different fields to be configured, see them below:

| Field         | Description                                 | Example                                              | Required? | Default |
| ------------- | ------------------------------------------- | ---------------------------------------------------- | --------- | ------- |
| ServiceClient | GRPC service client used to call the server | ChatServiceClient                                    | true      | -       |
| Function      | Function that will be called on the request | SayHello                                             | true      | -       |
| Message       | Message that will be sent with the request  | &chat.Message{Id: 1, Body: "Hello From the Server!"} | true      | -       |

#### Output

A GRPC output will be expected from your server depending on how it's configured the `Output` property on the `GRPCTestCase`. If your endpoints send a different response, the `Test` function will return an `error`. `Output` has different fields to be configured, see them below:

| Field   | Description                           | Example                                                                       | Required? | Default |
| ------- | ------------------------------------- | ----------------------------------------------------------------------------- | --------- | ------- |
| Message | Message expected in the GRPC response | &chat.Message{Id: 1, Body: "Hello From the Server!", Comment: "<<PRESENCE>>"} | false     | -       |
| Err     | Error expected in the GRPC response   | status.New(codes.Unavailable, "error message")                                | false     | -       |

You can also ignore a JSON message field assertion adding the annotation `<<PRESENSE>>`. More info [here](https://github.com/kinbiko/jsonassert)

### Websocket

A Websocket call can be tested using the `WebsocketTestCase` struct. See below how to use it:

#### Example

```go
integration.WebsocketTestCase{
	Description: "Example",
	Call: call.Websocket{
		URL:    "localhost:8080",
		Path:   "/websocket",
		Message: `{
			"title": "some title",
			"userId": 1
		}`,
	},
	Receive: &expect.Message{
		Content: `{
			"title": "some title",
			"description": "<<PRESENCE>>",
			"userId": 1,
			"comments": [
				{ "id": 1, "text": "foo" },
				{ "id": 2, "text": "bar" }
			]
		}`,
	},
	Assertions: []assertion.Assertion{
		&assertion.HTTP{
			Request: expect.Request{
				URL:    "https://jsonplaceholder.typicode.com/posts/1",
			},
		},
	},
}
```

#### Fields

| Field       | Description                                                                                                                                                                                 | Example                 | Required? | Default |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------- | --------- | ------- |
| Description | Description describes a test case                                                                                                                                                           | My test                 | false     | -       |
| Call        | Call is the Websocket server the test case will try to connect and send a message                                                                                                           | call.Websocket{}        | true      | -       |
| Receive     | Receive is going to be used to assert if the Websocket server message returned what was expected. This field is optional as a Websocket server can never send a message back to the client. | &expect.Message{}       | false     | nil     |
| Assertions  | Assertions that will run in test case                                                                                                                                                       | []assertion.Assertion{} | false     | -       |

#### Call

A message call will be sent to the your Websocket server depending on how it's configured the `Call` property on the `WebsocketTestCase`. `Websocket` has many different fields to be configured, see them below:

| Field                    | Description                                                                                                                                                                                                                                     | Example                       | Required? | Default                   |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- | --------- | ------------------------- |
| URL                      | URL that will be used to connect to the Websocket server. It won't be required if `Connection` is passed as param                                                                                                                               | my-websocket-server:8080      | true      | -                         |
| Path                     | Path that will be used to connect to the Websocket server                                                                                                                                                                                       | /websocket                    | false     | -                         |
| Scheme                   | Scheme that will be used to connect to the Websocket server                                                                                                                                                                                     | ws or wss                     | false     | ws                        |
| Header                   | Header will be sent with the request                                                                                                                                                                                                            | content-type=application/json | false     | -                         |
| Message                  | Body that will be sent with the request. Multiline string is valid                                                                                                                                                                              | { "foo": "bar" }              | false     | -                         |
| MessageType              | Message type used to send the call. It's based on Gorilla's message types. Reference: https://pkg.go.dev/github.com/gorilla/websocket#pkg-constantstypes                                                                                        | websocket.PingMessage (9)     | false     | websocket.TextMessage (1) |
| Connection               | Connection is the Websocket connection that will be used to make the calls (this field is optional). If you want to reuse a connection, you can set it here. If you set a connection, the `URL`, `Path`, `Header` and `Scheme` will be ignored. | \*ws.WebsocketConnection      | false     | -                         |
| CloseConnectionAfterCall | CloseConnectionAfterCall will close the connection after the call is made                                                                                                                                                                       | true                          | false     | false                     |

#### Receive

A Websocket message can be expected from your Websocket server using the `Receive` property on the `WebsocketTestCase`. The `Receive` property is optional, in case nothing is passed, nothing will be verified. If your endpoint sends a different message, the `Test` function will return an `error`. `Message` has different fields to be configured, see them below:

| Field   | Description                                                             | Example     | Required? | Default   |
| ------- | ----------------------------------------------------------------------- | ----------- | --------- | --------- |
| Content | Content expected in the Websocket message. A multiline string is valid. | My test     | false     | -         |
| Timeout | Timeout is the time to wait for a message to be received.               | time.Second | false     | 5 seconds |

You can also ignore a JSON message field assertion adding the annotation `<<PRESENSE>>`. More info [here](https://github.com/kinbiko/jsonassert)

#### Connection

In case you want to reuse the Websocket connection of a test case, you can call the `.Connection()` function to get the connection. See below how to do it:

```go
initialTestCase := &integration.WebsocketTestCase{
	Description: "First test case with a new connection"
	Call: call.Websocket{
		URL:     "localhost:8080",
		Message:    `ping 1`,
	},
}

err := integration.Test(initialTestCase)
if err != nil {
	t.Fatal(err)
}

// Get the connection established in the first test case
conn := initialTestCase.Connection()

// Use the connection in a second test case
err = integration.Test(&integration.WebsocketTestCase{
	Description: "Second test case with the same connection",
	Call: call.Websocket{
		Connection: conn,
		Message:    `ping 2`,
	},
})
if err != nil {
	t.Fatal(err)
}
```

### Assertions

Assertions are a useful way of validating either a HTTP request or a database change made by your server. Assertions are also used to mock external HTTP APIs responses.

#### SQL

SQL assertion checks if an SQL query returns an expected result. See below how to use `assertion.SQL` for it.

##### Example

```go
integration.HTTPTestCase{
	Description: "Example",
	Request: call.Request{
		URL:    "http://localhost:8080/test",
	},
	Response: expect.Response{
		StatusCode: http.StatusOK,
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
}
```

##### Fields

| Field  | Description                                                            | Example                  | Required? | Default |
| ------ | ---------------------------------------------------------------------- | ------------------------ | --------- | ------- |
| DB     | DB database used to query the data to assert                           | sql.DB{}                 | true      | -       |
| Query  | Query that will run in the database                                    | call.Query{}             | true      | -       |
| Result | Result expects result in json that will be returned when the query run | expect.Result{{"id": 1}} | true      | -       |

##### Query

| Field     | Description                                | Example                     | Required? | Default |
| --------- | ------------------------------------------ | --------------------------- | --------- | ------- |
| Statement | Statement that will be queried             | eg: SELECT \* FROM products | true      | -       |
| Params    | Params that can be passed to the SQL query | []int{1, 2}                 | false     | -       |

#### HTTP

HTTP assertion checks if an HTTP request was sent while your endpoint was being called.
The test will fail if you don't call the endpoints configured on the HTTP assertion. However, if you call multiple times an endpoint and you just have one HTTP assertion configured, the test will pass.

IMPORTANT: `HTTP` assertions uses the library [httpmock](https://github.com/jarcoal/httpmock). The httpmock library works intercepting all HTTP requests and returns a mocked response. But, in order to make it work, you must run your application in the same process as your tests. Otherwise, the assertions will not work. Therefore, HTTP assertions will prevent any real requests to be made.

##### Example

```go
integration.HTTPTestCase{
	Description: "Example",
	Request: call.Request{
		URL:    "http://localhost:8080/test",
	},
	Response: expect.Response{
		StatusCode: http.StatusOK,
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
}
```

##### Fields

| Field    | Description                                                                                  | Example          | Required? | Default |
| -------- | -------------------------------------------------------------------------------------------- | ---------------- | --------- | ------- |
| Request  | Request will assert if request was made with correct parameters                              | expect.Request{} | true      | -       |
| Response | Response mocks a fake response to avoid your test making real http request over the internet | mock.Response{}  | true      | -       |

##### Request

| Field  | Description                                                                                             | Example                                    | Required? | Default |
| ------ | ------------------------------------------------------------------------------------------------------- | ------------------------------------------ | --------- | ------- |
| URL    | URL expected in the HTTP request                                                                        | https://jsonplaceholder.typicode.com/todos | true      | -       |
| Method | Method expected in the HTTP request                                                                     | POST                                       | false     | GET     |
| Body   | Body expected in the HTTP request. Multiline string is valid                                            | { "foo": "bar" }                           | false     | -       |
| Header | Header expected in the HTTP request. Every header set in here will be asserted, others will be ignored. | content-type=application/json              | false     | -       |
| Times  | How many times the request is expected to be called                                                     | 3                                          | false     | 1       |

You can also ignore a JSON response body field assertion adding the annotation `<<PRESENSE>>`. More info [here](https://github.com/kinbiko/jsonassert)

##### Response

| Field      | Description                                                                       | Example          | Required? | Default |
| ---------- | --------------------------------------------------------------------------------- | ---------------- | --------- | ------- |
| StatusCode | StatusCode that will be returned in the mocked HTTP response                      | 404              | false     | 200     |
| Body       | Body that will be returned in the mocked HTTP response. Multiline string is valid | { "foo": "bar" } | false     | -       |

## Contributing

If you want to contribute to this project, please read the [contributing guide](docs/contributing.md).

## License

You can see this project's license [here](LICENSE).

It's important to mention that this project contains the following libs:

- github.com/jarcoal/httpmock
- github.com/kinbiko/jsonassert
- google.golang.org/grpc
- github.com/gorilla/websocket
