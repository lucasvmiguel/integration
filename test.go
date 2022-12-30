package integration

import (
	"fmt"

	"github.com/pkg/errors"
)

type TestCase interface{}

func Test(x TestCase) error {
	switch v := x.(type) {
	case HTTPTestCase:
		return httpTest(v)
	case GRPCTestCase:
		return grpcTest(v)
	default:
		return errors.New("Test function accepts only `GRPCTestCase` or `HTTPTestCase at the moment`")
	}
}

func errString(err error, description string, message string) string {
	return errors.Wrap(err, fmt.Sprintf("%s: %s", description, message)).Error()
}
