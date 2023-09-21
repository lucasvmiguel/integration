package assertion

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/kinbiko/jsonassert"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/utils"
	"github.com/lucasvmiguel/integration/mock"

	"github.com/jarcoal/httpmock"
)

// HTTP asserts a http request
type HTTP struct {
	// Request will assert if request was made with correct parameters
	Request expect.Request
	// Response mocks a fake response to avoid your test making real http request over the internet
	Response mock.Response
}

// Setup sets up if request will be called as expected
func (a *HTTP) Setup() error {
	httpmock.RegisterResponder(a.method(), a.Request.URL,
		func(req *http.Request) (*http.Response, error) {
			if req.Body != nil {
				defer req.Body.Close()
				reqBody, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, fmt.Errorf("%s: failed to read request body: %w", a.Request.URL, err)
				}

				reqBodyString := string(reqBody)

				if utils.IsJSON(a.Request.Body) {
					je := utils.JsonError{}
					jsonassert.New(&je).Assertf(reqBodyString, a.Request.Body)
					if je.Err != nil {
						return nil, fmt.Errorf("response body is a JSON. response body does not match: %v", je.Err.Error())
					}
				} else {
					if reqBodyString != a.Request.Body {
						return nil, fmt.Errorf("response body is a regular string. response body should be '%s' it got '%s'", a.Request.Body, reqBodyString)
					}
				}
			}

			for key, values := range a.Request.Header {
				reqHeader := req.Header.Get(key)
				if reqHeader != values[0] {
					return nil, fmt.Errorf("%s: request header should be %s it got %s", a.Request.URL, values[0], reqHeader)
				}
			}

			statusCode := a.Response.StatusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			if utils.IsJSON(a.Response.Body) {
				return httpmock.NewStringResponse(statusCode, utils.Trim(a.Response.Body)), nil
			}
			return httpmock.NewStringResponse(statusCode, a.Response.Body), nil
		},
	)

	return nil
}

// Setup does not do anything because the assertions are created on the setup for the HTTP
func (a *HTTP) Assert() error {
	err := a.validate()
	if err != nil {
		return fmt.Errorf("failed to validate assertion: %w", err)
	}

	reqInfo := fmt.Sprintf("%s %s", a.method(), a.Request.URL)
	callCountInfo := httpmock.GetCallCountInfo()

	expectedTimes := a.Request.Times
	if expectedTimes == 0 {
		expectedTimes = 1
	}

	times, ok := callCountInfo[reqInfo]
	if !ok {
		return fmt.Errorf("HTTP request '%s' has never been called", reqInfo)
	}

	if expectedTimes > times || expectedTimes < times {
		return fmt.Errorf("HTTP request '%s' has been called %d times, expected %d", reqInfo, times, expectedTimes)
	}

	return nil
}

func (a *HTTP) method() string {
	method := a.Request.Method
	if method == "" {
		method = http.MethodGet
	}
	return method
}

func (a *HTTP) validate() error {
	if a.Request.URL == "" {
		return errors.New("URL is required")
	}

	return nil
}
