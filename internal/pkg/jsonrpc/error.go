package jsonrpc

func ParseError(id *string) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrIatk{
			Code:    -32700,
			Message: "Parse error",
		},
	}
}

func NoMethodFoundError(id *string) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrIatk{
			Code:    -32601,
			Message: "Method not found",
		},
	}
}

func InternalServiceError(id *string) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrIatk{
			Code:    -32603,
			Message: "Internal error",
		},
	}
}

func InvalidParamsError(id *string) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &ErrIatk{
			Code:    -32602,
			Message: "Invalid params",
		},
	}
}
