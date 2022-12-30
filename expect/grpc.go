package expect

import "google.golang.org/grpc/status"

// Output is used to validate if a GRPC response was returned with the correct parameters
type Output struct {
	// Message expected in the GRPC response
	// Eg: &chat.Message{Id: 1, Body: "Hello From the Server!", Comment: "<<PRESENCE>>"}
	Message interface{}

	// Error expected in the GRPC response
	// Eg: status.New(codes.Unavailable, "error message"),
	Err *status.Status
}
