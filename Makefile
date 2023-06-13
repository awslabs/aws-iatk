test: generate-mocks lint test-internal

build:
	go build -o ./bin/zion ./cmd/zion/main.go

integ-test: build
	go test ./integration/...

generate-mocks:
	go generate mockery ./...

test-internal:
	go test ./internal/...

lint:
	@echo "You may need to run make generate-mocks"
	golangci-lint run

generate-rpc-spec: 
	go run ./cmd/rpcspecs/main.go -o schema/rpc-specs.json
