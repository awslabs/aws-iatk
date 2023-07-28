// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package listener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	"zion/internal/pkg/harness"
	"zion/internal/pkg/harness/resource/eventbus"
	"zion/internal/pkg/harness/resource/eventrule"
	"zion/internal/pkg/harness/resource/queue"
	"zion/internal/pkg/harness/tags"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Output struct {
	ID         string             `json:"Id"`
	TestTarget harness.Resource   `json:"TargetUnderTest"`
	Components []harness.Resource `json:"Components"`
}

// Options for configuring dependency clients/funcs for Listener
type Options struct {
	EventBusName string
	// aws clients
	ebClient  ebClient
	sqsClient sqsClient

	// funcs
	getEventBus              getEventBusFunc
	listTargetsByRule        listTargetsByRuleFunc
	createQueue              createQueueFunc
	createRule               createRuleFunc
	putQueueTarget           putQueueTargetFunc
	deleteQueue              deleteQueueFunc
	deleteRule               deleteRuleFunc
	getQueueWithName         getQueueWithNameFunc
	getRule                  getRuleFunc
	getEventBusNameFromQueue getEventBusNameFromQueueFunc
}

func NewOptions(cfg aws.Config) Options {
	return Options{
		ebClient:  eventbridge.NewFromConfig(cfg),
		sqsClient: sqs.NewFromConfig(cfg),

		getEventBus:              eventbus.Get,
		listTargetsByRule:        eventrule.ListTargetsByRule,
		createQueue:              queue.Create,
		createRule:               eventrule.Create,
		putQueueTarget:           eventrule.PutQueueTarget,
		deleteQueue:              queue.Delete,
		deleteRule:               eventrule.Delete,
		getQueueWithName:         queue.GetWithName,
		getRule:                  eventrule.Get,
		getEventBusNameFromQueue: queue.GetEventBusNameFromQueue,
	}
}

// Listener struct
type Listener struct {
	id           string
	eventPattern string
	customTags   map[string]string

	target *ebtypes.Target

	// target
	eventBus *eventbus.EventBus
	// testing resources
	queue *queue.Queue
	rule  *eventrule.Rule

	opts Options
}

func (lr *Listener) ID() string {
	return IDPrefix + lr.id
}

func (lr *Listener) tags(ts time.Time) map[string]string {
	tags := map[string]string{
		string(tags.TestHarnessID):      lr.ID(),
		string(tags.TestHarnessType):    TestHarnessType,
		string(tags.TestHarnessTarget):  lr.eventBus.ARN.String(),
		string(tags.TestHarnessCreated): ts.Format(time.RFC3339),
	}
	for key, val := range lr.customTags {
		tags[key] = val
	}
	return tags
}

func (lr *Listener) String() string {
	return fmt.Sprintf("eb listener id: %v", lr.ID())
}

func (lr *Listener) Components() []harness.Resource {
	r := []harness.Resource{}
	if lr.queue != nil {
		r = append(r, lr.queue.Resource())
	}
	if lr.rule != nil {
		r = append(r, lr.rule.Resource())
	}
	return r
}

func (lr *Listener) Deploy(ctx context.Context) error {
	log.Printf("start deploy eb listener %v", lr.ID())
	ts := time.Now()
	tags := lr.tags(ts)
	partition := lr.eventBus.ARN.Partition
	region := lr.eventBus.ARN.Region
	accountID := lr.eventBus.ARN.AccountID

	rn := ruleName(lr.ID())
	description := fmt.Sprintf("rule for Listener %q; created by zion", lr.ID())
	r, err := lr.opts.createRule(ctx, lr.opts.ebClient, rn, lr.eventBus.Name, lr.eventPattern, description, tags)
	if err != nil {
		return fmt.Errorf("failed to deploy eb listener %v: %w", lr.ID(), err)
	}
	lr.rule = r

	qn := queueName(lr.ID())
	qpolicy := queuePolicy{
		// NOTE: generate queue ARN before the queue is created
		queueARN: arn.ARN{
			Partition: partition,
			Service:   "sqs",
			Region:    region,
			AccountID: accountID,
			Resource:  lr.ID(),
		},
		ruleARN: lr.rule.ARN,
	}

	q, err := lr.opts.createQueue(ctx, lr.opts.sqsClient, qn, tags, queue.Options{
		Policy:                 qpolicy.String(),
		MessageRetentionPeriod: 3600, // 1 hour
	})
	if err != nil {
		return fmt.Errorf("failed to deploy eb listener %v: %w", lr.ID(), err)
	}
	lr.queue = q

	err = lr.opts.putQueueTarget(ctx, lr.opts.ebClient, lr.ID(), lr.queue, lr.rule, lr.target)
	if err != nil {
		return fmt.Errorf("failed to deploy eb listener %v: %w", lr.ID(), err)
	}

	log.Printf("complete deploy eb listener %v", lr.ID())
	return nil
}

func (lr *Listener) Destroy(ctx context.Context) error {
	if lr.rule == nil && lr.queue == nil {
		log.Printf("nothing to destroy for eb listener %v", lr.ID())
		return nil
	}

	log.Printf("destroy start (%v)", lr.String())
	if lr.rule != nil {
		if err := lr.opts.deleteRule(ctx, lr.opts.ebClient, lr.rule.EventBusName, lr.rule.Name); err != nil {
			return fmt.Errorf("failed to destroy eb listener %v: %w", lr.ID(), err)
		}
	}
	lr.rule = nil

	if lr.queue != nil {
		if err := lr.opts.deleteQueue(ctx, lr.opts.sqsClient, lr.queue.QueueURL); err != nil {
			return fmt.Errorf("failed to destroy eb listener %v: %w", lr.ID(), err)
		}
	}
	lr.queue = nil

	log.Printf("complete destroy eb listener %v", lr.ID())
	return nil
}

func (lr *Listener) JSON() Output {
	return Output{
		ID:         lr.ID(),
		TestTarget: lr.eventBus.Resource(),
		Components: lr.Components(),
	}
}

func (lr *Listener) ReceiveEvents(ctx context.Context, waitTimeSeconds, maxNumberOfMessages int32) ([]Event, error) {
	messages, err := lr.opts.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(lr.queue.QueueURL),
		MaxNumberOfMessages: maxNumberOfMessages,
		WaitTimeSeconds:     waitTimeSeconds,
		VisibilityTimeout:   waitTimeSeconds + 5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to receive events: %w", err)
	}
	events := make([]Event, 0, maxNumberOfMessages)
	for _, m := range messages.Messages {
		e := Event{aws.ToString(m.Body), aws.ToString(m.ReceiptHandle)}
		events = append(events, e)
	}
	return events, nil
}

func (lr *Listener) DeleteEvents(ctx context.Context, receiptHandles []string) error {
	if len(receiptHandles) == 0 {
		return errors.New("receiptHandles must have at least one item")
	}
	entries := make([]sqstypes.DeleteMessageBatchRequestEntry, 0, len(receiptHandles))
	for i, h := range receiptHandles {
		entries = append(entries, sqstypes.DeleteMessageBatchRequestEntry{
			Id:            aws.String(strconv.Itoa(i)),
			ReceiptHandle: aws.String(h),
		})
	}
	_, err := lr.opts.sqsClient.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(lr.queue.QueueURL),
		Entries:  entries,
	})
	if err != nil {
		return fmt.Errorf("failed to delete events: %w", err)
	}
	return nil
}

type Event struct {
	Body          string `json:"Body"`
	ReceiptHandle string `json:"ReceiptHandle"`
}
type queuePolicy struct {
	queueARN arn.ARN
	ruleARN  arn.ARN
}

func (qp *queuePolicy) statementForEvents() string {
	return fmt.Sprintf(
		`{"Sid": "eblistener", "Effect": "Allow", "Principal": {"Service": "events.amazonaws.com"}, "Action": "sqs:SendMessage", "Resource": %q, "Condition": {"ArnEquals": {"aws:SourceArn": %q}}}`,
		qp.queueARN.String(),
		qp.ruleARN.String(),
	)
}

func (qp *queuePolicy) String() string {
	return fmt.Sprintf(
		`{"Version": "2012-10-17", "Id": "Write_Permission_for_Rule_%v", "Statement": [%v]}`,
		qp.ruleARN.Resource,
		qp.statementForEvents(),
	)
}

//go:generate mockery --name ebClient
type ebClient interface {
	eventbus.DescribeEventBusAPI
	eventrule.EbListTargetsByRuleAPI
	eventrule.EbPutRuleAPI
	eventrule.EbDeleteRuleAPI
	eventrule.EbDescribeRuleAPI
	eventrule.EbPutTargetsAPI
}

//go:generate mockery --name sqsClient
type sqsClient interface {
	queue.CreateQueueAPI
	queue.DeleteQueueAPI
	queue.GetQueueUrlAPI
	queue.ListQueueTagsAPI
	queue.ReceiveMessageAPI
	queue.DeleteMessageBatchAPI
}

//go:generate mockery --name getEventBusFunc
type getEventBusFunc func(ctx context.Context, api eventbus.DescribeEventBusAPI, eventBusName string) (*eventbus.EventBus, error)

//go:generate mockery --name listTargetsByRuleFunc
type listTargetsByRuleFunc func(ctx context.Context, api eventrule.EbListTargetsByRuleAPI, targetId, ruleName string, eventBusName string) (*ebtypes.Target, error)

//go:generate mockery --name createQueueFunc
type createQueueFunc func(ctx context.Context, api queue.CreateQueueAPI, name string, tags map[string]string, opts queue.Options) (*queue.Queue, error)

//go:generate mockery --name createRuleFunc
type createRuleFunc func(ctx context.Context, api eventrule.EbPutRuleAPI, ruleName, eventBusName, eventPattern, description string, tags map[string]string) (*eventrule.Rule, error)

//go:generate mockery --name putQueueTargetFunc
type putQueueTargetFunc func(ctx context.Context, api eventrule.EbPutTargetsAPI, listenerID string, qu *queue.Queue, ru *eventrule.Rule, target *ebtypes.Target) error

//go:generate mockery --name deleteQueueFunc
type deleteQueueFunc func(ctx context.Context, api queue.DeleteQueueAPI, queueURL string) error

//go:generate mockery --name deleteRuleFunc
type deleteRuleFunc func(ctx context.Context, api eventrule.EbDeleteRuleAPI, eventBusName, ruleName string) error

//go:generate mockery --name getQueueWithNameFunc
type getQueueWithNameFunc func(ctx context.Context, api queue.GetQueueUrlAPI, name string) (*queue.Queue, error)

//go:generate mockery --name getRuleFunc
type getRuleFunc func(ctx context.Context, api eventrule.EbDescribeRuleAPI, ruleName, eventBusName string) (*eventrule.Rule, error)

//go:generate mockery --name getEventBusNameFromQueueFunc
type getEventBusNameFromQueueFunc func(ctx context.Context, api queue.ListQueueTagsAPI, queueURL string) (string, error)
