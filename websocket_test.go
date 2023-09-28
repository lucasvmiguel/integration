package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
)

func jsonHandler(w http.ResponseWriter, req *http.Request) {
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
			body := struct {
				Title  string `json:"title"`
				UserID int    `json:"userId"`
			}{}
			err := json.NewDecoder(bytes.NewBuffer(messageReceived)).Decode(&body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				break
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
				break
			}

			err = c.WriteMessage(mt, messageSent)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			break
		}
	}
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

	for {
		mt, messageReceived, err := c.ReadMessage()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			break
		}

		if messageReceived != nil {
			body := struct {
				Title  string `json:"title"`
				UserID int    `json:"userId"`
			}{}
			err := json.NewDecoder(bytes.NewBuffer(messageReceived)).Decode(&body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				break
			}

			messageSent := `
			foo
			 bar`

			err = c.WriteMessage(mt, []byte(messageSent))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			break
		}
	}
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
			err = c.WriteMessage(mt, []byte(""))
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

	for {
		_, messageReceived, err := c.ReadMessage()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			break
		}

		if messageReceived != nil {
			break
		}
	}
}

func pongHandler(w http.ResponseWriter, req *http.Request) {
	var upgrader = websocket.Upgrader{}

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer c.Close()

	c.SetPongHandler(func(data string) error {
		return c.WriteMessage(websocket.PingMessage, []byte(data))
	})

	for {
		_, messageReceived, err := c.ReadMessage()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			break
		}

		if messageReceived != nil {
			break
		}
	}
}

func init() {
	http.HandleFunc("/handler-json", jsonHandler)
	http.HandleFunc("/handler-string", stringHandler)
	http.HandleFunc("/infinite-handler", infiniteHandler)
	http.HandleFunc("/ping-handler", pingHandler)
	http.HandleFunc("/pong-handler", pongHandler)

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
		Receive: expect.Message{
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
	u := url.URL{Scheme: "ws", Host: "localhost:8090", Path: "/handler-json"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
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
		Receive: expect.Message{
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
			Scheme:     call.WebsocketSchemeWSS,
			URL:        "ignored",
			Message:    `{}`,
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
		Receive: expect.Message{
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
		Receive: expect.Message{
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

	if err == nil {
		t.Fatal("it should return an error due to an invalid method")
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
		Receive: expect.Message{
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

	if err == nil {
		t.Fatal("it should return an error due to an invalid method")
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
		Receive: expect.Message{
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

func TestWebsocket_Pong(t *testing.T) {
	err := Test(&WebsocketTestCase{
		Description: "TestWebsocket_Pong",
		Call: call.Websocket{
			Scheme:      call.WebsocketSchemeWS,
			URL:         fmt.Sprintf("localhost:%d", 8090),
			Path:        "/pong-handler",
			MessageType: websocket.PongMessage,
			Message: `{
				"title": "some title",
				"userId": 1
			}`,
		},
		Receive: expect.Message{
			Content: `{
				"title": "some title",
				"userId": 1
			}`,
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
