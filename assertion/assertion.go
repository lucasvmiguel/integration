package assertion

type Assertion interface {
	Setup() error
	Assert() error
}
