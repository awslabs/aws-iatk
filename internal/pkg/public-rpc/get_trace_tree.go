package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"zion/internal/pkg/aws/config"
	"zion/internal/pkg/public-rpc/types"
	zionxray "zion/internal/pkg/xray"

	"github.com/aws/aws-sdk-go-v2/service/xray"
)

type GetTraceTreeParams struct {
	TraceId                string `json:"TraceId"`
	FetchChildLinkedTraces bool   `json:"FetchChildLinkedTraces,omitempty"`
	Profile                string `json:"Profile,omitempty"`
	Region                 string `json:"Region,omitempty"`
}

func (p *GetTraceTreeParams) RPCMethod() (*types.Result, error) {

	if p.TraceId == "" {
		return nil, errors.New(`missing required param "TraceId"`)
	}

	ctx := context.TODO()
	cfg, err := config.GetAWSConfig(context.TODO(), p.Region, p.Profile)

	if err != nil {
		return nil, fmt.Errorf("error when loading AWS config: %v", err)
	}

	xrayClient := xray.NewFromConfig(cfg)

	traceTree, err := zionxray.NewTree(ctx, xrayClient, p.TraceId, p.FetchChildLinkedTraces)

	return &types.Result{
		Output: traceTree,
	}, err
}

func (p *GetTraceTreeParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(zionxray.NewTree)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}
