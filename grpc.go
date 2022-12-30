package integration

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jarcoal/httpmock"
	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
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

func grpcTest(testCase GRPCTestCase) error {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, assertion := range testCase.Assertions {
		err := assertion.Setup()
		if err != nil {
			return errors.New(errString(err, testCase.Description, "failed to setup assertion"))
		}
	}

	args := []reflect.Value{reflect.ValueOf(context.Background()), reflect.ValueOf(testCase.Call.Argument)}
	resp := reflect.ValueOf(testCase.Call.Client).MethodByName(testCase.Call.Function).Call(args)

	err := assertGRPCResponse(testCase.Output, resp)
	if err != nil {
		return errors.New(errString(err, testCase.Description, "failed to assert GRPC response"))
	}

	for _, assertion := range testCase.Assertions {
		err := assertion.Assert()
		if err != nil {
			return errors.New(errString(err, testCase.Description, "failed to assert"))
		}
	}

	return nil
}

func assertGRPCResponse(expected expect.Output, resp []reflect.Value) error {
	respValue := fmt.Sprintf("%v", resp[0])
	respErr, _ := resp[1].Interface().(error)
	expectedRespValue := fmt.Sprintf("%v", expected.Response)

	if respValue != expectedRespValue {
		return errors.Errorf("output response should be %v it got %v", expectedRespValue, respValue)
	}

	if respErr == nil && expected.Err == nil {
		return nil
	}

	if (respErr != nil && expected.Err == nil) || (respErr == nil && expected.Err != nil) {
		return errors.Errorf("error response should be %v it got %v", expected.Err, respErr)
	}

	if respErr != nil && expected.Err != nil {
		status, ok := status.FromError(respErr)
		if !ok {
			return errors.Errorf("failed to get error status %v", respErr)
		}

		if expected.Err.Code() != status.Code() {
			return errors.Errorf("error response status should be %v it got %v", expected.Err.Code(), status.Code())
		}

		if expected.Err.Message() != status.Message() {
			return errors.Errorf("error response message should be %v it got %v", expected.Err.Message(), status.Message())
		}
	}

	return nil
}
