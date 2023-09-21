package integration

import (
	"fmt"
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
	return fmt.Errorf("%s: %s : %w", description, message, err).Error()
}
