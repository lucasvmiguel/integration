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
	// It can be really useful  to understand which tests are breaking
	Description string

	// Call is the Websocket server the test case will try to connect and send a message
	Call call.Websocket

	// Receive is going to be used to assert if the Websocket server message returned what was expected.
	// This field is optional as a Websocket server can never send a message to the client.
	Receive *expect.Message

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

	if assertion.AnyHTTP(t.Assertions) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
	}

	err = t.setupAssertions()
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to setup assertions"))
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

	var resp []byte

	if t.Call.MessageType == websocket.PingMessage {
		resp, err = t.readAndSendPing()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to read and send ping message"))
		}
	} else {
		resp, err = t.readAndSendMessage()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to read and send message"))
		}
	}

	if t.Receive != nil {
		err = t.assert(resp)
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to assert Websocket response"))
		}
	}

	if t.Assertions != nil {
		for _, assertion := range t.Assertions {
			err := assertion.Assert()
			if err != nil {
				return errors.New(errString(err, t.Description, "failed to assert"))
			}
		}
	}

	if t.Call.CloseConnectionAfterCall {
		err = t.connection.Close()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to close Websocket connection"))
		}
	}

	return nil
}

// Connection returns the Websocket connection
func (t *WebsocketTestCase) Connection() *ws.WebsocketConnection {
	return t.connection
}

func (t *WebsocketTestCase) setupAssertions() error {

	if t.Assertions != nil {
		for _, assertion := range t.Assertions {
			err := assertion.Setup()
			if err != nil {
				return fmt.Errorf("failed to setup assertion: %w", err)
			}
		}
	}

	return nil
}

func (t *WebsocketTestCase) readAndSendMessage() ([]byte, error) {
	messageType := t.Call.MessageType
	if messageType == 0 {
		messageType = websocket.TextMessage
	}

	if t.Receive == nil {
		err := t.connection.Send(messageType, []byte(t.Call.Message))
		if err != nil {
			return nil, fmt.Errorf("failed to send message: %w", err)
		}
		return nil, nil
	}

	var resp []byte
	msg := make(chan []byte)
	errChan := make(chan error)

	go func() {
		_, m, err := t.connection.Read()
		if err != nil {
			errChan <- err
			return
		}

		msg <- m
	}()

	err := t.connection.Send(messageType, []byte(t.Call.Message))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	select {
	case resp = <-msg:
		return resp, nil
	case err := <-errChan:
		return nil, fmt.Errorf("failed to send read message, channel error: %w", err)
	case <-time.After(t.timeout()):
		return nil, errors.New("timeout to reading message from the Websocket server")
	}
}

func (t *WebsocketTestCase) readAndSendPing() ([]byte, error) {
	if t.Receive == nil {
		err := t.connection.Send(t.Call.MessageType, []byte(t.Call.Message))
		if err != nil {
			return nil, fmt.Errorf("failed to ping message: %w", err)
		}
		return nil, nil
	}

	msg := make(chan []byte)

	t.connection.SetPongHandler(func(data string) error {
		m := []byte(data)
		msg <- m
		return t.connection.Send(websocket.PingMessage, m)
	})

	err := t.connection.Send(t.Call.MessageType, []byte(t.Call.Message))
	if err != nil {
		return nil, fmt.Errorf("failed to ping message: %w", err)
	}

	// for some reason, this is required to read the pong message
	go func() { t.connection.Read() }()

	select {
	case resp := <-msg:
		return resp, nil
	case <-time.After(t.timeout()):
		return nil, errors.New("timeout to reading pong from the Websocket server")
	}
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

func (t *WebsocketTestCase) timeout() time.Duration {
	defaultTimeout := 5 * time.Second

	if t.Receive == nil {
		return defaultTimeout
	}

	timeout := t.Receive.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return timeout
}

func (t *WebsocketTestCase) validate() error {
	if t.Call.Connection == nil && t.Call.URL == "" {
		return errors.New("URL is required when Connection is nil")
	}

	return nil
}
