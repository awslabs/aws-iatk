// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package queue

import (
	"context"
	"errors"
	"iatk/internal/pkg/harness"
	"iatk/internal/pkg/harness/tags"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testQueueName string = "my-queue"
	testQueueURL  string = "my-queue-url"
	testPartition string = "aws"
	testService   string = "sqs"
	testRegion    string = "us-west-2"
	testAccountID string = "123456789012"
)

func TestCreate(t *testing.T) {
	cases := map[string]struct {
		name      string
		mockAPI   func(ctx context.Context, expect *Queue) *MockCreateQueueAPI
		tags      map[string]string
		options   Options
		expect    *Queue
		expectErr error
	}{
		"should create queue": {
			expect: &Queue{
				Name:     testQueueName,
				QueueURL: testQueueURL,
				ARN: arn.ARN{
					Partition: testPartition,
					Service:   testService,
					Region:    testRegion,
					AccountID: testAccountID,
					Resource:  testQueueName,
				},
			},
			expectErr: nil,
			name:      testQueueName,
			tags:      map[string]string{"key": "value"},
			options: Options{
				Policy:                 "{}",
				MessageRetentionPeriod: 3,
			},
			mockAPI: func(ctx context.Context, expect *Queue) *MockCreateQueueAPI {
				api := NewMockCreateQueueAPI(t)
				api.EXPECT().
					CreateQueue(ctx, &sqs.CreateQueueInput{
						QueueName: aws.String(expect.Name),
						Tags:      map[string]string{"key": "value"},
						Attributes: map[string]string{
							"Policy":                 "{}",
							"MessageRetentionPeriod": "3",
						},
					}).
					Return(&sqs.CreateQueueOutput{
						QueueUrl: &expect.QueueURL,
					}, nil)
				api.EXPECT().GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
					QueueUrl: aws.String(expect.QueueURL),
					AttributeNames: []sqstypes.QueueAttributeName{
						sqstypes.QueueAttributeNameQueueArn,
					},
				}).Return(&sqs.GetQueueAttributesOutput{
					Attributes: map[string]string{"QueueArn": expect.ARN.String()},
				}, nil)
				return api
			},
		},
		"should return error if CreateQueue failed": {
			expect:    nil,
			expectErr: errors.New(`create queue "my-queue" failed: something failed`),
			name:      testQueueName,
			tags:      map[string]string{"key": "value"},
			options: Options{
				Policy:                 "{}",
				MessageRetentionPeriod: 3,
			},
			mockAPI: func(ctx context.Context, expect *Queue) *MockCreateQueueAPI {
				api := NewMockCreateQueueAPI(t)
				api.EXPECT().CreateQueue(ctx, mock.AnythingOfType("*sqs.CreateQueueInput")).Return(nil, errors.New("something failed"))
				return api
			},
		},
		"should return error if GetQueueAttributes failed": {
			expect:    nil,
			expectErr: errors.New("get queue attributes failed: something failed"),
			name:      testQueueName,
			tags:      map[string]string{"key": "value"},
			options: Options{
				Policy:                 "{}",
				MessageRetentionPeriod: 3,
			},
			mockAPI: func(ctx context.Context, expect *Queue) *MockCreateQueueAPI {
				api := NewMockCreateQueueAPI(t)
				api.EXPECT().CreateQueue(ctx, mock.AnythingOfType("*sqs.CreateQueueInput")).Return(&sqs.CreateQueueOutput{
					QueueUrl: aws.String("does not matter"),
				}, nil)
				api.EXPECT().GetQueueAttributes(ctx, mock.AnythingOfType("*sqs.GetQueueAttributesInput")).Return(nil, errors.New("something failed"))
				return api
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			api := tt.mockAPI(ctx, tt.expect)
			queue, err := Create(ctx, api, tt.name, tt.tags, tt.options)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, tt.expect, queue)
			}
			api.AssertExpectations(t)
		})
	}
}

func TestDelete(t *testing.T) {
	cases := map[string]struct {
		queueURL  string
		mockAPI   func(ctx context.Context, queueURL string) *MockDeleteQueueAPI
		expectErr error
	}{
		"Should delete queue successfully": {
			queueURL:  testQueueURL,
			expectErr: nil,
			mockAPI: func(ctx context.Context, queueURL string) *MockDeleteQueueAPI {
				m := NewMockDeleteQueueAPI(t)
				m.EXPECT().
					DeleteQueue(ctx, &sqs.DeleteQueueInput{
						QueueUrl: aws.String(queueURL),
					}).
					Return(&sqs.DeleteQueueOutput{}, nil)
				return m
			},
		},
		"Should return error if api call failed": {
			queueURL:  testQueueURL,
			expectErr: errors.New(`failed to delete queue "my-queue-url": something failed`),
			mockAPI: func(ctx context.Context, queueURL string) *MockDeleteQueueAPI {
				m := NewMockDeleteQueueAPI(t)
				m.EXPECT().
					DeleteQueue(ctx, &sqs.DeleteQueueInput{
						QueueUrl: aws.String(queueURL),
					}).
					Return(nil, errors.New("something failed"))
				return m
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			mockAPI := tt.mockAPI(ctx, tt.queueURL)
			err := Delete(ctx, mockAPI, tt.queueURL)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Nil(t, err)
			}
			mockAPI.AssertExpectations(t)
		})
	}
}

func TestGetWithName(t *testing.T) {
	cases := map[string]struct {
		mockAPI   func(ctx context.Context, name string) *MockGetQueueUrlAPI
		queueName string
		expectErr error
		expect    *Queue
	}{
		"success": {
			queueName: testQueueName,
			expectErr: nil,
			expect: &Queue{
				Name:     testQueueName,
				QueueURL: testQueueURL,
				ARN: arn.ARN{
					Partition: testPartition,
					Service:   testService,
					Region:    testRegion,
					AccountID: testAccountID,
					Resource:  testQueueName,
				},
			},
			mockAPI: func(ctx context.Context, name string) *MockGetQueueUrlAPI {
				api := NewMockGetQueueUrlAPI(t)
				api.EXPECT().
					GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
						QueueName: aws.String(name),
					}).
					Return(&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String(testQueueURL),
					}, nil)
				api.EXPECT().
					GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
						QueueUrl: aws.String(testQueueURL),
						AttributeNames: []sqstypes.QueueAttributeName{
							sqstypes.QueueAttributeNameQueueArn,
						},
					}).
					Return(&sqs.GetQueueAttributesOutput{
						Attributes: map[string]string{"QueueArn": arn.ARN{
							Partition: testPartition,
							Service:   testService,
							Region:    testRegion,
							AccountID: testAccountID,
							Resource:  testQueueName,
						}.String()},
					}, nil)
				return api
			},
		},
		"failed to get queue url": {
			queueName: testQueueName,
			expectErr: errors.New(`failed to get queue with name "my-queue": error on aws`),
			expect:    nil,
			mockAPI: func(ctx context.Context, name string) *MockGetQueueUrlAPI {
				api := NewMockGetQueueUrlAPI(t)
				api.EXPECT().
					GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
						QueueName: aws.String(name),
					}).
					Return(nil, errors.New("error on aws"))
				return api
			},
		},
		"failed to get queue attributes": {
			queueName: testQueueName,
			expectErr: errors.New(`failed to get queue attributes: error on aws`),
			expect:    nil,
			mockAPI: func(ctx context.Context, name string) *MockGetQueueUrlAPI {
				api := NewMockGetQueueUrlAPI(t)
				api.EXPECT().
					GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
						QueueName: aws.String(name),
					}).
					Return(&sqs.GetQueueUrlOutput{
						QueueUrl: aws.String(testQueueURL),
					}, nil)
				api.EXPECT().
					GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
						QueueUrl: aws.String(testQueueURL),
						AttributeNames: []sqstypes.QueueAttributeName{
							sqstypes.QueueAttributeNameQueueArn,
						},
					}).
					Return(nil, errors.New("error on aws"))
				return api
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			api := tt.mockAPI(context.TODO(), tt.queueName)
			q, err := GetWithName(context.TODO(), api, tt.queueName)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expect, q)
			}
		})
	}
}

func TestGetEventBusNameFromQueue(t *testing.T) {
	cases := map[string]struct {
		queueURL  string
		mockAPI   func(ctx context.Context, queueURL, eventBusName string) *MockListQueueTagsAPI
		expect    string
		expectErr error
	}{
		"success": {
			queueURL:  testQueueURL,
			expect:    "my-event-bus",
			expectErr: nil,
			mockAPI: func(ctx context.Context, queueURL, eventBusName string) *MockListQueueTagsAPI {
				api := NewMockListQueueTagsAPI(t)
				api.EXPECT().
					ListQueueTags(ctx, &sqs.ListQueueTagsInput{
						QueueUrl: aws.String(queueURL),
					}).
					Return(&sqs.ListQueueTagsOutput{
						Tags: map[string]string{string(tags.TestHarnessTarget): eventBusName},
					}, nil)
				return api
			},
		},
		"api failed": {
			queueURL:  testQueueURL,
			expect:    "",
			expectErr: errors.New("failed to list queue tags: api failed"),
			mockAPI: func(ctx context.Context, queueURL, eventBusName string) *MockListQueueTagsAPI {
				api := NewMockListQueueTagsAPI(t)
				api.EXPECT().
					ListQueueTags(ctx, &sqs.ListQueueTagsInput{
						QueueUrl: aws.String(queueURL),
					}).
					Return(nil, errors.New("api failed"))
				return api
			},
		},
		"no tag": {
			queueURL:  testQueueURL,
			expect:    "",
			expectErr: errors.New("cannot get event bus name from queue"),
			mockAPI: func(ctx context.Context, queueURL, eventBusName string) *MockListQueueTagsAPI {
				api := NewMockListQueueTagsAPI(t)
				api.EXPECT().
					ListQueueTags(ctx, &sqs.ListQueueTagsInput{
						QueueUrl: aws.String(queueURL),
					}).
					Return(&sqs.ListQueueTagsOutput{
						Tags: map[string]string{},
					}, nil)
				return api
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			api := tt.mockAPI(ctx, tt.queueURL, tt.expect)
			actual, err := GetEventBusNameFromQueue(ctx, api, tt.queueURL)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			}
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestQueue_Resource(t *testing.T) {
	queue := &Queue{
		Name:     testQueueName,
		QueueURL: testQueueURL,
		ARN: arn.ARN{
			Partition: testPartition,
			Service:   testService,
			Region:    testRegion,
			AccountID: testAccountID,
			Resource:  testQueueName,
		},
	}
	assert.Equal(t, queue.Resource(), harness.Resource{Type: "AWS::SQS::Queue", PhysicalID: testQueueURL, ARN: queue.ARN.String()})
}
