package expect

import "time"

// Message is used to validate if a Websocket message is correct
type Message struct {
	// Content expected in the Websocket message.
	// A multiline string is valid.
	// eg: { "foo": "bar" }
	Content string

	// Timeout is the time to wait for a message to be received.
	Timeout time.Duration
}
