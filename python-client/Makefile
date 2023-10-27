# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

build-ctk-service:
	go build -o src/ctk_service/ ../cmd/ctk

init: build-client
	pip install -e '.[test]'

integ-test:
	pytest -vvv tests/integration

contract-test:
	pytest -vvv tests/contract

unit-test:
	pytest -vvv tests/unit

build-client:
	pip3 install --upgrade build
	python -m build
	pip3 install -e .

copy-service-source-for-sdist:
	rm -rf ./src/ckt_src
	rm -rf ./src/ctk_service
	mkdir -p ./src/ctk_src
	mkdir -p ./src/ctk_service
	cp -r ../cmd ./src/ctk_src
	cp -r ../internal ./src/ctk_src
	cp ../go.mod ./src/ctk_src/go.mod
	cp ../go.sum ./src/ctk_src/go.sum
	cp ../THIRD_PARTY_LICENSES ./src/ctk_service/THIRD_PARTY_LICENSES

ci-build-distribution-mac: copy-sevice-source-for-sdist
	cibuildwheel --platform macos
