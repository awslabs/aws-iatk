package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"zion/internal/pkg/aws/config"
	"zion/internal/pkg/public-rpc/types"
	zionxray "zion/internal/pkg/xray"
)

type GetTraceTreeParams struct {
	TracingHeader string `json:"TracingHeader"`
	Profile       string `json:"Profile,omitempty"`
	Region        string `json:"Region,omitempty"`
}

func (p *GetTraceTreeParams) RPCMethod() (*types.Result, error) {

	if p.TracingHeader == "" {
		return nil, errors.New(`missing required param "TraceId"`)
	}

	traceId, err := getTracIdFromTracingHeader(p.TracingHeader)

	if err != nil {
		return nil, fmt.Errorf("error while getting trace_id from the tracing header: %v", err)
	}

	ctx := context.TODO()
	cfg, err := config.GetAWSConfig(context.TODO(), p.Region, p.Profile)

	if err != nil {
		return nil, fmt.Errorf("error when loading AWS config: %v", err)
	}

	traceTree, err := zionxray.NewTree(ctx, zionxray.NewTreeOptions(cfg), *traceId)

	return &types.Result{
		Output: traceTree,
	}, err
}

// Folows the logic set in the sdk https://github.com/aws/aws-xray-sdk-python/blob/master/aws_xray_sdk/core/models/trace_header.py
func getTracIdFromTracingHeader(tracingHeader string) (*string, error) {

	splitHeader := strings.Split(tracingHeader, ";")
	for _, headerComponent := range splitHeader {
		splitComponent := strings.Split(headerComponent, "=")
		if splitComponent[0] == "Root" {
			return &splitComponent[1], nil
		}
	}
	return nil, errors.New(`invalid tracing header provided`)

}

func (p *GetTraceTreeParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(zionxray.NewTree)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}
