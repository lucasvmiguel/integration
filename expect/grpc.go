package expect

import "google.golang.org/grpc/status"

// Output is used to validate if a GRPC response was returned with the correct parameters
type Output struct {
	// Response expected in the GRPC response
	Response interface{}
	// Error expected in the GRPC response
	Err *status.Status
}
