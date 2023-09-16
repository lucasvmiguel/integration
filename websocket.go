package integration

import (
	"bytes"
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
	"github.com/pkg/errors"
)

// WebsocketTestCase describes a Websocket test case that will run
type WebsocketTestCase struct {
	// Description describes a test case
	// It can be really useful to understand which tests are breaking
	Description string

	// Call is the Websocket server the test case will try to connect and send a message
	Call call.Message

	// Message is going to be used to assert if the Websocket server message returned what was expected.
	Message expect.Message

	// Assertions that will run in test case
	Assertions []assertion.Assertion
}

// Test runs an Websocket test case
func (t *WebsocketTestCase) Test() error {
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

func (t *WebsocketTestCase) assert(message []byte) error {
	content, err := io.ReadAll(bytes.NewBuffer(message))
	if err != nil {
		return errors.Wrap(err, "failed to read message content")
	}

	contentString := string(content)

	if utils.IsJSON(t.Message.Content) {
		je := utils.JsonError{}
		jsonassert.New(&je).Assertf(contentString, t.Message.Content)
		if je.Err != nil {
			return errors.Errorf("content is a JSON. content does not match: %v", je.Err.Error())
		}
	} else {
		if contentString != t.Message.Content {
			return errors.Errorf("content is a regular string. content should be '%s' it got '%s'", t.Message.Content, contentString)
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
		return nil, errors.Errorf("error to connect to the Websocket server: %s", err.Error())
	}

	return conn, nil
}

func (t *WebsocketTestCase) call(conn *websocket.Conn) error {
	err := conn.WriteMessage(websocket.TextMessage, []byte(t.Call.Message))
	if err != nil {
		return errors.Errorf("error to send message to the Websocket server: %s", err.Error())
	}

	return nil
}

func (t *WebsocketTestCase) listenAndCall(conn *websocket.Conn) ([]byte, error) {
	messageChan := make(chan []byte)
	timeout := t.Message.Timeout
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
		return nil, errors.Errorf("error to send message to the Websocket server: %s", err.Error())
	}

	select {
	case message := <-messageChan:
		return message, nil
	case <-time.After(timeout):
		return nil, errors.Errorf("timeout to read message from the Websocket server")
	}
}
