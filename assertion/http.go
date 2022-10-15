package assertion

import (
	"fmt"
	"io"
	"net/http"

	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/utils"
	"github.com/lucasvmiguel/integration/mock"

	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
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
	httpmock.RegisterResponder(a.Request.Method, a.Request.URL,
		func(req *http.Request) (*http.Response, error) {
			if req.Body != nil {
				defer req.Body.Close()
				reqBody, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, errors.Wrap(err, fmt.Sprintf("%s: failed to read request body", a.Request.URL))
				}

				reqBodyString := utils.Trim(string(reqBody))
				expectedReqBody := utils.Trim(a.Request.Body)

				for _, field := range a.Request.IgnoreBodyFields {
					reqBodyString, err = sjson.Delete(reqBodyString, field)
					if err != nil {
						return nil, errors.Errorf("%s: failed to ignore field: %s", a.Request.URL, field)
					}
				}

				if reqBodyString != expectedReqBody {
					return nil, errors.Errorf("%s: request body should be %s it got %s", a.Request.URL, expectedReqBody, reqBodyString)
				}
			}

			for key, values := range a.Request.Header {
				reqHeader := req.Header.Get(key)
				if reqHeader != values[0] {
					return nil, errors.Errorf("%s: request header should be %s it got %s", a.Request.URL, values[0], reqHeader)
				}
			}

			statusCode := a.Response.StatusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			return httpmock.NewStringResponse(statusCode, utils.Trim(a.Response.Body)), nil
		},
	)

	return nil
}

// Setup does not do anything because the assertions are created on the setup for the HTTP
func (a *HTTP) Assert() error {
	reqInfo := fmt.Sprintf("%s %s", a.Request.Method, a.Request.URL)
	callCountInfo := httpmock.GetCallCountInfo()

	times, ok := callCountInfo[reqInfo]
	if ok && times > 0 {
		return nil
	}

	return fmt.Errorf("HTTP request '%s' has never been called", reqInfo)
}
