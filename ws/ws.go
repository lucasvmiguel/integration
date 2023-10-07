package ws

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// Connection is a Websocket connection
type WebsocketConnection struct {
	conn *websocket.Conn
	mux  sync.Mutex
}

// NewWebsocketConnection creates a new Websocket connection
func NewWebsocketConnection(scheme, host, path string, headers http.Header) (*WebsocketConnection, error) {
	if scheme == "" {
		scheme = "ws"
	}

	u := url.URL{Scheme: scheme, Host: host, Path: path}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		return nil, fmt.Errorf("error to connect to the Websocket server (%s): %s", u.String(), err.Error())
	}

	return &WebsocketConnection{
		conn: conn,
	}, nil
}

// ReadMessage reads a message from the Websocket server
func (wc *WebsocketConnection) Read() (int, []byte, error) {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	return wc.conn.ReadMessage()
}

// SetPingHandler sets a handler for ping messages
func (wc *WebsocketConnection) SetPingHandler(handler func(data string) error) {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	wc.conn.SetPingHandler(handler)
}

// SetPongHandler sets a handler for pong messages
func (wc *WebsocketConnection) SetPongHandler(handler func(data string) error) {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	wc.conn.SetPongHandler(handler)
}

// Send sends a message to the Websocket server
// messageType is based on Gorilla's message types
// https://pkg.go.dev/github.com/gorilla/websocket#pkg-constants
func (wc *WebsocketConnection) Send(messageType int, data []byte) error {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	return wc.conn.WriteMessage(messageType, []byte(data))
}

// Close closes the Websocket connection
func (wc *WebsocketConnection) Close() error {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	return wc.conn.Close()
}
