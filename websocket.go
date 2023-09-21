package integration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jarcoal/httpmock"
	"github.com/kinbiko/jsonassert"
	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/utils"
)

// WebsocketTestCase describes a Websocket test case that will run
type WebsocketTestCase struct {
	// Description describes a test case
	// It can be really useful to understand which tests are breaking
	Description string

	// Call is the Websocket server the test case will try to connect and send a message
	Call call.Websocket

	// Receive is going to be used to assert if the Websocket server message returned what was expected.
	// This field is optional as a Websocket server can never send a message to the client.
	Receive expect.Message

	// Assertions that will run in test case
	Assertions []assertion.Assertion
}

// Test runs an Websocket test case
func (t *WebsocketTestCase) Test() error {
	err := t.validate()
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to validate test case"))
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, assertion := range t.Assertions {
		err := assertion.Setup()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to setup assertion"))
		}
	}

	if t.Call.Connection == nil {
		conn, err := t.connect()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to connect Websocket endpoint"))
		}
		t.Call.Connection = conn
	}

	resp, err := t.listenAndCall(t.Call.Connection)
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to call and/or listen the Websocket server"))
	}

	err = t.assert(resp)
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to assert Websocket response"))
	}

	for _, assertion := range t.Assertions {
		err := assertion.Assert()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to assert"))
		}
	}

	return nil
}

// Connection returns the Websocket connection
func (t *WebsocketTestCase) Connection() *websocket.Conn {
	return t.Call.Connection
}

func (t *WebsocketTestCase) assert(message []byte) error {
	content, err := io.ReadAll(bytes.NewBuffer(message))
	if err != nil {
		return fmt.Errorf("failed to read message content: %w", err)
	}

	contentString := string(content)

	if utils.IsJSON(t.Receive.Content) {
		je := utils.JsonError{}
		jsonassert.New(&je).Assertf(contentString, t.Receive.Content)
		if je.Err != nil {
			return fmt.Errorf("content is a JSON. content does not match: %v", je.Err.Error())
		}
	} else {
		if contentString != t.Receive.Content {
			return fmt.Errorf("content is a regular string. content should be '%s' it got '%s'", t.Receive.Content, contentString)
		}
	}

	return nil
}

func (t *WebsocketTestCase) connect() (*websocket.Conn, error) {
	if t.Call.Scheme == "" {
		t.Call.Scheme = call.WebsocketSchemeWS
	}

	u := url.URL{Scheme: string(t.Call.Scheme), Host: t.Call.URL, Path: t.Call.Path}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), t.Call.Header)
	if err != nil {
		return nil, fmt.Errorf("error to connect to the Websocket server (%s): %s", u.String(), err.Error())
	}

	return conn, nil
}

func (t *WebsocketTestCase) call(conn *websocket.Conn) error {
	err := conn.WriteMessage(websocket.TextMessage, []byte(t.Call.Message))
	if err != nil {
		return fmt.Errorf("error to send message to the Websocket server: %s", err.Error())
	}

	return nil
}

func (t *WebsocketTestCase) listenAndCall(conn *websocket.Conn) ([]byte, error) {
	messageChan := make(chan []byte)
	timeout := t.Receive.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	go func() {
		for {
			_, m, err := conn.ReadMessage()
			if err != nil {
				break
			}

			if m != nil {
				messageChan <- m
				break
			}
		}
	}()

	err := t.call(conn)
	if err != nil {
		return nil, fmt.Errorf("error to send message to the Websocket server: %s", err.Error())
	}

	select {
	case message := <-messageChan:
		return message, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout to read message from the Websocket server")
	}
}

func (t *WebsocketTestCase) validate() error {
	if t.Call.URL == "" {
		return errors.New("URL is required")
	}

	return nil
}
