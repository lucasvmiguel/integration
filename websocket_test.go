package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/websocket"
	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/ws"
)

type FakeReqBody struct {
	Title  string `json:"title"`
	UserID int    `json:"userId"`
}

func jsonHandler(w http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer c.Close()

	mt, messageReceived, err := c.ReadMessage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	body := FakeReqBody{}
	err = json.NewDecoder(bytes.NewBuffer(messageReceived)).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	messageSent := []byte(`{
			"title": "some title",
			"description": "some description",
			"userId": 1,
			"comments": [
				{ "id": 1, "text": "foo" },
				{ "id": 2, "text": "bar" }
			]
		}`)

	_, err = http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = c.WriteMessage(mt, messageSent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var quit = make(chan struct{})
	<-quit
}

func stringHandler(w http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{}

	if req.Header.Get("foo") != "bar" {
		http.Error(w, "invalid header", http.StatusInternalServerError)
		return
	}

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer c.Close()

	mt, messageReceived, err := c.ReadMessage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	body := FakeReqBody{}
	err = json.NewDecoder(bytes.NewBuffer(messageReceived)).Decode(&body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	messageSent := `
			foo
			 bar`

	err = c.WriteMessage(mt, []byte(messageSent))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var quit = make(chan struct{})
	<-quit
}

func stringHandlerWithoutReply(w http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{}

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer c.Close()

	_, err = http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, _, err = c.ReadMessage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var quit = make(chan struct{})
	<-quit
}

func infiniteHandler(w http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{}

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer c.Close()

	for {
		mt, messageReceived, err := c.ReadMessage()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			break
		}

		if messageReceived != nil {
			err = c.WriteMessage(mt, messageReceived)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				break
			}
		}
	}
}

func pingHandler(w http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{}

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer c.Close()

	_, err = http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.SetPingHandler(func(data string) error {
		return c.WriteMessage(websocket.PongMessage, []byte(data))
	})

	// for some reason, required to read the ping message
	go func() { spew.Dump(c.ReadMessage()) }()

	var quit = make(chan struct{})
	<-quit
}

func init() {
	http.HandleFunc("/handler-json", jsonHandler)
	http.HandleFunc("/handler-string", stringHandler)
	http.HandleFunc("/infinite-handler", infiniteHandler)
	http.HandleFunc("/ping-handler", pingHandler)
	http.HandleFunc("/handler-string-without-reply", stringHandlerWithoutReply)

	go http.ListenAndServe(":8090", nil)
}

func TestWebsocket_SuccessJSON(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_SuccessJSON",
		Call: call.Websocket{
			Scheme: call.WebsocketSchemeWS,
			URL:    fmt.Sprintf("localhost:%d", 8090),
			Path:   "/handler-json",
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
					Method: http.MethodGet,
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestWebsocket_SuccessWithConnectionAlreadyCreated(t *testing.T) {
	conn, err := ws.NewWebsocketConnection("ws", "localhost:8090", "/handler-json", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = Test(&WebsocketTestCase{
		Description: "TestWebsocket_SuccessWithConnectionAlreadyCreated",
		Call: call.Websocket{
			Connection: conn,
			Scheme:     call.WebsocketSchemeWSS,
			URL:        "ignored",
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
					Method: http.MethodGet,
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestWebsocket_EmptyMessage(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_EmptyMessage",
		Call: call.Websocket{
			Scheme: call.WebsocketSchemeWS,
			URL:    fmt.Sprintf("localhost:%d", 8090),
			Path:   "/infinite-handler",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestWebsocket_EmptyReturn(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_EmptyReturn",
		Call: call.Websocket{
			Scheme:  call.WebsocketSchemeWS,
			URL:     fmt.Sprintf("localhost:%d", 8090),
			Path:    "/infinite-handler",
			Message: `{}`,
		},
		Receive: &expect.Message{
			Content: `{}`,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestWebsocket_SuccessWithConnectionWithReturnedConnection(t *testing.T) {
	initialTestCase := &WebsocketTestCase{
		Description: "TestWebsocket_SuccessWithConnectionWithReturnedConnection_1",
		Call: call.Websocket{
			Scheme:  call.WebsocketSchemeWS,
			URL:     fmt.Sprintf("localhost:%d", 8090),
			Path:    "/infinite-handler",
			Message: `{}`,
		},
		Receive: &expect.Message{
			Content: `{}`,
		},
	}

	err := Test(initialTestCase)
	if err != nil {
		t.Fatal(err)
	}

	conn := initialTestCase.Connection()

	err = Test(&WebsocketTestCase{
		Description: "TestWebsocket_SuccessWithConnectionWithReturnedConnection_2",
		Call: call.Websocket{
			Connection: conn,
			Message:    `foo`,
		},
		Receive: &expect.Message{
			Content: `foo`,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWebsocket_SuccessString(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_SuccessString",
		Call: call.Websocket{
			Scheme: call.WebsocketSchemeWS,
			URL:    fmt.Sprintf("localhost:%d", 8090),
			Path:   "/handler-string",
			Header: http.Header{
				"foo": []string{"bar"},
			},
			Message: `{
				"title": "some title",
				"userId": 1
			}`,
		},
		Receive: &expect.Message{
			Content: `
			foo
			 bar`,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestWebsocket_InvalidURL(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_InvalidURL",
		Call: call.Websocket{
			Scheme: call.WebsocketSchemeWS,
			URL:    fmt.Sprintf("invalid:%d", 8090),
			Path:   "/handler-json",
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
			Timeout: 10 * time.Second,
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: http.MethodGet,
				},
			},
		},
	})

	if err == nil || !strings.Contains(err.Error(), "error to connect to the Websocket server") {
		t.Fatal("it should return an error due to an invalid path")
	}
}

func TestWebsocket_InvalidPath(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_InvalidPath",
		Call: call.Websocket{
			Scheme: call.WebsocketSchemeWS,
			URL:    fmt.Sprintf("localhost:%d", 8090),
			Path:   "/invalid",
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
					Method: http.MethodGet,
				},
			},
		},
	})

	if err == nil || !strings.Contains(err.Error(), "error to connect to the Websocket server") {
		t.Fatal("it should return an error due to an invalid path")
	}
}

func TestWebsocket_Ping(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_Ping",
		Call: call.Websocket{
			Scheme:      call.WebsocketSchemeWS,
			URL:         fmt.Sprintf("localhost:%d", 8090),
			Path:        "/ping-handler",
			MessageType: websocket.PingMessage,
			Message: `{
				"title": "some title",
				"userId": 1
			}`,
		},
		Receive: &expect.Message{
			Content: `{
				"title": "some title",
				"userId": 1
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
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestWebsocket_SuccessButWithoutReply(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_SuccessButWithoutReply",
		Call: call.Websocket{
			Scheme:  call.WebsocketSchemeWS,
			URL:     fmt.Sprintf("localhost:%d", 8090),
			Path:    "/handler-string-without-reply",
			Message: `foo`,
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: http.MethodGet,
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestWebsocket_SuccessCloseConnection(t *testing.T) {
	testCase := &WebsocketTestCase{
		Description: "TestWebsocket_SuccessCloseConnection",
		Call: call.Websocket{
			Scheme: call.WebsocketSchemeWS,
			URL:    fmt.Sprintf("localhost:%d", 8090),
			Path:   "/handler-string",
			Header: http.Header{
				"foo": []string{"bar"},
			},
			Message: `{
				"title": "some title",
				"userId": 1
			}`,
			CloseConnectionAfterCall: true,
		},
		Receive: &expect.Message{
			Content: `
			foo
			 bar`,
		},
	}
	err := Test(testCase)
	if err != nil {
		t.Fatal(err)
	}

	conn := testCase.Connection()

	err = conn.Send(websocket.TextMessage, []byte("foo"))
	if err == nil || !strings.Contains(err.Error(), "use of closed network connection") {
		t.Fatal("it should return an error due to a closed connection")
	}
}
