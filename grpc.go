package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jarcoal/httpmock"
	"github.com/kinbiko/jsonassert"
	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/utils"
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
)

// GRPCTestCase describes a GRPC test case that will run
type GRPCTestCase struct {
	// Description describes a test case
	// It can be really useful to understand which tests are breaking
	Description string

	// Call is what the test case will try to call
	Call call.Call

	// Output is going to be used to assert if the GRPC response returned what was expected.
	Output expect.Output

	// Assertions that will run in test case
	Assertions []assertion.Assertion
}

// Test runs an GRPC test case
func (t *GRPCTestCase) Test() error {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, assertion := range t.Assertions {
		err := assertion.Setup()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to setup assertion"))
		}
	}

	resp, err := t.call()
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to call GRPC endpoint"))
	}

	err = t.assert(resp)
	if err != nil {
		return errors.New(errString(err, t.Description, "failed to assert GRPC response"))
	}

	for _, assertion := range t.Assertions {
		err := assertion.Assert()
		if err != nil {
			return errors.New(errString(err, t.Description, "failed to assert"))
		}
	}

	return nil
}

func (t *GRPCTestCase) assert(resp []reflect.Value) error {
	respErr, _ := resp[1].Interface().(error)

	respValueJSON, err := json.Marshal(resp[0].Interface())
	if err != nil {
		return errors.Wrap(err, "failed to marshal grpc response to json")
	}

	expectedValueJSON, err := json.Marshal(t.Output.Message)
	if err != nil {
		return errors.Wrap(err, "failed to marshal grpc expected response to json")
	}

	je := utils.JsonError{}
	jsonassert.New(&je).Assertf(string(respValueJSON), string(expectedValueJSON))
	if je.Err != nil {
		return errors.Errorf(" body does not match: %v", je.Err.Error())
	}

	if respErr == nil && t.Output.Err == nil {
		return nil
	}

	if (respErr != nil && t.Output.Err == nil) || (respErr == nil && t.Output.Err != nil) {
		return errors.Errorf("error response should be %v it got %v", t.Output.Err, respErr)
	}

	if respErr != nil && t.Output.Err != nil {
		status, ok := status.FromError(respErr)
		if !ok {
			return errors.Errorf("failed to get error status %v", respErr)
		}

		if t.Output.Err.Code() != status.Code() {
			return errors.Errorf("error response status should be %v it got %v", t.Output.Err.Code(), status.Code())
		}

		if t.Output.Err.Message() != status.Message() {
			return errors.Errorf("error response message should be %v it got %v", t.Output.Err.Message(), status.Message())
		}
	}

	return nil
}

func (t *GRPCTestCase) call() ([]reflect.Value, error) {
	if t.Call.ServiceClient == nil {
		return nil, errors.New(fmt.Sprintf("%s: failed because GRPC client is nil", t.Description))
	}

	args := []reflect.Value{reflect.ValueOf(context.Background()), reflect.ValueOf(t.Call.Message)}
	function := reflect.ValueOf(t.Call.ServiceClient).MethodByName(t.Call.Function)
	if !function.IsValid() {
		return nil, errors.New(fmt.Sprintf("%s: failed because GRPC function is not valid", t.Description))
	}

	return function.Call(args), nil
}
