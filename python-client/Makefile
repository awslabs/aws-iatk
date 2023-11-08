# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

build-iatk-service:
	go build -o src/iatk_service/ ../cmd/iatk

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
	rm -rf ./src/iatk_src
	rm -rf ./src/iatk_service
	mkdir -p ./src/iatk_src
	mkdir -p ./src/iatk_service
	cp -r ../cmd ./src/iatk_src
	cp -r ../internal ./src/iatk_src
	cp ../go.mod ./src/iatk_src/go.mod
	cp ../go.sum ./src/iatk_src/go.sum
	cp ../THIRD_PARTY_LICENSES ./src/iatk_service/THIRD_PARTY_LICENSES

ci-build-distribution-mac: copy-sevice-source-for-sdist
	cibuildwheel --platform macos
