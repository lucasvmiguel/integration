lint:
	go vet ./...

test:
	go test ./... -cover

generate-proto:
	protoc --go_out=. --go-grpc_out=. chat.proto

release:
	git tag v$(VERSION)
	git push origin v$(VERSION)
