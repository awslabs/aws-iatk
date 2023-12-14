// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package eventbus

import (
	"context"
	"errors"
	"fmt"
	"iatk/internal/pkg/harness"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/stretchr/testify/assert"
)

const (
	testBusName   string = "my-event-bus"
	testPartition string = "aws"
	testService   string = "sqs"
	testRegion    string = "us-west-2"
	testAccountID string = "123456789012"
)

func TestEventBus_Resource(t *testing.T) {
	eb := &EventBus{
		Name: testBusName,
		ARN: arn.ARN{
			Partition: testPartition,
			Service:   testService,
			Region:    testRegion,
			AccountID: testAccountID,
			Resource:  "event-bus/" + testBusName,
		},
	}
	assert.Equal(t, eb.Resource(), harness.Resource{Type: "AWS::Events::EventBus", PhysicalID: testBusName, ARN: eb.ARN.String()})
}

func TestGet(t *testing.T) {
	cases := map[string]struct {
		eventBusName            string
		mockDescribeEventBusAPI func(ctx context.Context, eventBusName, arn string) *MockDescribeEventBusAPI
		expectErr               error
		expect                  *EventBus
	}{
		"success": {
			eventBusName: testBusName,
			mockDescribeEventBusAPI: func(ctx context.Context, eventBusName, arn string) *MockDescribeEventBusAPI {
				mock := NewMockDescribeEventBusAPI(t)
				mock.EXPECT().
					DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{
						Name: aws.String(eventBusName),
					}).
					Return(&eventbridge.DescribeEventBusOutput{
						Arn: aws.String(arn),
					}, nil)
				return mock
			},
			expectErr: nil,
			expect: &EventBus{
				Name: testBusName,
				ARN: arn.ARN{
					Partition: testPartition,
					Service:   testService,
					Region:    testRegion,
					AccountID: testAccountID,
					Resource:  "event-bus/" + testBusName,
				},
			},
		},
		"failed": {
			eventBusName: testBusName,
			mockDescribeEventBusAPI: func(ctx context.Context, eventBusName, arn string) *MockDescribeEventBusAPI {
				mock := NewMockDescribeEventBusAPI(t)
				mock.EXPECT().
					DescribeEventBus(ctx, &eventbridge.DescribeEventBusInput{
						Name: aws.String(eventBusName),
					}).
					Return(nil, errors.New("failed"))
				return mock
			},
			expectErr: fmt.Errorf(`cannot get event bus "%v": failed`, testBusName),
			expect:    nil,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var arn string
			if tt.expect != nil {
				arn = tt.expect.ARN.String()
			}
			ctx := context.TODO()
			api := tt.mockDescribeEventBusAPI(ctx, tt.eventBusName, arn)
			actual, err := Get(ctx, api, tt.eventBusName)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, tt.expect, actual)
			}
		})
	}
}
