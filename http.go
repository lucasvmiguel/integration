package integration

import (
	"bytes"
	"io"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/kinbiko/jsonassert"
	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/utils"
	"github.com/pkg/errors"
)

// HTTPTestCase describes a HTTP test case that will run
type HTTPTestCase struct {
	// Description describes a test case
	// It can be really useful to understand which tests are breaking
	Description string

	// Request is what the test case will try to call
	// eg: [POST] https://jsonplaceholder.typicode.com/todos
	Request call.Request

	// Response is going to be used to assert if the HTTP endpoint returned what was expected.
	Response expect.Response

	// Assertions that will run in test case
	Assertions []assertion.Assertion
}

// Test runs an HTTP test case
func (t *HTTPTestCase) Test() error {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(t.Request.Method, t.Request.URL, httpmock.InitialTransport.RoundTrip)

	for _, assertion := range t.Assertions {
		err := assertion.Setup()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to setup assertion"))
		}
	}

	resp, err := t.call()
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to call HTTP endpoint"))
	}

	err = t.assert(resp)
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to assert HTTP response"))
	}

	for _, assertion := range t.Assertions {
		err := assertion.Assert()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to assert"))
		}
	}

	return nil
}

func (t *HTTPTestCase) assert(resp *http.Response) error {
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	respBodyString := string(respBody)

	if resp.StatusCode != t.Response.StatusCode {
		return errors.Errorf("response status code should be %d it got %d", t.Response.StatusCode, resp.StatusCode)
	}

	if utils.IsJSON(t.Response.Body) {
		je := utils.JsonError{}
		jsonassert.New(&je).Assertf(respBodyString, t.Response.Body)
		if je.Err != nil {
			return errors.Errorf("response body is a JSON. response body does not match: %v", je.Err.Error())
		}
	} else {
		if respBodyString != t.Response.Body {
			return errors.Errorf("response body is a regular string. response body should be '%s' it got '%s'", t.Response.Body, respBodyString)
		}
	}

	for key, values := range t.Response.Header {
		respHeader := resp.Header.Get(key)
		if respHeader != values[0] {
			return errors.Errorf("response header should be '%s' it got '%s'", values[0], respHeader)
		}
	}

	return nil
}

func (t *HTTPTestCase) call() (*http.Response, error) {
	req, err := t.createHTTPRequest()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call endpoint")
	}
	return resp, nil
}

func (t *HTTPTestCase) createHTTPRequest() (*http.Request, error) {
	var reqBody io.Reader
	reqBodyString := t.Request.Body

	if reqBodyString == "" {
		reqBody = nil
	} else {
		reqBody = bytes.NewBufferString(t.Request.Body)
	}

	req, err := http.NewRequest(t.Request.Method, t.Request.URL, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new http request")
	}
	req.Header = t.Request.Header

	return req, nil
}
