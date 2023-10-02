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

	wc := &WebsocketConnection{
		conn: conn,
	}

	wc.conn.SetPingHandler(func(data string) error {
		m := []byte(data)
		return wc.Send(websocket.PongMessage, m)
	})

	wc.conn.SetPongHandler(func(data string) error {
		m := []byte(data)
		return wc.Send(websocket.PingMessage, m)
	})

	return wc, nil
}

// ReadMessage reads a message from the Websocket server
func (wc *WebsocketConnection) Read() (int, []byte, error) {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	return wc.conn.ReadMessage()
}

// Send sends a message to the Websocket server
// messageType is based on Gorilla's message types
// https://pkg.go.dev/github.com/gorilla/websocket#pkg-constants
func (wc *WebsocketConnection) Send(messageType int, data []byte) error {
	// wc.mux.Lock()
	// defer wc.mux.Unlock()

	return wc.conn.WriteMessage(messageType, []byte(data))
}

// Close closes the Websocket connection
func (wc *WebsocketConnection) Close() error {
	wc.mux.Lock()
	defer wc.mux.Unlock()

	return wc.conn.Close()
}
