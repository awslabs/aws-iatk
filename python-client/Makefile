# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

build-zion-service:
	go build -o src/zion_service/ ../cmd/zion

init: build-client
	pip install -e '.[test]'

integ-test:
	pytest -vvv tests/integration

contract-test:
	pytest -vvv tests/contract

unit-test:
	pytest -vvv tests/unit

build-client:
	python -m build
	pip3 install -e .

copy-service-source-for-sdist:
	mkdir -p ./src/zion_src
	cp -r ../cmd ./src/zion_src
	cp -r ../internal ./src/zion_src
	cp ../go.mod ./src/zion_src/go.mod
	cp ../go.sum ./src/zion_src/go.sum

ci-build-distribution-mac: copy-sevice-source-for-sdist
	cibuildwheel --platform macos

install-go-build-binary:
	curl -O -L "https://golang.org/dl/go1.21.1.linux-amd64.tar.gz" 
	tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz 
	export PATH=$PATH:/usr/local/go/bin 
	env GOOS=linux GOARCH=$LINUX_ARCHITECTURE go build -C ./src/zion_src -o ../zion_service/ ./cmd/zion	
