build-zion-service:
	go build -o src/zion_service/ ../cmd/zion

init: build-client
	pip install -e '.[test]'

integ-test: build-client
	pytest -vvv tests/integration

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