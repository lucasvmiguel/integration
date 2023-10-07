package assertion

type Assertion interface {
	Setup() error
	Assert() error
}

// AnyHTTP returns true if the assertions contains at least one HTTP assertion
func AnyHTTP(assertions []Assertion) bool {
	for _, assertion := range assertions {
		_, ok := assertion.(*HTTP)
		if ok {
			return ok
		}
	}
	return false
}
