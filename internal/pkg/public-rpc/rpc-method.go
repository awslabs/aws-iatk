package publicrpc

import (
	"bytes"
	"ctk/internal/pkg/jsonrpc"
	"ctk/internal/pkg/public-rpc/types"
	"encoding/json"
	"fmt"
	"reflect"
)

type Method interface {
	RPCMethod(metadata *jsonrpc.Metadata) (*types.Result, error)
	ReflectOutput() reflect.Value
}

var MethodMap = map[string]Method{}

func init() {
	MethodMap["get_physical_id"] = new(GetPhysicalIdParams)
	MethodMap["get_stack_outputs"] = new(GetStackOutputParams)
	MethodMap["test_harness.eventbridge.add_listener"] = new(AddEbListenerParams)
	MethodMap["test_harness.eventbridge.remove_listeners"] = new(RemoveEbListenersParams)
	MethodMap["test_harness.eventbridge.poll_events"] = new(PollEventsParams)
	MethodMap["get_trace_tree"] = new(GetTraceTreeParams)
	MethodMap["mock.generate_barebone_event"] = new(GenerateBareboneEventsParams)
}

func GetRPCStruct(methodName string, params json.RawMessage) (Method, error) {
	requestParams, ok := MethodMap[methodName]
	if !ok {
		return nil, &ErrNoMethodFound{
			err: fmt.Sprintf("No Method found for %s", methodName),
		}
	}

	decoder := newDecoder(params)
	if err := decoder.Decode(requestParams); err != nil {
		return nil, &ErrParameter{
			ParentErr: err,
		}
	}
	return requestParams, nil
}

func newDecoder(params json.RawMessage) *json.Decoder {
	reader := bytes.NewReader(params)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	return decoder
}
