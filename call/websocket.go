package call

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WebsocketScheme string

const (
	WebsocketSchemeWS  WebsocketScheme = "ws"
	WebsocketSchemeWSS WebsocketScheme = "wss"
)

// Message sets up how a Websocket message will be sent
type Websocket struct {
	// URL that will be used to connect to the Websocket server.
	// eg: my-websocket-server:8080
	URL string

	// Path that will be used to connect to the Websocket server.
	// eg: /websocket
	Path string

	// Scheme that will be used to connect to the Websocket server.
	// if nothing is set, the default will be `ws`.
	// eg: ws or wss
	Scheme WebsocketScheme

	// Header will be used to connect to the Websocket server.
	// eg: content-type=application/json
	Header http.Header

	// Connection is the Websocket connection that will be used to make the calls (this field is optional).
	// If you want to reuse a connection, you can set it here.
	// If you set a connection, the `URL`, `Path`, `Header` and `Scheme` will be ignored.
	Connection *websocket.Conn

	// Message that will be sent with the request.
	// a multiline string is valid.
	// eg: { "foo": "bar" }
	Message string

	// Message type used to send the call. It's based on Gorilla's message types
	// https://pkg.go.dev/github.com/gorilla/websocket#pkg-constants
	// eg: websocket.TextMessage
	MessageType int

	// CloseConnectionAfterCall will close the connection after the call is made.
	CloseConnectionAfterCall bool
}
