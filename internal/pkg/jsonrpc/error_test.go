package jsonrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseError(t *testing.T) {
	id := "1"
	errResponse := ParseError(&id)

	assert.Equal(t, "1", *errResponse.ID, "ParseError.ID should match value passed in")
	assert.Equal(t, -32700, errResponse.Error.Code, "ParseError.Error.Code should match -32700")
	assert.Equal(t, "Parse error", errResponse.Error.Message, `ParseError.Error.Message should be "Parse error"`)
}

func TestNoMethodFound(t *testing.T) {
	id := "1"
	errResponse := NoMethodFoundError(&id)

	assert.Equal(t, "1", *errResponse.ID, "NoMethodFoundError.ID should match value passed in")
	assert.Equal(t, -32601, errResponse.Error.Code, "NoMethodFoundError.Error.Code should match -32700")
	assert.Equal(t, "Method not found", errResponse.Error.Message, `NoMethodFoundError.Error.Message should be "Parse error"`)
}

func TestInternalServiceError(t *testing.T) {
	id := "1"
	errResponse := InternalServiceError(&id)

	assert.Equal(t, "1", *errResponse.ID, "InternalServiceError.ID should match value passed in")
	assert.Equal(t, -32603, errResponse.Error.Code, "InternalServiceError.Error.Code should match -32603")
	assert.Equal(t, "Internal error", errResponse.Error.Message, `InternalServiceError.Error.Message should be "Internal error"`)
}

func TestInvalidParamsError(t *testing.T) {
	id := "1"
	errResponse := InvalidParamsError(&id)

	assert.Equal(t, "1", *errResponse.ID, "InternalServiceError.ID should match value passed in")
	assert.Equal(t, -32602, errResponse.Error.Code, "InternalServiceError.Error.Code should match -32602")
	assert.Equal(t, "Invalid params", errResponse.Error.Message, `InternalServiceError.Error.Message should be "Invalid params"`)
}
