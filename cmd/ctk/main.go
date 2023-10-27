// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"ctk/internal/pkg/jsonrpc"
	publicrpc "ctk/internal/pkg/public-rpc"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/smithy-go"
)

func main() {

	var jsonrpcData jsonrpc.Request
	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	errDecode := decoder.Decode(&jsonrpcData)

	var resp jsonrpc.Response
	if errDecode != nil {
		resp = jsonrpc.ParseError(jsonrpcData.ID)
		buffer, err := resp.Encode()

		// This shouldn't happen but in the event it does, report an Internal Service Error
		if err != nil {
			buffer, _ := jsonrpc.InternalServiceError(jsonrpcData.ID).Encode()
			fmt.Print(string(buffer))
			log.Fatal(err)
		}

		fmt.Print(string(buffer))
		log.Fatal(errDecode)
	}

	rpcStruct, err := publicrpc.GetRPCStruct(jsonrpcData.Method, jsonrpcData.Params)
	if err != nil {
		var errNoMethodFound *publicrpc.ErrNoMethodFound
		if errors.As(err, &errNoMethodFound) {
			resp = jsonrpc.NoMethodFoundError(jsonrpcData.ID)
		}
		var errParameter *publicrpc.ErrParameter
		if errors.As(err, &errParameter) {
			resp = jsonrpc.InvalidParamsError(jsonrpcData.ID)
		}

		exitWithResponse(resp, jsonrpcData)
	}

	respString, errRPC := rpcStruct.RPCMethod(jsonrpcData.Metadata)

	if errRPC != nil {
		var oe *smithy.OperationError
		if errors.As(errRPC, &oe) {
			resp = jsonrpc.Response{
				JSONRPC: "2.0",
				ID:      jsonrpcData.ID,
				Error: &jsonrpc.ErrCtk{
					Code:    10,
					Message: fmt.Sprintf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap()),
				},
			}
			exitWithResponse(resp, jsonrpcData)
		}

		var apiErr smithy.APIError
		if errors.As(errRPC, &apiErr) {
			resp = jsonrpc.Response{
				JSONRPC: "2.0",
				ID:      jsonrpcData.ID,
				Error: &jsonrpc.ErrCtk{
					Code:    10,
					Message: apiErr.ErrorMessage(),
				},
			}

			exitWithResponse(resp, jsonrpcData)
		}

		resp = jsonrpc.Response{
			JSONRPC: "2.0",
			ID:      jsonrpcData.ID,
			Error: &jsonrpc.ErrCtk{
				Code:    10,
				Message: errRPC.Error(),
			},
		}

		exitWithResponse(resp, jsonrpcData)

	}

	resp = jsonrpc.Response{
		JSONRPC: "2.0",
		ID:      jsonrpcData.ID,
		Result:  respString,
	}

	// Print "Response"
	exitWithResponse(resp, jsonrpcData)
}

func exitWithResponse(resp jsonrpc.Response, jsonrpcData jsonrpc.Request) {
	buffer, err := resp.Encode()

	// This shouldn't happen but in the event it does, report an Internal Service Error
	if err != nil {
		buffer, _ := jsonrpc.InternalServiceError(jsonrpcData.ID).Encode()
		fmt.Print(string(buffer))
		log.Fatal(err)
	}

	fmt.Print(string(buffer))
	os.Exit(0)
}
