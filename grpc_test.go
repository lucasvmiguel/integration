package integration

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"testing"

	"github.com/lucasvmiguel/integration/assertion"
	"github.com/lucasvmiguel/integration/call"
	"github.com/lucasvmiguel/integration/expect"
	"github.com/lucasvmiguel/integration/internal/chat"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	errMessage = "ERROR"
	port       = 9000
)

type Server struct {
	chat.ChatServiceServer
}

func (s *Server) SayHello(ctx context.Context, in *chat.Message) (*chat.Message, error) {
	if in.Body == errMessage {
		return nil, status.Error(codes.Unavailable, errMessage)
	}

	_, err := http.Get("https://jsonplaceholder.typicode.com/posts/1")
	if err != nil {
		return nil, errors.Wrap(err, "failed to call endpoint")
	}

	return &chat.Message{Body: "Hello From the Server!"}, nil
}

func init() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := Server{}

	grpcServer := grpc.NewServer()

	chat.RegisterChatServiceServer(grpcServer, &s)

	go grpcServer.Serve(lis)
}

func TestGRPC_Successfully(t *testing.T) {
	c, err := client()
	if err != nil {
		t.Fatal(c)
	}

	err = Test(GRPCTestCase{
		Description: "TestGRPC_Successfully",
		Call: call.Call{
			Client:   c,
			Function: "SayHello",
			Argument: &chat.Message{Body: "Hello From Client!"},
		},
		Output: expect.Output{
			Response: &chat.Message{Body: "Hello From the Server!"},
		},
		Assertions: []assertion.Assertion{
			&assertion.HTTP{
				Request: expect.Request{
					URL:    "https://jsonplaceholder.typicode.com/posts/1",
					Method: http.MethodGet,
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestGRPC_Error(t *testing.T) {
	c, err := client()
	if err != nil {
		t.Fatal(c)
	}

	err = Test(GRPCTestCase{
		Description: "TestGRPC_Successfully",
		Call: call.Call{
			Client:   c,
			Function: "SayHello",
			Argument: &chat.Message{Body: errMessage},
		},
		Output: expect.Output{
			Err: status.New(codes.Unavailable, errMessage),
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func client() (chat.ChatServiceClient, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return chat.NewChatServiceClient(conn), nil
}
