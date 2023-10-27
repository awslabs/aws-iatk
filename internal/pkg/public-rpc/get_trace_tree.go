package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"ctk/internal/pkg/aws/config"
	"ctk/internal/pkg/jsonrpc"
	"ctk/internal/pkg/public-rpc/types"
	ctkxray "ctk/internal/pkg/xray"
)

type GetTraceTreeParams struct {
	TracingHeader string `json:"TracingHeader"`
	Profile       string `json:"Profile,omitempty"`
	Region        string `json:"Region,omitempty"`
}

func (p *GetTraceTreeParams) RPCMethod(metadata *jsonrpc.Metadata) (*types.Result, error) {

	if p.TracingHeader == "" {
		return nil, errors.New(`missing required param "TracingHeader"`)
	}

	traceId, err := getTracIdFromTracingHeader(p.TracingHeader)

	if err != nil {
		return nil, fmt.Errorf("error while getting trace_id from the tracing header: %v", err)
	}

	ctx := context.TODO()
	cfg, err := config.GetAWSConfig(context.TODO(), p.Region, p.Profile, metadata)

	if err != nil {
		return nil, fmt.Errorf("error when loading AWS config: %v", err)
	}

	traceTree, err := ctkxray.NewTree(ctx, ctkxray.NewTreeOptions(cfg), *traceId)
	if err != nil {
		return nil, fmt.Errorf("error building trace tree: %w", err)
	}

	return &types.Result{
		Output: traceTree,
	}, nil
}

// Folows the logic set in the sdk https://github.com/aws/aws-xray-sdk-python/blob/master/aws_xray_sdk/core/models/trace_header.py
func getTracIdFromTracingHeader(tracingHeader string) (*string, error) {

	splitHeader := strings.Split(tracingHeader, ";")
	for _, headerComponent := range splitHeader {
		splitComponent := strings.Split(headerComponent, "=")
		if strings.ToLower(splitComponent[0]) == "root" {
			if splitComponent[1] != "" {
				return &splitComponent[1], nil
			} else {
				return nil, errors.New(`invalid tracing header provided`)
			}
		}
	}
	return nil, errors.New(`invalid tracing header provided`)

}

func (p *GetTraceTreeParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(ctkxray.NewTree)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}
