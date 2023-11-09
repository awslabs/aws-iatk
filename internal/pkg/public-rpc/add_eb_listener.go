package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"iatk/internal/pkg/aws/config"
	"iatk/internal/pkg/harness/eventbridge/listener"
	"iatk/internal/pkg/harness/tags"
	"iatk/internal/pkg/jsonrpc"
	"iatk/internal/pkg/public-rpc/types"
	"reflect"
)

type AddEbListenerParams struct {
	EventBusName string
	TargetId     string
	RuleName     string
	Tags         map[string]string
	Profile      string
	Region       string
}

func (p *AddEbListenerParams) RPCMethod(metadata *jsonrpc.Metadata) (*types.Result, error) {
	err := p.validateParams()
	if err != nil {
		return nil, err
	}

	err = tags.ValidateTags(p.Tags)
	if err != nil {
		return nil, fmt.Errorf("invalid tags: %v", err)
	}

	ctx := context.TODO()

	cfg, err := config.GetAWSConfig(ctx, p.Region, p.Profile, metadata)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %v", err)
	}

	lr, err := listener.New(ctx, p.EventBusName, p.TargetId, p.RuleName, p.Tags, listener.NewOptions(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to locate test target: %w", err)
	}

	output, err := listener.Create(ctx, lr)
	if err != nil {
		return nil, fmt.Errorf("failed to create eb listener: %w", err)
	}

	return &types.Result{
		Output: output,
	}, nil
}

func (p *AddEbListenerParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(listener.Create)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}

func (p *AddEbListenerParams) validateParams() error {
	if p.EventBusName == "" {
		return errors.New(`missing required param "EventBusName"`)
	}
	if p.RuleName == "" {
		return errors.New(`missing required param "RuleName"`)
	}
	return nil
}
