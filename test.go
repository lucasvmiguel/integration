package integration

import (
	"fmt"

	"github.com/pkg/errors"
)

// Tester allows to test a case
type Tester interface {
	Test() error
}

// Test runs a test case
func Test(tester Tester) error {
	return tester.Test()
}

func errString(err error, description string, message string) string {
	return errors.Wrap(err, fmt.Sprintf("%s: %s", description, message)).Error()
}
