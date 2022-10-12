package http

import (
	"bytes"
	"fmt"
	"integration/assertion"
	"integration/internal/utils"
	"io"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

// TestCase describes a test case that will run
type TestCase struct {
	// Description describes a test case
	// It can be really useful to understand which tests are breaking
	Description string

	// Request is what the test case will try to call
	// eg: [POST] https://jsonplaceholder.typicode.com/todos
	Request Request

	// ResponseExpected is going to be used to assert if the HTTP endpoint returned what was expected.
	ResponseExpected Response

	// Assertions that will run in test case
	Assertions []assertion.Assertion
}

// Request represents an HTTP request
type Request struct {
	// URL that will be called on the request
	// eg: https://jsonplaceholder.typicode.com/todos
	URL string
	// Method that will be called on the request
	// eg: POST
	Method string
	// Body that will be sent with the request
	// a multiline string is valid
	// eg: { "foo": "bar" }
	Body string
	// Header will be sent with the request
	// eg: content-type=application/json
	Header http.Header
}

// Response represents an HTTP response
type Response struct {
	// StatusCode is the HTTP status code of the response
	StatusCode int
	// Body is the HTTP response body
	Body string
	// IgnoreBodyFields is used to ignore the assertion of some of the body field
	// The syntax used to ignore fields can be found here: https://github.com/tidwall/sjson
	// eg: ["data.transaction_id", "id"]
	IgnoreBodyFields []string
	// Header is the HTTP response headers.
	// Every header set in here will be asserted, others will be ignored.
	// eg: content-type=application/json
	Header http.Header
}

// Test runs a test case
func Test(testCase TestCase) error {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(testCase.Request.Method, testCase.Request.URL, httpmock.InitialTransport.RoundTrip)

	for _, assertion := range testCase.Assertions {
		err := assertion.Setup()
		if err != nil {
			return errors.New(errString(err, testCase, "failed to setup assertion"))
		}
	}

	req, err := createHTTPRequest(testCase)
	if err != nil {
		return errors.New(errString(err, testCase, "failed to create HTTP request"))
	}

	resp, err := callHTTP(req)
	if err != nil {
		return errors.New(errString(err, testCase, "failed to call HTTP endpoint"))
	}

	err = assertResponse(testCase.ResponseExpected, resp)
	if err != nil {
		return errors.New(errString(err, testCase, "failed to assert HTTP response"))
	}

	for _, assertion := range testCase.Assertions {
		err := assertion.Assert()
		if err != nil {
			return errors.New(errString(err, testCase, "failed to assert"))
		}
	}

	return nil
}

func assertResponse(expected Response, resp *http.Response) error {
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	respBodyString := utils.Trim(string(respBody))
	expectedRespBody := utils.Trim(expected.Body)

	if resp.StatusCode != expected.StatusCode {
		return errors.Errorf("response status code should be %d it got %d", expected.StatusCode, resp.StatusCode)
	}

	for _, field := range expected.IgnoreBodyFields {
		respBodyString, err = sjson.Delete(respBodyString, field)
		if err != nil {
			return errors.Errorf("failed to ignore field: %s", field)
		}
	}

	if respBodyString != expectedRespBody {
		return errors.Errorf("response body should be '%s' it got '%s'", expectedRespBody, respBodyString)
	}

	for key, values := range expected.Header {
		respHeader := resp.Header.Get(key)
		if respHeader != values[0] {
			return errors.Errorf("response header should be '%s' it got '%s'", values[0], respHeader)
		}
	}

	return nil
}

func createHTTPRequest(testCase TestCase) (*http.Request, error) {
	var reqBody io.Reader
	reqBodyString := utils.Trim(testCase.Request.Body)

	if reqBodyString == "" {
		reqBody = nil
	} else {
		reqBody = bytes.NewBufferString(reqBodyString)
	}

	req, err := http.NewRequest(testCase.Request.Method, testCase.Request.URL, reqBody)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new http request")
	}
	req.Header = testCase.Request.Header

	return req, nil
}

func callHTTP(req *http.Request) (*http.Response, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call endpoint")
	}
	return resp, nil
}

func errString(err error, testCase TestCase, message string) string {
	return errors.Wrap(err, fmt.Sprintf("%s: %s", testCase.Description, message)).Error()
}
