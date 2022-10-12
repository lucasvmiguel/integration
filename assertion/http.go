package assertion

import (
	"fmt"
	"integration/internal/utils"
	"io"
	"net/http"

	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

// HTTPAssertion asserts a http request
type HTTPAssertion struct {
	// RequestExpected will assert if request was made with correct parameters
	RequestExpected RequestExpected
	// ResponseMock mocks a fake response to avoid your test making real http request over the internet
	ResponseMock ResponseMock
}

// RequestExpected struct is used to validate if a request was made with the correct parameters
type RequestExpected struct {
	// URL request url that must be called
	URL string
	// Method request method that must be called with
	Method string
	// Header request header that will be asserted with
	// Every header set in here will be asserted, others will be ignored.
	Header http.Header
	// Body request body that must be called with
	// a multiline string is valid
	// eg: { "foo": "bar" }
	Body string
	// IgnoreBodyFields is used to ignore the assertion of some of the body field
	// The syntax used to ignore fields can be found here: https://github.com/tidwall/sjson
	// eg: ["data.transaction_id", "id"]
	IgnoreBodyFields []string
}

// ResponseMock struct is used to return a fake http response to your application
type ResponseMock struct {
	// StatusCode is the HTTP status code of the response
	StatusCode int
	// Body is the HTTP response body
	// a multiline string is valid
	// eg: { "foo": "bar" }
	Body string
}

// Setup sets up if request will be called as expected
func (a *HTTPAssertion) Setup() error {
	httpmock.RegisterResponder(a.RequestExpected.Method, a.RequestExpected.URL,
		func(req *http.Request) (*http.Response, error) {
			if req.Body != nil {
				defer req.Body.Close()
				reqBody, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("%s: failed to read request body", a.RequestExpected.URL))
				}

				reqBodyString := utils.Trim(string(reqBody))
				expectedReqBody := utils.Trim(a.RequestExpected.Body)

				for _, field := range a.RequestExpected.IgnoreBodyFields {
					reqBodyString, err = sjson.Delete(reqBodyString, field)
					if err != nil {
						return nil, errors.Errorf("%s: failed to ignore field: %s", a.RequestExpected.URL, field)
					}
				}

				if reqBodyString != expectedReqBody {
					return nil, errors.Errorf("%s: request body should be %s it got %s", a.RequestExpected.URL, expectedReqBody, reqBodyString)
				}
			}

			for key, values := range a.RequestExpected.Header {
				reqHeader := req.Header.Get(key)
				if reqHeader != values[0] {
					return nil, errors.Errorf("%s: request header should be %s it got %s", a.RequestExpected.URL, values[0], reqHeader)
				}
			}

			return httpmock.NewStringResponse(a.ResponseMock.StatusCode, utils.Trim(a.ResponseMock.Body)), nil
		},
	)

	return nil
}

// Setup does not do anything because the assertions are created on the setup for the HTTPassertion
func (a *HTTPAssertion) Assert() error {
	return nil
}
