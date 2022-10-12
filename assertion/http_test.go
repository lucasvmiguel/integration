package assertion

import (
	"bytes"
	"integration/internal/utils"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestHTTPAssert_Success(t *testing.T) {
	assertion := HTTPAssertion{}

	err := assertion.Assert()
	if err != nil {
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
	assertion := HTTPAssertion{
		RequestExpected: RequestExpected{
			URL:    url,
			Method: http.MethodPost,
			Body:   reqBody,
			Header: header,
		},
		ResponseMock: ResponseMock{
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
	assertion := HTTPAssertion{
		RequestExpected: RequestExpected{
			IgnoreBodyFields: []string{"userId"},
			URL:              url,
			Method:           http.MethodPost,
			Body:             reqBodyExpected,
			Header:           header,
		},
		ResponseMock: ResponseMock{
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

	assertion := HTTPAssertion{
		RequestExpected: RequestExpected{
			URL:    "http://unknown",
			Method: http.MethodPost,
		},
		ResponseMock: ResponseMock{
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
	assertion := HTTPAssertion{
		RequestExpected: RequestExpected{
			URL:    url,
			Method: http.MethodGet,
			Header: header,
		},
		ResponseMock: ResponseMock{
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
	assertion := HTTPAssertion{
		RequestExpected: RequestExpected{
			URL:    url,
			Method: http.MethodGet,
		},
		ResponseMock: ResponseMock{
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
	assertion := HTTPAssertion{
		RequestExpected: RequestExpected{
			URL:    url,
			Method: http.MethodPost,
			Body:   reqBody,
		},
		ResponseMock: ResponseMock{
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
