// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package eventrule

import (
	"context"
	"fmt"
	"log"
	"zion/internal/pkg/harness"
	"zion/internal/pkg/harness/resource/queue"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

type InputTransformer struct {
	InputTemplate string            `json:"InputTemplate"`
	InputPathsMap map[string]string `json:"InputPathsMap"`
}

type Rule struct {
	Name         string
	EventBusName string
	EventPattern string
	ARN          arn.ARN
}

func (r *Rule) Resource() harness.Resource {
	return harness.Resource{
		Type:       "AWS::Events::Rule",
		PhysicalID: r.Name,
		ARN:        r.ARN.String(),
	}
}

func ListTargetsByRule(ctx context.Context, api EbListTargetsByRuleAPI, targetId, ruleName string, eventBusName string) (*ebtypes.Target, error) {
	// If no rule provided, skip
	if ruleName == "" || targetId == "" {
		return nil, nil
	}

	input := eventbridge.ListTargetsByRuleInput{
		Rule:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	}

	for {
		output, err := api.ListTargetsByRule(ctx, &input)

		if err != nil {
			return nil, fmt.Errorf("ListTargetsByRule for Rule: %q failed: %v", ruleName, err)
		}

		targetLen := len(output.Targets)

		if targetLen == 0 {
			return nil, fmt.Errorf("TargetId was provided but not found on Rule: %s", ruleName)
		}

		for _, t := range output.Targets {
			if aws.ToString(t.Id) == targetId {
				return &t, nil
			}
		}
		if aws.ToString(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil, fmt.Errorf("TagetId: %s was not found on %s Rule", targetId, ruleName)
}

func Create(ctx context.Context, api EbPutRuleAPI, ruleName, eventBusName, eventPattern, description string, tags map[string]string) (*Rule, error) {

	log.Printf("creating event rule %q", ruleName)
	output, err := api.PutRule(ctx, &eventbridge.PutRuleInput{
		Name:         aws.String(ruleName),
		Description:  aws.String(description),
		EventBusName: aws.String(eventBusName),
		EventPattern: aws.String(eventPattern),
		Tags:         tagsToEBTags(tags),
	})

	if err != nil {
		return nil, fmt.Errorf("put rule %q failed: %v", ruleName, err)
	}

	arn, _ := arn.Parse(aws.ToString(output.RuleArn))
	log.Printf("created event rule %q", arn)
	return &Rule{
		Name:         ruleName,
		EventBusName: eventBusName,
		EventPattern: eventPattern,
		ARN:          arn,
	}, nil
}

func PutQueueTarget(ctx context.Context, api EbPutTargetsAPI, listenerID string, qu *queue.Queue, ru *Rule, target *ebtypes.Target) error {
	targetOverride := ebtypes.Target{
		Arn:              aws.String(qu.ARN.String()),
		Id:               aws.String(listenerID),
		Input:            target.Input,
		InputPath:        target.InputPath,
		InputTransformer: target.InputTransformer,
	}

	log.Printf("put sqs queue %q as target", qu.QueueURL)
	_, err := api.PutTargets(ctx, &eventbridge.PutTargetsInput{
		Rule:         aws.String(ru.Name),
		EventBusName: aws.String(ru.EventBusName),
		Targets:      []ebtypes.Target{targetOverride},
	})

	if err != nil {
		return fmt.Errorf("put rule target failed: %w", err)
	}

	log.Printf("put rule target complete")
	return nil
}

func Delete(ctx context.Context, api EbDeleteRuleAPI, eventBusName, ruleName string) error {
	targets, err := api.ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
		Rule:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete rule %q: %v", ruleName, err)
	}

	var ids []string
	for _, target := range targets.Targets {
		ids = append(ids, aws.ToString(target.Id))
	}
	if len(ids) > 0 {
		_, err = api.RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
			Ids:          ids,
			Rule:         aws.String(ruleName),
			EventBusName: aws.String(eventBusName),
		})
		if err != nil {
			return fmt.Errorf("failed to delete rule %q: %v", ruleName, err)
		}
	}

	log.Printf("deleting rule %q of event bus %q", ruleName, eventBusName)
	_, err = api.DeleteRule(ctx, &eventbridge.DeleteRuleInput{
		Name:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	})

	if err != nil {
		return fmt.Errorf("failed to delete rule %q: %v", ruleName, err)
	}

	log.Printf("deleted rule %q", ruleName)
	return nil
}

func Get(ctx context.Context, api EbDescribeRuleAPI, ruleName, eventBusName string) (*Rule, error) {
	output, err := api.DescribeRule(ctx, &eventbridge.DescribeRuleInput{
		Name:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe rule %q of event bus %q: %v", ruleName, eventBusName, err)
	}
	arn, _ := arn.Parse(aws.ToString(output.Arn))
	return &Rule{
		Name:         ruleName,
		EventBusName: eventBusName,
		EventPattern: aws.ToString(output.EventPattern),
		ARN:          arn,
	}, nil
}

func tagsToEBTags(tags map[string]string) []ebtypes.Tag {
	var out []ebtypes.Tag
	for key, val := range tags {
		out = append(out, ebtypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(val),
		})
	}
	return out
}

//go:generate mockery --name EbPutRuleAPI
type EbPutRuleAPI interface {
	PutRule(ctx context.Context, params *eventbridge.PutRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutRuleOutput, error)
}

//go:generate mockery --name EbListTargetsByRuleAPI
type EbListTargetsByRuleAPI interface {
	ListTargetsByRule(ctx context.Context, params *eventbridge.ListTargetsByRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTargetsByRuleOutput, error)
}

//go:generate mockery --name EbDescribeRuleAPI
type EbDescribeRuleAPI interface {
	DescribeRule(ctx context.Context, params *eventbridge.DescribeRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DescribeRuleOutput, error)
}

//go:generate mockery --name EbDeleteRuleAPI
type EbDeleteRuleAPI interface {
	DeleteRule(ctx context.Context, params *eventbridge.DeleteRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.DeleteRuleOutput, error)
	ListTargetsByRule(ctx context.Context, params *eventbridge.ListTargetsByRuleInput, optFns ...func(*eventbridge.Options)) (*eventbridge.ListTargetsByRuleOutput, error)
	RemoveTargets(ctx context.Context, params *eventbridge.RemoveTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.RemoveTargetsOutput, error)
}

//go:generate mockery --name EbPutTargetsAPI
type EbPutTargetsAPI interface {
	PutTargets(ctx context.Context, params *eventbridge.PutTargetsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutTargetsOutput, error)
}
