// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"context"
	"errors"
	"iatk/internal/pkg/aws/config"
	"iatk/internal/pkg/harness"
	"iatk/internal/pkg/harness/resource/eventbus"
	"iatk/internal/pkg/harness/resource/eventrule"
	"iatk/internal/pkg/harness/resource/queue"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

const (
	testBusName    string = "my-event-bus"
	testListenerID string = "iatk_eb_9m4e2mr0ui3e8a215n4g"
	testPartition  string = "aws"
	testService    string = "sqs"
	testRegion     string = "us-west-2"
	testAccountID  string = "123456789012"
	testQueueURL   string = "my-queue-url"
)

func testBusARN() arn.ARN {
	return arn.ARN{
		Partition: testPartition,
		Service:   testService,
		Region:    testRegion,
		AccountID: testAccountID,
		Resource:  "event-bus/" + testBusName,
	}
}

func testQueueARN() arn.ARN {
	return arn.ARN{
		Partition: testPartition,
		Service:   testService,
		Region:    testRegion,
		AccountID: testAccountID,
		Resource:  testListenerID,
	}
}

func testRuleARN() arn.ARN {
	return arn.ARN{
		Partition: testPartition,
		Service:   testService,
		Region:    testRegion,
		AccountID: testAccountID,
		Resource:  "rule/" + testBusName + "/" + testListenerID,
	}
}

func TestNew(t *testing.T) {
	cases := map[string]struct {
		mockGetEventBus       func(ctx context.Context, ebClient ebClient, eventBusName string, arn arn.ARN) *mockGetEventBusFunc
		mockGetRule           func(ctx context.Context, ebClient ebClient, ruleName string, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc
		mockListTargetsByRule func(ctx context.Context, ebClient ebClient, targetId, ruleName, eventBusName string) *mockListTargetsByRuleFunc
		eventBusName          string
		arn                   arn.ARN
		ruleName              string
		ruleARN               arn.ARN
		tags                  map[string]string
		expectErr             error
	}{
		"success": {
			eventBusName: testBusName,
			tags:         map[string]string{"foo": "bar"},
			arn:          testBusARN(),
			ruleName:     "InputRuleName",
			ruleARN:      testRuleARN(),
			expectErr:    nil,
			mockGetEventBus: func(ctx context.Context, ebClient ebClient, eventBusName string, arn arn.ARN) *mockGetEventBusFunc {
				mock := newMockGetEventBusFunc(t)
				mock.EXPECT().
					Execute(ctx, ebClient, eventBusName).
					Return(&eventbus.EventBus{
						Name: eventBusName,
						ARN:  arn,
					}, nil)
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				mock.EXPECT().Execute(ctx, ebClient, ruleName, eventBusName).Return(&eventrule.Rule{
					Name:         ruleName,
					EventBusName: eventBusName,
					ARN:          ruleARN,
				}, nil)
				return mock
			},
			mockListTargetsByRule: func(ctx context.Context, ebClient ebClient, targetId, ruleName, eventBusName string) *mockListTargetsByRuleFunc {
				mock := newMockListTargetsByRuleFunc(t)
				mock.EXPECT().
					Execute(ctx, ebClient, targetId, ruleName, eventBusName).
					Return(&ebtypes.Target{
						Input: aws.String("input"),
					}, nil)
				return mock
			},
		},
		"eventbus not found": {
			eventBusName: testBusName,
			arn:          testBusARN(),
			expectErr:    errors.New("failed to create resource group: event bus not found"),
			mockGetEventBus: func(ctx context.Context, ebClient ebClient, eventBusName string, arn arn.ARN) *mockGetEventBusFunc {
				mock := newMockGetEventBusFunc(t)
				mock.EXPECT().
					Execute(ctx, ebClient, eventBusName).
					Return(nil, errors.New("event bus not found"))
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				return mock
			},
			mockListTargetsByRule: func(ctx context.Context, ebClient ebClient, targetId, ruleName, eventBusName string) *mockListTargetsByRuleFunc {
				mock := newMockListTargetsByRuleFunc(t)
				return mock
			},
		},
		"getRule failed": {
			eventBusName: testBusName,
			arn:          testBusARN(),
			ruleName:     "InputRuleName",
			ruleARN:      testRuleARN(),
			expectErr:    errors.New("RuleName \"InputRuleName\" was provided but not found for eventbus \"my-event-bus\" failed: getRule failed"),
			mockGetEventBus: func(ctx context.Context, ebClient ebClient, eventBusName string, arn arn.ARN) *mockGetEventBusFunc {
				mock := newMockGetEventBusFunc(t)
				mock.EXPECT().
					Execute(ctx, ebClient, eventBusName).
					Return(&eventbus.EventBus{
						Name: eventBusName,
						ARN:  arn,
					}, nil)
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				mock.EXPECT().Execute(ctx, ebClient, ruleName, eventBusName).Return(nil, errors.New("getRule failed"))
				return mock
			},
			mockListTargetsByRule: func(ctx context.Context, ebClient ebClient, targetId, ruleName, eventBusName string) *mockListTargetsByRuleFunc {
				mock := newMockListTargetsByRuleFunc(t)
				return mock
			},
		},
		"ListTargetsByRule failed": {
			eventBusName: testBusName,
			arn:          testBusARN(),
			ruleName:     "InputRuleName",
			ruleARN:      testRuleARN(),
			expectErr:    errors.New("failed to create resource group: ListTargetByRule failed"),
			mockGetEventBus: func(ctx context.Context, ebClient ebClient, eventBusName string, arn arn.ARN) *mockGetEventBusFunc {
				mock := newMockGetEventBusFunc(t)
				mock.EXPECT().
					Execute(ctx, ebClient, eventBusName).
					Return(&eventbus.EventBus{
						Name: eventBusName,
						ARN:  arn,
					}, nil)
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				mock.EXPECT().Execute(ctx, ebClient, ruleName, eventBusName).Return(&eventrule.Rule{
					Name:         ruleName,
					EventBusName: eventBusName,
					ARN:          ruleARN,
				}, nil)
				return mock
			},
			mockListTargetsByRule: func(ctx context.Context, ebClient ebClient, targetId, ruleName, eventBusName string) *mockListTargetsByRuleFunc {
				mock := newMockListTargetsByRuleFunc(t)
				mock.EXPECT().
					Execute(ctx, ebClient, targetId, ruleName, eventBusName).
					Return(nil, errors.New("ListTargetByRule failed"))
				return mock
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			opts := Options{
				ebClient: newMockEbClient(t),
			}
			getEventBus := tt.mockGetEventBus(ctx, opts.ebClient, tt.eventBusName, tt.arn)
			getRule := tt.mockGetRule(ctx, opts.ebClient, tt.ruleName, tt.eventBusName, tt.ruleARN)
			listTargetsByRule := tt.mockListTargetsByRule(ctx, opts.ebClient, "", tt.ruleName, tt.eventBusName)
			opts.getEventBus = getEventBus.Execute
			opts.getRule = getRule.Execute
			opts.listTargetsByRule = listTargetsByRule.Execute
			expectEventBus := &eventbus.EventBus{
				Name: tt.eventBusName,
				ARN:  tt.arn,
			}
			listener, err := New(ctx, tt.eventBusName, "", tt.ruleName, tt.tags, opts)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, expectEventBus, listener.eventBus)
				assert.Equal(t, tt.tags, listener.customTags)
				assert.NotNil(t, listener.opts)
			}
		})
	}
}

func TestGet(t *testing.T) {
	cases := map[string]struct {
		mockGetQueueWithName         func(ctx context.Context, sqsClient sqsClient, queueName string, queueURL string, queueARN arn.ARN) *mockGetQueueWithNameFunc
		mockGetEventBusNameFromQueue func(ctx context.Context, sqslClient sqsClient, queueURL string, eventBusName string) *mockGetEventBusNameFromQueueFunc
		mockGetRule                  func(ctx context.Context, ebClient ebClient, ruleName string, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc
		listenerID                   string
		queueName                    string
		queueURL                     string
		queueARN                     arn.ARN
		ruleName                     string
		eventBusName                 string
		ruleARN                      arn.ARN
		expectErr                    error
	}{
		"success": {
			listenerID:   testListenerID,
			expectErr:    nil,
			queueName:    testListenerID,
			queueURL:     testQueueURL,
			queueARN:     testQueueARN(),
			ruleName:     testListenerID,
			eventBusName: testBusName,
			ruleARN:      testRuleARN(),
			mockGetQueueWithName: func(ctx context.Context, sqsClient sqsClient, queueName, queueURL string, queueARN arn.ARN) *mockGetQueueWithNameFunc {
				mock := newMockGetQueueWithNameFunc(t)
				mock.EXPECT().
					Execute(ctx, sqsClient, queueName).
					Return(&queue.Queue{
						Name:     queueName,
						QueueURL: queueURL,
						ARN:      queueARN,
					}, nil)
				return mock
			},
			mockGetEventBusNameFromQueue: func(ctx context.Context, sqsClient sqsClient, queueURL, eventBusName string) *mockGetEventBusNameFromQueueFunc {
				mock := newMockGetEventBusNameFromQueueFunc(t)
				mock.EXPECT().Execute(ctx, sqsClient, queueURL).Return(eventBusName, nil)
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				mock.EXPECT().Execute(ctx, ebClient, ruleName, eventBusName).Return(&eventrule.Rule{
					Name:         ruleName,
					EventBusName: eventBusName,
					ARN:          ruleARN,
				}, nil)
				return mock
			},
		},
		"failed due to invalid id": {
			listenerID: "",
			expectErr:  errors.New("invalid ID"),
			mockGetQueueWithName: func(ctx context.Context, sqsClient sqsClient, queueName, queueURL string, queueARN arn.ARN) *mockGetQueueWithNameFunc {
				mock := newMockGetQueueWithNameFunc(t)
				return mock
			},
			mockGetEventBusNameFromQueue: func(ctx context.Context, sqsClient sqsClient, queueURL, eventBusName string) *mockGetEventBusNameFromQueueFunc {
				mock := newMockGetEventBusNameFromQueueFunc(t)
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				return mock
			},
		},
		"failed due to failure to get queue": {
			listenerID: testListenerID,
			queueName:  testListenerID,
			expectErr:  errors.New("faied to get eb listener iatk_eb_9m4e2mr0ui3e8a215n4g: permission denied"),
			mockGetQueueWithName: func(ctx context.Context, sqsClient sqsClient, queueName, queueURL string, queueARN arn.ARN) *mockGetQueueWithNameFunc {
				mock := newMockGetQueueWithNameFunc(t)
				mock.EXPECT().
					Execute(ctx, sqsClient, queueName).
					Return(nil, errors.New("permission denied"))
				return mock
			},
			mockGetEventBusNameFromQueue: func(ctx context.Context, sqsClient sqsClient, queueURL, eventBusName string) *mockGetEventBusNameFromQueueFunc {
				mock := newMockGetEventBusNameFromQueueFunc(t)
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				return mock
			},
		},
		"cannot get event bus name from queue; succeed with nil rule": {
			listenerID: testListenerID,
			expectErr:  nil,
			queueName:  testListenerID,
			queueURL:   testQueueURL,
			queueARN:   testQueueARN(),
			mockGetQueueWithName: func(ctx context.Context, sqsClient sqsClient, queueName, queueURL string, queueARN arn.ARN) *mockGetQueueWithNameFunc {
				mock := newMockGetQueueWithNameFunc(t)
				mock.EXPECT().
					Execute(ctx, sqsClient, queueName).
					Return(&queue.Queue{
						Name:     queueName,
						QueueURL: queueURL,
						ARN:      queueARN,
					}, nil)
				return mock
			},
			mockGetEventBusNameFromQueue: func(ctx context.Context, sqsClient sqsClient, queueURL, eventBusName string) *mockGetEventBusNameFromQueueFunc {
				mock := newMockGetEventBusNameFromQueueFunc(t)
				mock.EXPECT().Execute(ctx, sqsClient, queueURL).Return("", errors.New("tag not found"))
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				return mock
			},
		},
		"cannot find rule; succeed with nil rule": {
			listenerID:   testListenerID,
			expectErr:    nil,
			queueName:    testListenerID,
			queueURL:     testQueueURL,
			queueARN:     testQueueARN(),
			ruleName:     testListenerID,
			eventBusName: testBusName,
			mockGetQueueWithName: func(ctx context.Context, sqsClient sqsClient, queueName, queueURL string, queueARN arn.ARN) *mockGetQueueWithNameFunc {
				mock := newMockGetQueueWithNameFunc(t)
				mock.EXPECT().
					Execute(ctx, sqsClient, queueName).
					Return(&queue.Queue{
						Name:     queueName,
						QueueURL: queueURL,
						ARN:      queueARN,
					}, nil)
				return mock
			},
			mockGetEventBusNameFromQueue: func(ctx context.Context, sqsClient sqsClient, queueURL, eventBusName string) *mockGetEventBusNameFromQueueFunc {
				mock := newMockGetEventBusNameFromQueueFunc(t)
				mock.EXPECT().Execute(ctx, sqsClient, queueURL).Return(eventBusName, nil)
				return mock
			},
			mockGetRule: func(ctx context.Context, ebClient ebClient, ruleName, eventBusName string, ruleARN arn.ARN) *mockGetRuleFunc {
				mock := newMockGetRuleFunc(t)
				mock.EXPECT().Execute(ctx, ebClient, ruleName, eventBusName).Return(nil, errors.New("rule not found"))
				return mock
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			sqsClient := newMockSqsClient(t)
			ebClient := newMockEbClient(t)
			mockGetQueueWithName := tt.mockGetQueueWithName(ctx, sqsClient, tt.queueName, tt.queueURL, tt.queueARN)
			mockGetEventBusNameFromQueue := tt.mockGetEventBusNameFromQueue(ctx, sqsClient, tt.queueURL, tt.eventBusName)
			mockGetRule := tt.mockGetRule(ctx, ebClient, tt.ruleName, tt.eventBusName, tt.ruleARN)
			opts := Options{
				sqsClient: sqsClient,
				ebClient:  ebClient,

				getQueueWithName:         mockGetQueueWithName.Execute,
				getEventBusNameFromQueue: mockGetEventBusNameFromQueue.Execute,
				getRule:                  mockGetRule.Execute,
			}
			lr, err := Get(ctx, tt.listenerID, opts)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				expectQueue := &queue.Queue{
					Name:     tt.queueName,
					QueueURL: tt.queueURL,
					ARN:      tt.queueARN,
				}
				assert.Equal(t, expectQueue, lr.queue)

				zeroARN := arn.ARN{}
				if tt.ruleARN != zeroARN {
					expectRule := &eventrule.Rule{
						Name:         tt.ruleName,
						EventBusName: tt.eventBusName,
						ARN:          tt.ruleARN,
					}
					assert.Equal(t, expectRule, lr.rule)
				} else {
					assert.Nil(t, lr.rule)
				}
			}
		})
	}
}

func Test_destroySingle(t *testing.T) {
	cases := map[string]struct {
		mockResourceGroupDestroyer func(context.Context) *mockDestroyer
		expectErr                  error
	}{
		"success": {
			mockResourceGroupDestroyer: func(ctx context.Context) *mockDestroyer {
				lr := newMockDestroyer(t)
				lr.EXPECT().ID().Return("my-listner-id")
				lr.EXPECT().Destroy(ctx).Return(nil)
				return lr
			},
			expectErr: nil,
		},
		"destroy failed": {
			mockResourceGroupDestroyer: func(ctx context.Context) *mockDestroyer {
				lr := newMockDestroyer(t)
				lr.EXPECT().ID().Return("my-listner-id")
				lr.EXPECT().Destroy(ctx).Return(errors.New("destroy failed"))
				lr.EXPECT().Components().Return([]harness.Resource{
					{Type: "AWS::Events::Rule", PhysicalID: "my-rule", ARN: "arn"},
				})
				return lr
			},
			expectErr: errors.New(`failed to destroy eb listener "my-listner-id": destroy failed`),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			rg := tt.mockResourceGroupDestroyer(context.TODO())
			err := destroySingle(context.TODO(), rg)
			if err != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			}
		})
	}

}

func TestDestroyMultiple(t *testing.T) {
	cases := map[string]struct {
		mockGet           func(ctx context.Context, lrs map[string]*Listener) *MockGetFunc
		mockDestroySingle func(ctx context.Context, lrs map[string]*Listener) *mockDestroySingleFunc
		listeners         map[string]*Listener
		listenerIDs       []string
		expectErr         error
	}{
		"success": {
			listenerIDs: []string{
				"iatk_eb_listener-1",
				"iatk_eb_listener-2",
				"iatk_eb_listener-3",
			},
			listeners: map[string]*Listener{
				"iatk_eb_listener-1": &Listener{},
				"iatk_eb_listener-2": &Listener{},
				"iatk_eb_listener-3": &Listener{},
			},
			expectErr: nil,
			mockGet: func(ctx context.Context, lrs map[string]*Listener) *MockGetFunc {
				f := NewMockGetFunc(t)
				for id, lr := range lrs {
					f.EXPECT().Execute(ctx, id, mock.AnythingOfType("Options")).Return(lr, nil)
				}
				return f
			},
			mockDestroySingle: func(ctx context.Context, lrs map[string]*Listener) *mockDestroySingleFunc {
				f := newMockDestroySingleFunc(t)
				for _, lr := range lrs {
					f.EXPECT().Execute(ctx, lr).Return(nil)
				}
				return f
			},
		},
		"succeed with duplicated resource group ids": {
			listenerIDs: []string{
				"iatk_eb_listener-1",
				"iatk_eb_listener-1",
				"iatk_eb_listener-2",
				"iatk_eb_listener-2",
				"iatk_eb_listener-3",
			},
			listeners: map[string]*Listener{
				"iatk_eb_listener-1": &Listener{},
				"iatk_eb_listener-2": &Listener{},
				"iatk_eb_listener-3": &Listener{},
			},
			expectErr: nil,
			mockGet: func(ctx context.Context, lrs map[string]*Listener) *MockGetFunc {
				f := NewMockGetFunc(t)
				for id, lr := range lrs {
					f.EXPECT().Execute(ctx, id, mock.AnythingOfType("Options")).Return(lr, nil)
				}
				return f
			},
			mockDestroySingle: func(ctx context.Context, lrs map[string]*Listener) *mockDestroySingleFunc {
				f := newMockDestroySingleFunc(t)
				for _, lr := range lrs {
					f.EXPECT().Execute(ctx, lr).Return(nil)
				}
				return f
			},
		},
		"one of provided listener could not be get": {
			listenerIDs: []string{
				"iatk_eb_listener-1",
				"iatk_eb_listener-2",
				"iatk_eb_listener-3",
			},
			listeners: map[string]*Listener{
				"iatk_eb_listener-1": &Listener{},
				"iatk_eb_listener-2": nil,
				"iatk_eb_listener-3": &Listener{},
			},
			expectErr: errors.New("failed to destroy following listener(s): {resource group id: iatk_eb_listener-2, reason: invalid ID}"),
			mockGet: func(ctx context.Context, lrs map[string]*Listener) *MockGetFunc {
				f := NewMockGetFunc(t)
				for id, lr := range lrs {
					if lr != nil {
						f.EXPECT().Execute(ctx, id, mock.AnythingOfType("Options")).Return(lr, nil)
					} else {
						f.EXPECT().Execute(ctx, id, mock.AnythingOfType("Options")).Return(nil, errors.New("invalid ID"))
					}
				}
				return f
			},
			mockDestroySingle: func(ctx context.Context, lrs map[string]*Listener) *mockDestroySingleFunc {
				f := newMockDestroySingleFunc(t)
				for _, lr := range lrs {
					if lr != nil {
						f.EXPECT().Execute(ctx, lr).Return(nil)
					}
				}
				return f
			},
		},
		"fail to destroy two of the groups": {
			listenerIDs: []string{
				"iatk_eb_listener-1",
				"iatk_eb_listener-2",
				"iatk_eb_listener-3",
			},
			listeners: map[string]*Listener{
				"iatk_eb_listener-1": {id: "listener-1"},
				"iatk_eb_listener-2": {id: "listener-2"},
				"iatk_eb_listener-3": {id: "listener-3"},
			},
			expectErr: errors.New("failed to destroy following listener(s): {resource group id: iatk_eb_listener-2, reason: api failed}, {resource group id: iatk_eb_listener-3, reason: api failed}"),
			mockGet: func(ctx context.Context, lrs map[string]*Listener) *MockGetFunc {
				f := NewMockGetFunc(t)
				for id, lr := range lrs {
					f.EXPECT().Execute(ctx, id, mock.AnythingOfType("Options")).Return(lr, nil)
				}
				return f
			},
			mockDestroySingle: func(ctx context.Context, lrs map[string]*Listener) *mockDestroySingleFunc {
				f := newMockDestroySingleFunc(t)
				for id, lr := range lrs {
					if id == "iatk_eb_listener-1" {
						f.EXPECT().Execute(ctx, lr).Return(nil)
					} else {
						f.EXPECT().Execute(ctx, lr).Return(errors.New("api failed"))
					}
				}
				return f
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			mockGet := tt.mockGet(ctx, tt.listeners)
			mockDestroySingle := tt.mockDestroySingle(ctx, tt.listeners)
			opts := destroyMultipleOptions{
				destroySingle: mockDestroySingle.Execute,
				Get:           mockGet.Execute,
			}
			cfg, err := config.GetAWSConfig(context.TODO(), "us-west-2", "default", nil)
			if err != nil {
				t.Fatalf("error when loading AWS config: %v", err)
			}
			err = DestroyMultiple(ctx, tt.listenerIDs, cfg, opts)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestCreate(t *testing.T) {
	cases := map[string]struct {
		mockDeployer func(ctx context.Context, expectOutput Output) *mockDeployer
		expectErr    error
		expectOutput *Output
	}{
		"success": {
			mockDeployer: func(ctx context.Context, expectOutput Output) *mockDeployer {
				d := newMockDeployer(t)
				d.EXPECT().ID().Return("my-eb-listener")
				d.EXPECT().JSON().Return(expectOutput)
				d.EXPECT().Deploy(ctx).Return(nil)
				return d
			},
			expectErr: nil,
			expectOutput: &Output{
				ID: "my-eb-listener",
				TestTarget: harness.Resource{
					Type:       "AWS::Events::EventBus",
					PhysicalID: testBusName,
					ARN:        "my-eventbus-arn",
				},
				Components: []harness.Resource{
					{Type: "AWS::SQS::Queue", PhysicalID: testQueueURL, ARN: "my-queue-arn"},
					{Type: "AWS::Events::Rule", PhysicalID: "my-rule-id", ARN: "my-rule-arn"},
				},
			},
		},
		"deploy failed, destroy success": {
			mockDeployer: func(ctx context.Context, expectOutput Output) *mockDeployer {
				d := newMockDeployer(t)
				d.EXPECT().ID().Return("my-eb-listener")
				d.EXPECT().Deploy(ctx).Return(errors.New("deploy failed"))
				d.EXPECT().Destroy(ctx).Return(nil)
				return d
			},
			expectErr:    errors.New("failed to create eb listener my-eb-listener: deploy failed"),
			expectOutput: nil,
		},
		"deploy failed, destroy failed": {
			mockDeployer: func(ctx context.Context, expectOutput Output) *mockDeployer {
				d := newMockDeployer(t)
				d.EXPECT().ID().Return("my-eb-listener")
				d.EXPECT().Deploy(ctx).Return(errors.New("deploy failed"))
				d.EXPECT().Destroy(ctx).Return(errors.New("destroy failed"))
				d.EXPECT().Components().Return([]harness.Resource{
					{Type: "AWS::SQS::Queue", PhysicalID: "my-queue", ARN: "arn"},
				})
				return d
			},
			expectErr:    errors.New("failed to create eb listener my-eb-listener: deploy failed"),
			expectOutput: nil,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var expectOutput Output
			if tt.expectOutput != nil {
				expectOutput = *tt.expectOutput
			}
			lr := tt.mockDeployer(context.TODO(), expectOutput)
			output, err := Create(context.TODO(), lr)
			if err != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, output, tt.expectOutput)
			}
		})
	}
}

func Test_isValidId(t *testing.T) {
	cases := []struct {
		name   string
		id     string
		expect bool
	}{
		{"valid", testListenerID, true},
		{"empty string", "", false},
		{"invalid prefix", "eb_iatk_9m4e2mr0ui3e8a215n4g", false},
		{"invalid suffix", "iatk_eb_xxxxxxxxxxxxxxxxxxxx", false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := isValidID(tt.id)
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestPollEvents(t *testing.T) {
	cases := map[string]struct {
		waitTimeSeconds     int32
		maxNumberOfMessages int32
		mock                func(ctx context.Context, waitTimeSeconds, maxNumberOfMessages int32) *mockPoller
		expect              []string
		expectErr           error
	}{
		"should succeed": {
			waitTimeSeconds:     5,
			maxNumberOfMessages: 5,
			mock: func(ctx context.Context, waitTimeSeconds, maxNumberOfMessages int32) *mockPoller {
				m := newMockPoller(t)
				m.EXPECT().
					ReceiveEvents(ctx, waitTimeSeconds, maxNumberOfMessages).
					Return([]Event{
						{Body: "{}", ReceiptHandle: "123"},
						{Body: "{}", ReceiptHandle: "456"},
						{Body: `{"foo":"bar"}`, ReceiptHandle: "789"},
					}, nil)
				m.EXPECT().
					DeleteEvents(ctx, []string{"123", "456", "789"}).
					Return(nil)
				return m
			},
			expect:    []string{"{}", "{}", `{"foo":"bar"}`},
			expectErr: nil,
		},
		"should succeed and not calling delete events if not event is received": {
			waitTimeSeconds:     5,
			maxNumberOfMessages: 5,
			mock: func(ctx context.Context, waitTimeSeconds, maxNumberOfMessages int32) *mockPoller {
				m := newMockPoller(t)
				m.EXPECT().
					ReceiveEvents(ctx, waitTimeSeconds, maxNumberOfMessages).
					Return([]Event{}, nil)
				return m
			},
			expect:    []string{},
			expectErr: nil,
		},
		"should fail due to ReceiveEvents failure": {
			waitTimeSeconds:     5,
			maxNumberOfMessages: 5,
			mock: func(ctx context.Context, waitTimeSeconds, maxNumberOfMessages int32) *mockPoller {
				m := newMockPoller(t)
				m.EXPECT().
					ReceiveEvents(ctx, waitTimeSeconds, maxNumberOfMessages).
					Return(nil, errors.New("receive events failed"))
				return m
			},
			expect:    nil,
			expectErr: errors.New("failed to poll events: receive events failed"),
		},
		"should fail due to DeleteEvents failure": {
			waitTimeSeconds:     5,
			maxNumberOfMessages: 5,
			mock: func(ctx context.Context, waitTimeSeconds, maxNumberOfMessages int32) *mockPoller {
				m := newMockPoller(t)
				m.EXPECT().
					ReceiveEvents(ctx, waitTimeSeconds, maxNumberOfMessages).
					Return([]Event{
						{Body: "{}", ReceiptHandle: "123"},
						{Body: "{}", ReceiptHandle: "456"},
						{Body: `{"foo":"bar"}`, ReceiptHandle: "789"},
					}, nil)
				m.EXPECT().
					DeleteEvents(ctx, []string{"123", "456", "789"}).
					Return(errors.New("delete events failed"))
				return m
			},
			expect:    nil,
			expectErr: errors.New("failed to poll events: delete events failed"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			listener := tt.mock(ctx, tt.waitTimeSeconds, tt.maxNumberOfMessages)
			actual, err := PollEvents(ctx, listener, tt.waitTimeSeconds, tt.maxNumberOfMessages)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, tt.expect, actual)
			}
		})
	}
}
