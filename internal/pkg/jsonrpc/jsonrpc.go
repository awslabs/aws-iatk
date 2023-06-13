package jsonrpc

import (
	"bytes"
	"encoding/json"
)

type Request struct {
	// JSONRPC describes the version of the JSON RPC protocol. Defaults to `2.0`.
	JSONRPC string `json:"jsonrpc"`
	// ID identifies a unique request.
	ID *string `json:"id"`
	// Method describes the intention of the request.
	Method string `json:"method"`
	// Params contains the payload of the request. Usually parsed into a specific struct for further processing.
	Params json.RawMessage `json:"params"`
}

type Response struct {
	// JSONRPC describes the version of the JSON RPC protocol. Defaults to `2.0`.
	JSONRPC string `json:"jsonrpc"`
	// ID identifies a unique request.
	ID *string `json:"id"`
	// Result contains the payload of the response.
	Result interface{} `json:"result,omitempty"`
	// Error contains the error response details.
	Error *ErrZion `json:"error,omitempty"`
}

type ErrZion struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	// Data should be here but not sure about the type (jfuss)
}

func (r Response) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	enc := json.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
