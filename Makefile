# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

dev-test: generate-mocks lint test-internal

test: generate-mocks test-internal

build:
	go build -o ./bin/zion ./cmd/zion/main.go

integ-test: build
	PATH=$(PWD)/bin:$(PATH) go test ./integration/...

generate-mocks:
	go generate mockery ./...

test-internal:
	go test ./internal/...

lint:
	@echo "You may need to run make generate-mocks"
	golangci-lint run

generate-rpc-spec: 
	go run ./cmd/rpcspecs/main.go -o schema/rpc-specs.json
