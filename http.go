package integration

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/utils"
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
	Request call.Request

	// Response is going to be used to assert if the HTTP endpoint returned what was expected.
	Response expect.Response

	// Assertions that will run in test case
	Assertions []assertion.Assertion
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

	err = assertResponse(testCase.Response, resp)
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

func assertResponse(expected expect.Response, resp *http.Response) error {
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
