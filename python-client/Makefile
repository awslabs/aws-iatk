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
	pip3 install --upgrade build
	python -m build
	pip3 install -e .

copy-service-source-for-sdist:
	rm -rf ./src/zion_src
	rm -rf ./src/zion_service
	mkdir -p ./src/zion_src
	mkdir -p ./src/zion_service
	cp -r ../cmd ./src/zion_src
	cp -r ../internal ./src/zion_src
	cp ../go.mod ./src/zion_src/go.mod
	cp ../go.sum ./src/zion_src/go.sum
	cp ../THIRD_PARTY_LICENSES ./src/zion_service/THIRD_PARTY_LICENSES

ci-build-distribution-mac: copy-sevice-source-for-sdist
	cibuildwheel --platform macos
