## Prerequisite

First ensure you have Go 1.20+ and Python 3.8+ installed. Note: you only need python installed if you are editing the python code under `python-client` directory. You will also need to clone the repo.

## Package setup

The codebase is split into to chunks: a JSON RPC and Python client. The Python client is isolated to `python-client` directory.

## Go RPC

Assuming you have installed Go, navigate to the root of the repo.

You can build the Go binary by running `make build`. `make lint` will run `golangci-lint` to lint the code. You can run all the unit tests through `make test` or `make dev-test`. 

Before committing and submitting a PR, please run the following:

1. `make dev-test`
2. `make integ-test`

## Python Client

Navigate to `python-client`. Make changes as needed.

Before committing and submitting a PR, please run the following:

1. `make build-ctk-service`
2. `make init`
3. `make unit-test`
4. `make contract-test`



