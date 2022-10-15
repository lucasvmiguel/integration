package assertion

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/utils"
	"github.com/lucasvmiguel/integration/mock"

	"github.com/jarcoal/httpmock"
)

func TestHTTPAssert_NoRequestCalled(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assertion := HTTP{}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	err = assertion.Assert()
	if err == nil {
		t.Fatal(err)
	}
}

func TestHTTPAssert_RequestCalled(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assertion := HTTP{
		Request: expect.Request{
			URL:    "https://jsonplaceholder.typicode.com/posts/1",
			Method: http.MethodGet,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("default status code should be %d", http.StatusOK)
	}

	err = assertion.Assert()
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPAssert_RequestCalledMoreThanOnce(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assertion := HTTP{
		Request: expect.Request{
			URL:    "https://jsonplaceholder.typicode.com/posts/1",
			Method: http.MethodGet,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		t.Fatal(err)
	}

	err = assertion.Assert()
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPAssert_RequestCalledAnotherEndpoint(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assertion := HTTP{
		Request: expect.Request{
			URL:    "https://jsonplaceholder.typicode.com/posts/1",
			Method: http.MethodGet,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get("https://unknown")
	if err == nil {
		t.Fatal(err)
	}

	err = assertion.Assert()
	if err == nil {
		t.Fatal(err)
	}
}

func TestHTTPSetup_Success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "https://jsonplaceholder.typicode.com/posts"
	headerKey := "Content-type"
	headerValues := []string{"application/json; charset=UTF-8"}
	header := http.Header{headerKey: headerValues}
	reqBody := `{
		"title": "foo",
		"body": "bar",
		"userId": 1
	}`
	respBody := `{
		"message": "success"
	}`
	respStatusCode := http.StatusAccepted
	assertion := HTTP{
		Request: expect.Request{
			URL:    url,
			Method: http.MethodPost,
			Body:   reqBody,
			Header: header,
		},
		Response: mock.Response{
			StatusCode: respStatusCode,
			Body:       respBody,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	reqBodyBytes := []byte(utils.Trim(reqBody))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBodyBytes))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add(headerKey, headerValues[0])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != respStatusCode {
		t.Fatalf("status code should be %d but it got %d", respStatusCode, resp.StatusCode)
	}

	defer resp.Body.Close()
	respBodyInBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	respBodyTrim := strings.ReplaceAll(string(respBodyInBytes), "\\", "")
	expectedRespBodyTrim := utils.Trim(respBody)
	if respBodyTrim != expectedRespBodyTrim {
		t.Fatalf("response body should be '%s' but it got '%s'", expectedRespBodyTrim, respBodyTrim)
	}
}

func TestHTTPSetup_SuccessWithIgnoreBodyFields(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "https://jsonplaceholder.typicode.com/posts"
	headerKey := "Content-type"
	headerValues := []string{"application/json; charset=UTF-8"}
	header := http.Header{headerKey: headerValues}
	reqBody := `{
		"title": "foo",
		"body": "bar",
		"userId": 1
	}`
	reqBodyExpected := `{
		"title": "foo",
		"body": "bar"
	}`
	respBody := `{
		"message": "success"
	}`
	respStatusCode := http.StatusAccepted
	assertion := HTTP{
		Request: expect.Request{
			IgnoreBodyFields: []string{"userId"},
			URL:              url,
			Method:           http.MethodPost,
			Body:             reqBodyExpected,
			Header:           header,
		},
		Response: mock.Response{
			StatusCode: respStatusCode,
			Body:       respBody,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	reqBodyBytes := []byte(utils.Trim(reqBody))
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBodyBytes))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add(headerKey, headerValues[0])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != respStatusCode {
		t.Fatalf("status code should be %d but it got %d", respStatusCode, resp.StatusCode)
	}

	defer resp.Body.Close()
	respBodyInBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	respBodyTrim := strings.ReplaceAll(string(respBodyInBytes), "\\", "")
	expectedRespBodyTrim := utils.Trim(respBody)
	if respBodyTrim != expectedRespBodyTrim {
		t.Fatalf("response body should be '%s' but it got '%s'", expectedRespBodyTrim, respBodyTrim)
	}
}

func TestHTTPSetup_FailedURL(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	assertion := HTTP{
		Request: expect.Request{
			URL:    "http://unknown",
			Method: http.MethodPost,
		},
		Response: mock.Response{
			StatusCode: http.StatusOK,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.Get("http://google.com")
	if err == nil {
		t.Fatal(err)
	}
}

func TestHTTPSetup_FailedRequestHeader(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "https://jsonplaceholder.typicode.com/posts"
	headerKey := "Content-type"
	headerValues := []string{"application/json; charset=UTF-8"}
	header := http.Header{headerKey: headerValues}
	assertion := HTTP{
		Request: expect.Request{
			URL:    url,
			Method: http.MethodGet,
			Header: header,
		},
		Response: mock.Response{
			StatusCode: http.StatusOK,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.DefaultClient.Do(req)
	if err == nil {
		t.Fatal(err)
	}
}

func TestHTTPSetup_FailedMethod(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "https://jsonplaceholder.typicode.com/posts"
	assertion := HTTP{
		Request: expect.Request{
			URL:    url,
			Method: http.MethodGet,
		},
		Response: mock.Response{
			StatusCode: http.StatusOK,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.DefaultClient.Do(req)
	if err == nil {
		t.Fatal(err)
	}
}

func TestHTTPSetup_FailedRequestBody(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := "https://jsonplaceholder.typicode.com/posts"
	reqBody := `{
		"title": "foo",
		"body": "bar",
		"userId": 1
	}`
	assertion := HTTP{
		Request: expect.Request{
			URL:    url,
			Method: http.MethodPost,
			Body:   reqBody,
		},
		Response: mock.Response{
			StatusCode: http.StatusAccepted,
		},
	}

	err := assertion.Setup()
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		t.Fatal(err)
	}

	_, err = http.DefaultClient.Do(req)
	if err == nil {
		t.Fatal(err)
	}
}
