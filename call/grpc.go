package call

// Call sets up how a GRPC request will be called
type Call struct {
	// GRPC service client used to call the server
	// eg: ChatServiceClient
	ServiceClient interface{}

	// Function that will be called on the request
	// eg: SayHello
	Function string

	// Message that will be sent with the request
	// Eg: &chat.Message{Id: 1, Body: "Hello From the Server!"}
	Message interface{}
}
