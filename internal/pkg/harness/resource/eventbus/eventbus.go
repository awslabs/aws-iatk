// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"context"
	"ctk/internal/pkg/harness"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
)

const (
	ResourceType = "AWS::Events::EventBus"
)

type EventBus struct {
	Name string
	ARN  arn.ARN
}

func (b *EventBus) Resource() harness.Resource {
	return harness.Resource{
		Type:       ResourceType,
		PhysicalID: b.Name,
		ARN:        b.ARN.String(),
	}
}

func Get(ctx context.Context, api DescribeEventBusAPI, name string) (*EventBus, error) {
	output, err := api.DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{
		Name: aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("cannot get event bus %q: %v", name, err)
	}

	busARN, _ := arn.Parse(aws.ToString(output.Arn))

	return &EventBus{
		Name: name,
		ARN:  busARN,
	}, nil
}

//go:generate mockery --name DescribeEventBusAPI
type DescribeEventBusAPI interface {
	DescribeEventBus(ctx context.Context, params *eventbridge.DescribeEventBusInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeEventBusOutput, error)
}
