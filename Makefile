lint:
	go vet ./...

test:
	go test ./... -cover

generate-proto:
	protoc --go_out=. --go-grpc_out=. chat.proto