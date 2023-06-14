package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"zion/internal/pkg/aws/config"
	"zion/internal/pkg/harness/eventbridge/listener"
	"zion/internal/pkg/public-rpc/types"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type PollEventsParams struct {
	ListenerID          string `json:"ListenerId"`
	WaitTimeSeconds     *int32
	MaxNumberOfMessages *int32
	Profile             string
	Region              string
}

func (p *PollEventsParams) RPCMethod() (*types.Result, error) {
	p.setDefaultValues()

	err := p.validateParams()
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()

	cfg, err := config.GetAWSConfig(ctx, p.Region, p.Profile)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %w", err)
	}

	lr, err := listener.Get(ctx, p.ListenerID, listener.NewOptions(cfg))
	if err != nil {
		return nil, fmt.Errorf("error retreiving listener info: %w", err)
	}

	events, err := listener.PollEvents(ctx, lr, *p.WaitTimeSeconds, *p.MaxNumberOfMessages)
	if err != nil {
		return nil, fmt.Errorf("error polling events: %w", err)
	}

	return &types.Result{
		Output: events,
	}, nil
}

func (p *PollEventsParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(listener.PollEvents)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}

func (p *PollEventsParams) validateParams() error {
	if p.ListenerID == "" {
		return errors.New(`missing required param "ListenerId"`)
	}

	if *p.MaxNumberOfMessages <= 0 || *p.MaxNumberOfMessages > 10 {
		return errors.New(`"MaxNumberOfMessages" must be an integer between 1 and 10`)
	}

	if *p.WaitTimeSeconds < 0 || *p.WaitTimeSeconds > 20 {
		return errors.New(`"WaitTimeSeconds" must be an integer between 0 and 20`)
	}
	return nil
}

func (p *PollEventsParams) setDefaultValues() {
	if p.WaitTimeSeconds == nil {
		p.WaitTimeSeconds = aws.Int32(0)
	}
	if p.MaxNumberOfMessages == nil {
		p.MaxNumberOfMessages = aws.Int32(1)
	}
}
