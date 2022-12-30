package call

// Call sets up how a GRPC request will be called
type Call struct {
	// GRPC client used to call the server
	// eg: ChatServiceClient
	Client interface{}

	// Function that will be called on the request
	// eg: SayHello
	Function string

	// Arguments that will be sent with the request
	Argument interface{}
}
