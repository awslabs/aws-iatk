package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"zion/internal/pkg/aws/config"
	"zion/internal/pkg/harness/eventbridge/listener"
	"zion/internal/pkg/harness/resource/eventrule"
	"zion/internal/pkg/harness/tags"
	"zion/internal/pkg/public-rpc/types"
)

type AddEbListenerParams struct {
	EventBusName     string
	EventPattern     string
	Input            string
	InputPath        string
	InputTransformer *eventrule.InputTransformer
	Tags             map[string]string
	Profile          string
	Region           string
}

func (p *AddEbListenerParams) RPCMethod() (*types.Result, error) {
	err := p.validateParams()
	if err != nil {
		return nil, err
	}

	err = tags.ValidateTags(p.Tags)
	if err != nil {
		return nil, fmt.Errorf("invalid tags: %v", err)
	}

	ctx := context.TODO()

	cfg, err := config.GetAWSConfig(ctx, p.Region, p.Profile)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %v", err)
	}

	lr, err := listener.New(ctx, p.EventBusName, p.EventPattern, p.Input, p.InputPath, p.InputTransformer, p.Tags, listener.NewOptions(cfg))
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
	if p.EventPattern == "" {
		return errors.New(`missing required param "EventPattern"`)
	}
	if (p.Input != "" && p.InputPath != "") || (p.InputPath != "" && p.InputTransformer != nil) || (p.Input != "" && p.InputTransformer != nil) {
		return errors.New(`provide only one of "Input", "InputPath" and "InputTransformer"`)
	}
	return nil
}
