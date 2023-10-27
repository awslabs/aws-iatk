// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package queue

import (
	"context"
	"ctk/internal/pkg/harness"
	"ctk/internal/pkg/harness/tags"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Queue struct {
	Name     string
	QueueURL string
	ARN      arn.ARN
}

func (q *Queue) Resource() harness.Resource {
	return harness.Resource{
		Type:       "AWS::SQS::Queue",
		PhysicalID: q.QueueURL,
		ARN:        q.ARN.String(),
	}
}

type Options struct {
	Policy                 string
	MessageRetentionPeriod int64
	VisibilityTimeout      int64
}

func Create(ctx context.Context, api CreateQueueAPI, name string, tags map[string]string, opts Options) (*Queue, error) {
	log.Printf("start create queue %q", name)
	output, err := api.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(name),
		Tags:      tags,
		Attributes: map[string]string{
			"MessageRetentionPeriod": fmt.Sprint(opts.MessageRetentionPeriod),
			"Policy":                 opts.Policy,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("create queue %q failed: %v", name, err)
	}

	attrs, err := api.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: output.QueueUrl,
		AttributeNames: []sqstypes.QueueAttributeName{
			sqstypes.QueueAttributeNameQueueArn,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get queue attributes failed: %v", err)
	}
	queueARN, _ := arn.Parse(attrs.Attributes["QueueArn"])

	queueURL := aws.ToString(output.QueueUrl)
	log.Printf("created queue %q", queueURL)
	return &Queue{
		Name:     name,
		QueueURL: queueURL,
		ARN:      queueARN,
	}, nil
}

//go:generate mockery --name CreateQueueAPI
type CreateQueueAPI interface {
	CreateQueue(ctx context.Context, params *sqs.CreateQueueInput, optFns ...func(*sqs.Options)) (*sqs.CreateQueueOutput, error)
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

//go:generate mockery --name DeleteQueueAPI
type DeleteQueueAPI interface {
	DeleteQueue(ctx context.Context, params *sqs.DeleteQueueInput, optFns ...func(*sqs.Options)) (*sqs.DeleteQueueOutput, error)
}

func Delete(ctx context.Context, api DeleteQueueAPI, queueURL string) error {
	log.Printf("deleting queue %q", queueURL)
	_, err := api.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueURL),
	})

	if err != nil {
		return fmt.Errorf("failed to delete queue %q: %v", queueURL, err)
	}

	log.Printf("deleted queue %q", queueURL)
	return nil
}

//go:generate mockery --name GetQueueUrlAPI
type GetQueueUrlAPI interface {
	GetQueueUrl(ctx context.Context, params *sqs.GetQueueUrlInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueUrlOutput, error)
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

func GetWithName(ctx context.Context, api GetQueueUrlAPI, name string) (*Queue, error) {
	output, err := api.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue with name %q: %v", name, err)
	}

	attrs, err := api.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: output.QueueUrl,
		AttributeNames: []sqstypes.QueueAttributeName{
			sqstypes.QueueAttributeNameQueueArn,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get queue attributes: %v", err)
	}
	queueARN, _ := arn.Parse(attrs.Attributes["QueueArn"])

	return &Queue{
		Name:     name,
		QueueURL: aws.ToString(output.QueueUrl),
		ARN:      queueARN,
	}, nil
}

//go:generate mockery --name ListQueueTagsAPI
type ListQueueTagsAPI interface {
	ListQueueTags(ctx context.Context, params *sqs.ListQueueTagsInput, optFns ...func(*sqs.Options)) (*sqs.ListQueueTagsOutput, error)
}

func GetEventBusNameFromQueue(ctx context.Context, api ListQueueTagsAPI, queueURL string) (string, error) {
	output, err := api.ListQueueTags(ctx, &sqs.ListQueueTagsInput{
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list queue tags: %v", err)
	}
	if val, ok := output.Tags[string(tags.TestHarnessTarget)]; ok {
		return val, nil
	}
	return "", fmt.Errorf("cannot get event bus name from queue")
}

//go:generate mockery --name ReceiveMessageAPI
type ReceiveMessageAPI interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
}

//go:generate mockery --name DeleteMessageBatchAPI
type DeleteMessageBatchAPI interface {
	DeleteMessageBatch(ctx context.Context, params *sqs.DeleteMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageBatchOutput, error)
}
