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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	errMessage = "ERROR"
	grpcPort   = 9000
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
		return nil, fmt.Errorf("failed to call endpoint: %w", err)
	}

	return &chat.Message{
		Id:      1,
		Body:    "Hello From the Server!",
		Comment: "test",
	}, nil
}

func init() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
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

	err = Test(&GRPCTestCase{
		Description: "TestGRPC_Successfully",
		Call: call.Call{
			ServiceClient: c,
			Function:      "SayHello",
			Message: &chat.Message{
				Id:      1,
				Body:    "Hello From Client!",
				Comment: "Whatever",
			},
		},
		Output: expect.Output{
			Message: &chat.Message{
				Id:      1,
				Body:    "Hello From the Server!",
				Comment: "<<PRESENCE>>",
			},
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

	err = Test(&GRPCTestCase{
		Description: "TestGRPC_Successfully",
		Call: call.Call{
			ServiceClient: c,
			Function:      "SayHello",
			Message:       &chat.Message{Body: errMessage},
		},
		Output: expect.Output{
			Err: status.New(codes.Unavailable, errMessage),
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}

func TestGRPC_InvalidFunction(t *testing.T) {
	c, err := client()
	if err != nil {
		t.Fatal(c)
	}

	err = Test(&GRPCTestCase{
		Description: "TestGRPC_Successfully",
		Call: call.Call{
			ServiceClient: c,
			Function:      "Invalid",
			Message:       &chat.Message{Body: errMessage},
		},
	})

	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func TestGRPC_NilClient(t *testing.T) {
	err := Test(&GRPCTestCase{
		Description: "TestGRPC_Successfully",
		Call: call.Call{
			ServiceClient: nil,
			Function:      "SayHello",
			Message:       &chat.Message{Body: errMessage},
		},
	})

	if err == nil {
		t.Fatal("err should not be nil")
	}
}

func client() (chat.ChatServiceClient, error) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return chat.NewChatServiceClient(conn), nil
}
