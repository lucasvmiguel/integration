package integration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jarcoal/httpmock"
	"github.com/kinbiko/jsonassert"
	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/utils"
	"github.com/lucasvmiguel/integration/ws"
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

	connection *ws.WebsocketConnection
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
		t.connection = conn
	} else {
		t.connection = t.Call.Connection
	}

	resp, err := t.listenAndCall()
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to call and/or listen the Websocket server"))
	}

	if !t.isEmptyReceive() {
		err = t.assert(resp)
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to assert Websocket response"))
		}
	}

	for _, assertion := range t.Assertions {
		err := assertion.Assert()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to assert"))
		}
	}

	if t.Call.CloseConnectionAfterCall {
		t.connection.Close()
	}

	return nil
}

// Connection returns the Websocket connection
func (t *WebsocketTestCase) Connection() *ws.WebsocketConnection {
	return t.connection
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

func (t *WebsocketTestCase) connect() (*ws.WebsocketConnection, error) {
	conn, err := ws.NewWebsocketConnection(string(t.Call.Scheme), t.Call.URL, t.Call.Path, t.Call.Header)
	if err != nil {
		return nil, fmt.Errorf("error to connect to the Websocket server: %s", err.Error())
	}
	return conn, nil
}

func (t *WebsocketTestCase) send(messageType int) error {
	err := t.connection.Send(messageType, []byte(t.Call.Message))
	if err != nil {
		return fmt.Errorf("error to send message to the Websocket server: %s", err.Error())
	}
	return nil
}

func (t *WebsocketTestCase) listenAndCall() ([]byte, error) {
	messageType := t.Call.MessageType
	if messageType == 0 {
		messageType = websocket.TextMessage
	}

	if t.isEmptyReceive() || t.Call.MessageType == websocket.PingMessage || t.Call.MessageType == websocket.PongMessage || t.Call.MessageType == websocket.CloseMessage {
		err := t.send(messageType)
		if err != nil {
			return nil, fmt.Errorf("error to send message to the Websocket server: %s", err.Error())
		}
		return nil, nil
	}

	messageChan := make(chan []byte)
	timeout := t.Receive.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	go func() {
		for {
			_, m, err := t.connection.Read()
			if err != nil {
				break
			}

			if m != nil {
				messageChan <- m
				break
			}
		}
	}()

	err := t.send(messageType)
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

func (t *WebsocketTestCase) isEmptyReceive() bool {
	return t.Receive.Content == "" && t.Receive.Timeout == 0
}

func (t *WebsocketTestCase) validate() error {
	if t.Call.Connection == nil && t.Call.URL == "" {
		return errors.New("URL is required when Connection is nil")
	}

	return nil
}
