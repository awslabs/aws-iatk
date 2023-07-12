// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
	"zion/internal/pkg/harness"
	"zion/internal/pkg/harness/resource/eventbus"
	"zion/internal/pkg/harness/resource/eventrule"
	"zion/internal/pkg/harness/resource/queue"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListener_ID(t *testing.T) {
	id := xid.New().String()
	lr := &Listener{id: id}
	expect := fmt.Sprintf("zion_eb_%v", id)
	actual := lr.ID()
	assert.Equal(t, expect, actual)
}

func TestListener_tags(t *testing.T) {
	eventBusName := "physical-id"
	eventBusARN := testBusARN()
	id := xid.New().String()
	userProvidedTags := map[string]string{
		"foo": "bar",
		"bax": "baz",
	}
	lr := &Listener{
		id:         id,
		eventBus:   &eventbus.EventBus{Name: eventBusName, ARN: eventBusARN},
		customTags: userProvidedTags,
	}
	ts := time.Now()
	actual := lr.tags(ts)
	expect := map[string]string{
		"zion:TestHarness:ID":      lr.ID(),
		"zion:TestHarness:Type":    TestHarnessType,
		"zion:TestHarness:Target":  eventBusARN.String(),
		"zion:TestHarness:Created": ts.Format(time.RFC3339),
		"foo":                      "bar",
		"bax":                      "baz",
	}
	assert.Equal(t, expect, actual)
}

func TestListener_String(t *testing.T) {
	eventBusName := testBusName
	id := xid.New().String()
	lr := &Listener{
		id:       id,
		eventBus: &eventbus.EventBus{Name: eventBusName},
	}
	actual := lr.String()
	expect := fmt.Sprintf("eb listener id: zion_eb_%v", id)
	assert.Equal(t, expect, actual)
}

func TestListener_Components(t *testing.T) {
	rname := "my-rule"
	rarn := testRuleARN()
	qurl := "my-queue-url"
	qarn := testQueueARN()
	r := &eventrule.Rule{Name: rname, ARN: rarn}
	q := &queue.Queue{QueueURL: qurl, ARN: qarn}
	lr := &Listener{rule: r, queue: q}
	actual := lr.Components()
	expect := []harness.Resource{
		{Type: "AWS::SQS::Queue", PhysicalID: qurl, ARN: qarn.String()},
		{Type: "AWS::Events::Rule", PhysicalID: rname, ARN: rarn.String()},
	}
	assert.Equal(t, expect, actual)
}

func TestListener_Deploy(t *testing.T) {
	cases := map[string]struct {
		lr                 *Listener
		mockCreateQueue    func(ctx context.Context, lr *Listener, queue *queue.Queue) *mockCreateQueueFunc
		mockCreateRule     func(ctx context.Context, lr *Listener, rule *eventrule.Rule) *mockCreateRuleFunc
		mockPutQueueTarget func(ctx context.Context, lr *Listener, queue *queue.Queue, rule *eventrule.Rule) *mockPutQueueTargetFunc
		expectQueue        *queue.Queue
		expectRule         *eventrule.Rule
		expectErr          func(lr *Listener) error
	}{
		"should complete deploy": {
			lr: &Listener{
				id:           xid.New().String(),
				eventBus:     &eventbus.EventBus{Name: "my-event-bus"},
				eventPattern: "{}",
				input:        "",
				inputPath:    "",
				inputTransformer: &eventrule.InputTransformer{
					InputTemplate: "<instance> is in state <status>",
					InputPathsMap: map[string]string{
						"instance": "$.detail.instance",
						"status":   "$.detail.status",
					},
				},
				opts: Options{
					ebClient:  newMockEbClient(t),
					sqsClient: newMockSqsClient(t),
				},
			},
			mockCreateQueue: func(ctx context.Context, lr *Listener, queue *queue.Queue) *mockCreateQueueFunc {
				m := newMockCreateQueueFunc(t)
				m.EXPECT().Execute(ctx, lr.opts.sqsClient, lr.ID(), mock.AnythingOfType("map[string]string"), mock.AnythingOfType("queue.Options")).Return(queue, nil)
				return m
			},
			mockCreateRule: func(ctx context.Context, lr *Listener, rule *eventrule.Rule) *mockCreateRuleFunc {
				m := newMockCreateRuleFunc(t)
				description := fmt.Sprintf("rule for Listener %q; created by zion", lr.ID())
				m.EXPECT().
					Execute(ctx, lr.opts.ebClient, lr.ID(), lr.eventBus.Name, lr.eventPattern, description, mock.AnythingOfType("map[string]string")).
					Return(rule, nil)
				return m
			},
			mockPutQueueTarget: func(ctx context.Context, lr *Listener, queue *queue.Queue, rule *eventrule.Rule) *mockPutQueueTargetFunc {
				m := newMockPutQueueTargetFunc(t)
				m.EXPECT().
					Execute(ctx, lr.opts.ebClient, lr.ID(), queue, rule, lr.input, lr.inputPath, lr.inputTransformer).
					Return(nil)
				return m
			},
			expectQueue: &queue.Queue{QueueURL: ""},
			expectRule:  &eventrule.Rule{},
			expectErr:   nil,
		},
		"should return error if createQueue failed": {
			lr: &Listener{
				id:           xid.New().String(),
				eventBus:     &eventbus.EventBus{Name: "my-event-bus"},
				eventPattern: "{}",
				input:        "",
				inputPath:    "",
				inputTransformer: &eventrule.InputTransformer{
					InputTemplate: "<instance> is in state <status>",
					InputPathsMap: map[string]string{
						"instance": "$.detail.instance",
						"status":   "$.detail.status",
					},
				},
				opts: Options{
					ebClient:  newMockEbClient(t),
					sqsClient: newMockSqsClient(t),
				},
			},
			mockCreateQueue: func(ctx context.Context, lr *Listener, queue *queue.Queue) *mockCreateQueueFunc {
				m := newMockCreateQueueFunc(t)
				m.EXPECT().
					Execute(ctx, lr.opts.sqsClient, lr.ID(), mock.AnythingOfType("map[string]string"), mock.AnythingOfType("queue.Options")).
					Return(nil, errors.New("create queue failed"))
				return m
			},
			mockCreateRule: func(ctx context.Context, lr *Listener, rule *eventrule.Rule) *mockCreateRuleFunc {
				m := newMockCreateRuleFunc(t)
				description := fmt.Sprintf("rule for Listener %q; created by zion", lr.ID())
				m.EXPECT().
					Execute(ctx, lr.opts.ebClient, lr.ID(), lr.eventBus.Name, lr.eventPattern, description, mock.AnythingOfType("map[string]string")).
					Return(rule, nil)
				return m
			},
			mockPutQueueTarget: func(ctx context.Context, lr *Listener, queue *queue.Queue, rule *eventrule.Rule) *mockPutQueueTargetFunc {
				m := newMockPutQueueTargetFunc(t)
				return m
			},
			expectQueue: &queue.Queue{},
			expectRule:  &eventrule.Rule{},
			expectErr: func(lr *Listener) error {
				return fmt.Errorf("failed to deploy eb listener %v: create queue failed", lr.ID())
			},
		},
		"should return error if createRule failed": {
			lr: &Listener{
				id:           xid.New().String(),
				eventBus:     &eventbus.EventBus{Name: "my-event-bus"},
				eventPattern: "{}",
				input:        "",
				inputPath:    "",
				inputTransformer: &eventrule.InputTransformer{
					InputTemplate: "<instance> is in state <status>",
					InputPathsMap: map[string]string{
						"instance": "$.detail.instance",
						"status":   "$.detail.status",
					},
				},
				opts: Options{
					ebClient:  newMockEbClient(t),
					sqsClient: newMockSqsClient(t),
				},
			},
			mockCreateQueue: func(ctx context.Context, lr *Listener, queue *queue.Queue) *mockCreateQueueFunc {
				m := newMockCreateQueueFunc(t)
				// m.EXPECT().
				// 	Execute(ctx, lr.opts.sqsClient, lr.ID(), mock.AnythingOfType("map[string]string"), mock.AnythingOfType("queue.Options")).
				// 	Return(queue, nil)
				return m
			},
			mockCreateRule: func(ctx context.Context, lr *Listener, rule *eventrule.Rule) *mockCreateRuleFunc {
				m := newMockCreateRuleFunc(t)
				description := fmt.Sprintf("rule for Listener %q; created by zion", lr.ID())
				m.EXPECT().
					Execute(ctx, lr.opts.ebClient, lr.ID(), lr.eventBus.Name, lr.eventPattern, description, mock.AnythingOfType("map[string]string")).
					Return(nil, errors.New("create rule failed"))
				return m
			},
			mockPutQueueTarget: func(ctx context.Context, lr *Listener, queue *queue.Queue, rule *eventrule.Rule) *mockPutQueueTargetFunc {
				m := newMockPutQueueTargetFunc(t)
				return m
			},
			expectQueue: &queue.Queue{},
			expectRule:  &eventrule.Rule{},
			expectErr: func(lr *Listener) error {
				return fmt.Errorf("failed to deploy eb listener %v: create rule failed", lr.ID())
			},
		},
		"should return error if putQueueAsRuleTarget failed": {
			lr: &Listener{
				id:           xid.New().String(),
				eventBus:     &eventbus.EventBus{Name: "my-event-bus"},
				eventPattern: "{}",
				input:        "",
				inputPath:    "",
				inputTransformer: &eventrule.InputTransformer{
					InputTemplate: "<instance> is in state <status>",
					InputPathsMap: map[string]string{
						"instance": "$.detail.instance",
						"status":   "$.detail.status",
					},
				},
				opts: Options{
					ebClient:  newMockEbClient(t),
					sqsClient: newMockSqsClient(t),
				},
			},
			mockCreateQueue: func(ctx context.Context, lr *Listener, queue *queue.Queue) *mockCreateQueueFunc {
				m := newMockCreateQueueFunc(t)
				m.EXPECT().
					Execute(ctx, lr.opts.sqsClient, lr.ID(), mock.AnythingOfType("map[string]string"), mock.AnythingOfType("queue.Options")).
					Return(queue, nil)
				return m
			},
			mockCreateRule: func(ctx context.Context, lr *Listener, rule *eventrule.Rule) *mockCreateRuleFunc {
				m := newMockCreateRuleFunc(t)
				description := fmt.Sprintf("rule for Listener %q; created by zion", lr.ID())
				m.EXPECT().
					Execute(ctx, lr.opts.ebClient, lr.ID(), lr.eventBus.Name, lr.eventPattern, description, mock.AnythingOfType("map[string]string")).
					Return(rule, nil)
				return m
			},
			mockPutQueueTarget: func(ctx context.Context, lr *Listener, queue *queue.Queue, rule *eventrule.Rule) *mockPutQueueTargetFunc {
				m := newMockPutQueueTargetFunc(t)
				m.EXPECT().
					Execute(ctx, lr.opts.ebClient, lr.ID(), queue, rule, lr.input, lr.inputPath, lr.inputTransformer).
					Return(errors.New("put rule target failed"))
				return m
			},
			expectQueue: &queue.Queue{},
			expectRule:  &eventrule.Rule{},
			expectErr: func(lr *Listener) error {
				return fmt.Errorf("failed to deploy eb listener %v: put rule target failed", lr.ID())
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			mockCreateQueue := tt.mockCreateQueue(ctx, tt.lr, tt.expectQueue)
			mockCreateRule := tt.mockCreateRule(ctx, tt.lr, tt.expectRule)
			mockPutQueueTarget := tt.mockPutQueueTarget(ctx, tt.lr, tt.expectQueue, tt.expectRule)
			tt.lr.opts.createQueue = mockCreateQueue.Execute
			tt.lr.opts.createRule = mockCreateRule.Execute
			tt.lr.opts.putQueueTarget = mockPutQueueTarget.Execute
			err := tt.lr.Deploy(ctx)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr(tt.lr).Error())
			} else {
				assert.Equal(t, tt.expectQueue, tt.lr.queue)
				assert.Equal(t, tt.expectRule, tt.lr.rule)
			}
		})
	}
}

func TestListener_Destroy(t *testing.T) {
	cases := map[string]struct {
		lr              *Listener
		mockDeleteQueue func(ctx context.Context, lr *Listener) *mockDeleteQueueFunc
		mockDeleteRule  func(ctx context.Context, lr *Listener) *mockDeleteRuleFunc
		expectErr       func(lr *Listener) error
	}{
		"should destroy successfully": {
			lr: &Listener{
				id:       xid.New().String(),
				eventBus: &eventbus.EventBus{Name: "my-event-bus"},
				rule: &eventrule.Rule{
					Name:         "rule",
					EventBusName: "eventbus",
				},
				queue: &queue.Queue{
					QueueURL: "queue-url",
				},
				opts: Options{
					ebClient:  newMockEbClient(t),
					sqsClient: newMockSqsClient(t),
				},
			},
			expectErr: nil,
			mockDeleteQueue: func(ctx context.Context, lr *Listener) *mockDeleteQueueFunc {
				mk := newMockDeleteQueueFunc(t)
				mk.EXPECT().Execute(ctx, lr.opts.sqsClient, lr.queue.QueueURL).Return(nil)
				return mk
			},
			mockDeleteRule: func(ctx context.Context, lr *Listener) *mockDeleteRuleFunc {
				mk := newMockDeleteRuleFunc(t)
				mk.EXPECT().Execute(ctx, lr.opts.ebClient, lr.rule.EventBusName, lr.rule.Name).Return(nil)
				return mk
			},
		},
		"should return error if both rule and queue are created and DeleteRule failed": {
			lr: &Listener{
				id:       xid.New().String(),
				eventBus: &eventbus.EventBus{Name: "my-event-bus"},
				rule: &eventrule.Rule{
					Name:         "rule",
					EventBusName: "eventbus",
				},
				queue: &queue.Queue{
					QueueURL: "queue-url",
				},
				opts: Options{
					ebClient:  newMockEbClient(t),
					sqsClient: newMockSqsClient(t),
				},
			},
			expectErr: func(lr *Listener) error {
				return fmt.Errorf("failed to destroy eb listener %v: failed to delete rule", lr.ID())
			},
			mockDeleteQueue: func(ctx context.Context, lr *Listener) *mockDeleteQueueFunc {
				mk := newMockDeleteQueueFunc(t)
				return mk
			},
			mockDeleteRule: func(ctx context.Context, lr *Listener) *mockDeleteRuleFunc {
				mk := newMockDeleteRuleFunc(t)
				mk.EXPECT().Execute(ctx, lr.opts.ebClient, lr.rule.EventBusName, lr.rule.Name).Return(errors.New("failed to delete rule"))
				return mk
			},
		},
		"should return error if both rule and queue are created and DeleteQueue failed": {
			lr: &Listener{
				id:       xid.New().String(),
				eventBus: &eventbus.EventBus{Name: "my-event-bus"},
				rule: &eventrule.Rule{
					Name:         "rule",
					EventBusName: "eventbus",
				},
				queue: &queue.Queue{
					QueueURL: "queue-url",
				},
				opts: Options{
					ebClient:  newMockEbClient(t),
					sqsClient: newMockSqsClient(t),
				},
			},
			expectErr: func(lr *Listener) error {
				return fmt.Errorf("failed to destroy eb listener %v: failed to delete queue", lr.ID())
			},
			mockDeleteQueue: func(ctx context.Context, lr *Listener) *mockDeleteQueueFunc {
				mk := newMockDeleteQueueFunc(t)
				mk.EXPECT().Execute(ctx, lr.opts.sqsClient, lr.queue.QueueURL).Return(errors.New("failed to delete queue"))
				return mk
			},
			mockDeleteRule: func(ctx context.Context, lr *Listener) *mockDeleteRuleFunc {
				mk := newMockDeleteRuleFunc(t)
				mk.EXPECT().Execute(ctx, lr.opts.ebClient, lr.rule.EventBusName, lr.rule.Name).Return(nil)
				return mk
			},
		},
		"should succeed if nothing to destroy": {
			lr: &Listener{
				id:       xid.New().String(),
				eventBus: &eventbus.EventBus{Name: "my-event-bus"},
				opts: Options{
					ebClient:  newMockEbClient(t),
					sqsClient: newMockSqsClient(t),
				},
			},
			expectErr: nil,
			mockDeleteQueue: func(ctx context.Context, lr *Listener) *mockDeleteQueueFunc {
				mk := newMockDeleteQueueFunc(t)
				return mk
			},
			mockDeleteRule: func(ctx context.Context, lr *Listener) *mockDeleteRuleFunc {
				mk := newMockDeleteRuleFunc(t)
				return mk
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			mockDeleteQueue := tt.mockDeleteQueue(ctx, tt.lr)
			mockDeleteRule := tt.mockDeleteRule(ctx, tt.lr)
			tt.lr.opts.deleteQueue = mockDeleteQueue.Execute
			tt.lr.opts.deleteRule = mockDeleteRule.Execute
			err := tt.lr.Destroy(ctx)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr(tt.lr).Error())
			} else {
				assert.Nil(t, tt.lr.rule)
				assert.Nil(t, tt.lr.queue)
			}
		})
	}
}

func TestListener_JSON(t *testing.T) {
	listener := &Listener{
		id:       xid.New().String(),
		eventBus: &eventbus.EventBus{Name: testBusName, ARN: testBusARN()},
		rule: &eventrule.Rule{
			Name:         testListenerID,
			EventBusName: testBusName,
			ARN:          testRuleARN(),
		},
		queue: &queue.Queue{
			Name: testListenerID,
			ARN:  testQueueARN(),
		},
	}
	actual := listener.JSON()
	expect := Output{
		ID:         listener.ID(),
		TestTarget: listener.eventBus.Resource(),
		Components: []harness.Resource{
			listener.queue.Resource(),
			listener.rule.Resource(),
		},
	}
	assert.Equal(t, expect, actual)
}

func TestListner_ReceiveEvents(t *testing.T) {
	cases := map[string]struct {
		listener            *Listener
		waitTimeSeconds     int32
		maxNumberOfMessages int32
		mock                func(ctx context.Context, lr *Listener, waitTimeSeconds, maxNumberOfMessages int32) *mockSqsClient
		expect              []Event
		expectErr           error
	}{
		"should succeed": {
			listener: &Listener{
				id: xid.New().String(),
				queue: &queue.Queue{
					QueueURL: "queue-url",
				},
			},
			waitTimeSeconds:     10,
			maxNumberOfMessages: 10,
			mock: func(ctx context.Context, lr *Listener, waitTimeSeconds, maxNumberOfMessages int32) *mockSqsClient {
				client := newMockSqsClient(t)
				client.EXPECT().
					ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
						QueueUrl:            aws.String(lr.queue.QueueURL),
						MaxNumberOfMessages: maxNumberOfMessages,
						WaitTimeSeconds:     waitTimeSeconds,
						VisibilityTimeout:   waitTimeSeconds + 5,
					}).
					Return(&sqs.ReceiveMessageOutput{
						Messages: []sqstypes.Message{
							{Body: aws.String("{}"), ReceiptHandle: aws.String("123")},
							{Body: aws.String("{}"), ReceiptHandle: aws.String("456")},
							{Body: aws.String("{}"), ReceiptHandle: aws.String("789")},
						},
					}, nil)
				return client
			},
			expect: []Event{
				{"{}", "123"},
				{"{}", "456"},
				{"{}", "789"},
			},
			expectErr: nil,
		},
		"should return error due to api failure": {
			listener: &Listener{
				id: xid.New().String(),
				queue: &queue.Queue{
					QueueURL: "queue-url",
				},
			},
			waitTimeSeconds:     10,
			maxNumberOfMessages: 10,
			mock: func(ctx context.Context, lr *Listener, waitTimeSeconds, maxNumberOfMessages int32) *mockSqsClient {
				client := newMockSqsClient(t)
				client.EXPECT().
					ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
						QueueUrl:            aws.String(lr.queue.QueueURL),
						MaxNumberOfMessages: maxNumberOfMessages,
						WaitTimeSeconds:     waitTimeSeconds,
						VisibilityTimeout:   waitTimeSeconds + 5,
					}).
					Return(nil, errors.New("api failure"))
				return client
			},
			expect:    nil,
			expectErr: errors.New("failed to receive events: api failure"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			tt.listener.opts.sqsClient = tt.mock(ctx, tt.listener, tt.waitTimeSeconds, tt.maxNumberOfMessages)
			actual, err := tt.listener.ReceiveEvents(ctx, tt.waitTimeSeconds, tt.maxNumberOfMessages)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, tt.expect, actual)
			}
		})
	}
}

func TestListener_DeleteEvents(t *testing.T) {
	cases := map[string]struct {
		listener       *Listener
		receiptHandles []string
		mock           func(ctx context.Context, lr *Listener) *mockSqsClient
		expectErr      error
	}{
		"should succeed": {
			listener: &Listener{
				id: xid.New().String(),
				queue: &queue.Queue{
					QueueURL: "queue-url",
				},
			},
			receiptHandles: []string{"123", "456", "789"},
			mock: func(ctx context.Context, lr *Listener) *mockSqsClient {
				client := newMockSqsClient(t)
				client.EXPECT().
					DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
						QueueUrl: aws.String(lr.queue.QueueURL),
						Entries: []sqstypes.DeleteMessageBatchRequestEntry{
							{Id: aws.String("0"), ReceiptHandle: aws.String("123")},
							{Id: aws.String("1"), ReceiptHandle: aws.String("456")},
							{Id: aws.String("2"), ReceiptHandle: aws.String("789")},
						},
					}).
					Return(&sqs.DeleteMessageBatchOutput{}, nil)
				return client
			},
			expectErr: nil,
		},
		"should return error due to api failure": {
			listener: &Listener{
				id: xid.New().String(),
				queue: &queue.Queue{
					QueueURL: "queue-url",
				},
			},
			receiptHandles: []string{"123", "456", "789"},
			mock: func(ctx context.Context, lr *Listener) *mockSqsClient {
				client := newMockSqsClient(t)
				client.EXPECT().
					DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
						QueueUrl: aws.String(lr.queue.QueueURL),
						Entries: []sqstypes.DeleteMessageBatchRequestEntry{
							{Id: aws.String("0"), ReceiptHandle: aws.String("123")},
							{Id: aws.String("1"), ReceiptHandle: aws.String("456")},
							{Id: aws.String("2"), ReceiptHandle: aws.String("789")},
						},
					}).
					Return(nil, errors.New("api failure"))
				return client
			},
			expectErr: errors.New("failed to delete events: api failure"),
		},
		"should return error due to empty receipt handles": {
			listener: &Listener{
				id: xid.New().String(),
				queue: &queue.Queue{
					QueueURL: "queue-url",
				},
			},
			receiptHandles: []string{},
			mock: func(ctx context.Context, lr *Listener) *mockSqsClient {
				client := newMockSqsClient(t)
				return client
			},
			expectErr: errors.New("receiptHandles must have at least one item"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			tt.listener.opts.sqsClient = tt.mock(ctx, tt.listener)
			err := tt.listener.DeleteEvents(ctx, tt.receiptHandles)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			}
		})
	}
}

func Test_queuePolicy(t *testing.T) {
	qp := queuePolicy{
		queueARN: testQueueARN(),
		ruleARN:  testRuleARN(),
	}
	actual := qp.String()
	expect := `{"Version": "2012-10-17", "Id": "Write_Permission_for_Rule_rule/my-event-bus/zion_eb_9m4e2mr0ui3e8a215n4g", "Statement": [{"Sid": "eblistener", "Effect": "Allow", "Principal": {"Service": "events.amazonaws.com"}, "Action": "sqs:SendMessage", "Resource": "arn:aws:sqs:us-west-2:123456789012:zion_eb_9m4e2mr0ui3e8a215n4g", "Condition": {"ArnEquals": {"aws:SourceArn": "arn:aws:sqs:us-west-2:123456789012:rule/my-event-bus/zion_eb_9m4e2mr0ui3e8a215n4g"}}}]}`
	assert.Equal(t, expect, actual)
	var deserialized any
	err := json.Unmarshal([]byte(actual), &deserialized)
	assert.Nil(t, err)
}
