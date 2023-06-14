package eventrule

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"zion/internal/pkg/harness"
	"zion/internal/pkg/harness/resource/queue"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	testRuleName     string = "my-rule"
	testEventBusName string = "my-event-bus"
	testQueueName    string = "my-queue"
	testQueueURL     string = "my-queue-url"
	testEventPattern string = `{"source":[{"prefix":""}]}`
	testPartition    string = "aws"
	testService      string = "sqs"
	testRegion       string = "us-west-2"
	testAccountID    string = "123456789012"
)

func TestCreate(t *testing.T) {
	cases := map[string]struct {
		ruleName     string
		eventBusName string
		eventPattern string
		tags         map[string]string
		mockAPI      func(ctx context.Context, rule *Rule) *MockEbPutRuleAPI
		expect       *Rule
		expectErr    error
	}{
		"should put rule": {
			ruleName:     testRuleName,
			eventBusName: testEventBusName,
			eventPattern: testEventPattern,
			tags:         map[string]string{"key": "value"},
			expect: &Rule{
				Name:         testRuleName,
				EventBusName: testEventBusName,
				EventPattern: testEventPattern,
				ARN: arn.ARN{
					Partition: testPartition,
					Service:   testService,
					Region:    testRegion,
					AccountID: testAccountID,
					Resource:  "rule/" + testEventBusName + "/" + testRuleName,
				},
			},
			expectErr: nil,
			mockAPI: func(ctx context.Context, rule *Rule) *MockEbPutRuleAPI {
				api := NewMockEbPutRuleAPI(t)
				api.EXPECT().PutRule(ctx, &eventbridge.PutRuleInput{
					Name:         aws.String(rule.Name),
					Description:  aws.String(""),
					EventBusName: aws.String(rule.EventBusName),
					EventPattern: aws.String(rule.EventPattern),
					Tags:         tagsToEBTags(map[string]string{"key": "value"}),
				}).Return(&eventbridge.PutRuleOutput{
					RuleArn: aws.String(rule.ARN.String()),
				}, nil)
				return api
			},
		},
		"should return error if PutRule failed": {
			ruleName:     testRuleName,
			eventBusName: testEventBusName,
			tags:         map[string]string{"key": "value"},
			expect:       nil,
			expectErr:    fmt.Errorf(`put rule "%v" failed: something failed`, testRuleName),
			mockAPI: func(ctx context.Context, rule *Rule) *MockEbPutRuleAPI {
				api := NewMockEbPutRuleAPI(t)
				api.On("PutRule", ctx, mock.AnythingOfType("*eventbridge.PutRuleInput")).Return(nil, errors.New("something failed"))
				return api
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			api := tt.mockAPI(ctx, tt.expect)
			rule, err := Create(ctx, api, tt.ruleName, tt.eventBusName, tt.eventPattern, "", tt.tags)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, tt.expect, rule)
			}
			api.AssertExpectations(t)
		})
	}
}

func TestPutQueueTarget(t *testing.T) {
	cases := map[string]struct {
		resourceGroupID  string
		queue            *queue.Queue
		rule             *Rule
		input            string
		inputPath        string
		inputTransformer *InputTransformer
		expectErr        error
		mockAPI          func(ctx context.Context, resourceGroupID string, rule *Rule, queue *queue.Queue, input, inputPath string, inputTransformer *InputTransformer) *MockEbPutTargetsAPI
	}{
		"should put target": {
			resourceGroupID: "my-resource-group-id",
			queue: &queue.Queue{
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
			rule: &Rule{
				Name:         testRuleName,
				EventBusName: testEventBusName,
			},
			input:     "",
			inputPath: "",
			inputTransformer: &InputTransformer{
				InputTemplate: "<instance> is in state <status>",
				InputPathsMap: map[string]string{
					"instance": "$.detail.instance",
					"status":   "$.detail.status",
				},
			},
			expectErr: nil,
			mockAPI: func(ctx context.Context, resourceGroupID string, rule *Rule, queue *queue.Queue, input, inputPath string, inputTransformer *InputTransformer) *MockEbPutTargetsAPI {
				api := NewMockEbPutTargetsAPI(t)
				api.On("PutTargets", ctx, &eventbridge.PutTargetsInput{
					Rule:         aws.String(rule.Name),
					EventBusName: aws.String(rule.EventBusName),
					Targets: []ebtypes.Target{
						{
							Arn: aws.String(queue.ARN.String()),
							Id:  aws.String(resourceGroupID),
							InputTransformer: &ebtypes.InputTransformer{
								InputTemplate: aws.String(inputTransformer.InputTemplate),
								InputPathsMap: inputTransformer.InputPathsMap,
							},
						},
					},
				}).Return(nil, nil)
				return api
			},
		},
		"should return error if PutTargets failed": {
			resourceGroupID: "my-resource-group-id",
			queue: &queue.Queue{
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
			rule: &Rule{
				Name:         testRuleName,
				EventBusName: testEventBusName,
			},
			input:     "",
			inputPath: "",
			inputTransformer: &InputTransformer{
				InputTemplate: "<instance> is in state <status>",
				InputPathsMap: map[string]string{
					"instance": "$.detail.instance",
					"status":   "$.detail.status",
				},
			},
			expectErr: errors.New("put rule target failed: something failed from aws"),
			mockAPI: func(ctx context.Context, resourceGroupID string, rule *Rule, queue *queue.Queue, input, inputPath string, inputTransformer *InputTransformer) *MockEbPutTargetsAPI {
				api := NewMockEbPutTargetsAPI(t)
				api.On("PutTargets", ctx, &eventbridge.PutTargetsInput{
					Rule:         aws.String(rule.Name),
					EventBusName: aws.String(rule.EventBusName),
					Targets: []ebtypes.Target{
						{
							Arn: aws.String(queue.ARN.String()),
							Id:  aws.String(resourceGroupID),
							InputTransformer: &ebtypes.InputTransformer{
								InputTemplate: aws.String(inputTransformer.InputTemplate),
								InputPathsMap: inputTransformer.InputPathsMap,
							},
						},
					},
				}).Return(nil, errors.New("something failed from aws"))
				return api
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			api := tt.mockAPI(ctx, tt.resourceGroupID, tt.rule, tt.queue, tt.input, tt.inputPath, tt.inputTransformer)
			err := PutQueueTarget(ctx, api, tt.resourceGroupID, tt.queue, tt.rule, tt.input, tt.inputPath, tt.inputTransformer)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			}
			api.AssertExpectations(t)
		})
	}
}

func Test_setInputTraansformation(t *testing.T) {
	cases := map[string]struct {
		input            string
		inputPath        string
		inputTransformer *InputTransformer
		expectErr        error
	}{
		"should set target with input": {
			input:            "fixed",
			inputPath:        "",
			inputTransformer: nil,
			expectErr:        nil,
		},
		"should set target with inputPath": {
			input:            "",
			inputPath:        "$detail-type",
			inputTransformer: nil,
			expectErr:        nil,
		},
		"should set target with inputTransformer": {
			input:     "",
			inputPath: "",
			inputTransformer: &InputTransformer{
				InputTemplate: "<state>",
				InputPathsMap: map[string]string{
					"state": "$.detail.state",
				},
			},
			expectErr: nil,
		},
		"should fail with both input and inputPath provided": {
			input:            "fixed",
			inputPath:        "$.source",
			inputTransformer: nil,
			expectErr:        errors.New("input, inputPath, and inputTransformer are mutually exclusive"),
		},
		"should fail with both input and inputTransformer provided": {
			input:     "fixed",
			inputPath: "",
			inputTransformer: &InputTransformer{
				InputTemplate: "<state>",
				InputPathsMap: map[string]string{
					"state": "$.detail.state",
				},
			},
			expectErr: errors.New("input, inputPath, and inputTransformer are mutually exclusive"),
		},
		"should fail with both inputPath and inputTransformer provided": {
			input:     "",
			inputPath: "$.source",
			inputTransformer: &InputTransformer{
				InputTemplate: "<state>",
				InputPathsMap: map[string]string{
					"state": "$.detail.state",
				},
			},
			expectErr: errors.New("input, inputPath, and inputTransformer are mutually exclusive"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			target := ebtypes.Target{}
			err := setInputTraansformation(&target, tt.input, tt.inputPath, tt.inputTransformer)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				if tt.inputTransformer != nil {
					assert.Equal(t, target.InputTransformer.InputTemplate, aws.String(tt.inputTransformer.InputTemplate))
					assert.Equal(t, target.InputTransformer.InputPathsMap, tt.inputTransformer.InputPathsMap)
				}
				if tt.input != "" {
					assert.Equal(t, target.Input, aws.String(tt.input))
				}
				if tt.inputPath != "" {
					assert.Equal(t, target.InputPath, aws.String(tt.inputPath))
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cases := map[string]struct {
		ruleName     string
		eventBusName string
		mockAPI      func(ctx context.Context, ruleName, eventBusName string) *MockEbDeleteRuleAPI
		expectErr    error
	}{
		"should delete rule successfully": {
			ruleName:     testRuleName,
			eventBusName: testEventBusName,
			expectErr:    nil,
			mockAPI: func(ctx context.Context, ruleName, eventBusName string) *MockEbDeleteRuleAPI {
				m := NewMockEbDeleteRuleAPI(t)
				m.EXPECT().
					ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
						Rule:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(&eventbridge.ListTargetsByRuleOutput{
						Targets: []ebtypes.Target{
							{Id: aws.String("123")},
							{Id: aws.String("456")},
						},
					}, nil)

				m.EXPECT().
					RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
						Ids:          []string{"123", "456"},
						Rule:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(&eventbridge.RemoveTargetsOutput{}, nil)

				m.EXPECT().
					DeleteRule(ctx, &eventbridge.DeleteRuleInput{
						Name:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(&eventbridge.DeleteRuleOutput{}, nil)

				return m
			},
		},
		"should return error if ListTargetsByRule failed": {
			ruleName:     testRuleName,
			eventBusName: testEventBusName,
			expectErr:    fmt.Errorf(`failed to delete rule "%v": api ListTargetsByRule failed`, testRuleName),
			mockAPI: func(ctx context.Context, ruleName, eventBusName string) *MockEbDeleteRuleAPI {
				m := NewMockEbDeleteRuleAPI(t)
				m.EXPECT().
					ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
						Rule:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(nil, errors.New("api ListTargetsByRule failed"))

				return m
			},
		},
		"should return error if RemoveTargets failed": {
			ruleName:     testRuleName,
			eventBusName: testEventBusName,
			expectErr:    fmt.Errorf(`failed to delete rule "%v": api RemoveTargets failed`, testRuleName),
			mockAPI: func(ctx context.Context, ruleName, eventBusName string) *MockEbDeleteRuleAPI {
				m := NewMockEbDeleteRuleAPI(t)
				m.EXPECT().
					ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
						Rule:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(&eventbridge.ListTargetsByRuleOutput{
						Targets: []ebtypes.Target{
							{Id: aws.String("123")},
							{Id: aws.String("456")},
						},
					}, nil)

				m.EXPECT().
					RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
						Ids:          []string{"123", "456"},
						Rule:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(nil, errors.New("api RemoveTargets failed"))

				return m
			},
		},
		"should return error if DeleteRule failed": {
			ruleName:     testRuleName,
			eventBusName: testEventBusName,
			expectErr:    fmt.Errorf(`failed to delete rule "%v": api DeleteRule failed`, testRuleName),
			mockAPI: func(ctx context.Context, ruleName, eventBusName string) *MockEbDeleteRuleAPI {
				m := NewMockEbDeleteRuleAPI(t)
				m.EXPECT().
					ListTargetsByRule(ctx, &eventbridge.ListTargetsByRuleInput{
						Rule:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(&eventbridge.ListTargetsByRuleOutput{
						Targets: []ebtypes.Target{
							{Id: aws.String("123")},
							{Id: aws.String("456")},
						},
					}, nil)

				m.EXPECT().
					RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
						Ids:          []string{"123", "456"},
						Rule:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(&eventbridge.RemoveTargetsOutput{}, nil)

				m.EXPECT().
					DeleteRule(ctx, &eventbridge.DeleteRuleInput{
						Name:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(nil, errors.New("api DeleteRule failed"))

				return m
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			mockAPI := tt.mockAPI(ctx, tt.ruleName, tt.eventBusName)
			err := Delete(ctx, mockAPI, tt.eventBusName, tt.ruleName)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Nil(t, err)
			}
			mockAPI.AssertExpectations(t)
		})
	}

}

func TestGet(t *testing.T) {
	cases := map[string]struct {
		mockAPI      func(ctx context.Context, ruleName, eventBusName, arn string) *MockEbDescribeRuleAPI
		ruleName     string
		eventBusName string
		expectErr    error
		expect       *Rule
	}{
		"success": {
			ruleName:     testRuleName,
			eventBusName: testEventBusName,
			expectErr:    nil,
			expect: &Rule{
				Name:         testRuleName,
				EventBusName: testEventBusName,
				EventPattern: "{}",
				ARN: arn.ARN{
					Partition: testPartition,
					Service:   testService,
					Region:    testRegion,
					AccountID: testAccountID,
					Resource:  "rule/" + testEventBusName + "/" + testRuleName,
				},
			},
			mockAPI: func(ctx context.Context, ruleName, eventBusName, arn string) *MockEbDescribeRuleAPI {
				api := NewMockEbDescribeRuleAPI(t)
				api.EXPECT().
					DescribeRule(ctx, &eventbridge.DescribeRuleInput{
						Name:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(&eventbridge.DescribeRuleOutput{
						EventPattern: aws.String("{}"),
						Arn:          aws.String(arn),
					}, nil)
				return api
			},
		},
		"failed to describe rule": {
			ruleName:     testRuleName,
			eventBusName: testEventBusName,
			expectErr:    fmt.Errorf(`failed to describe rule "%v" of event bus "%v": error on aws`, testRuleName, testEventBusName),
			expect:       nil,
			mockAPI: func(ctx context.Context, ruleName, eventBusName, arn string) *MockEbDescribeRuleAPI {
				api := NewMockEbDescribeRuleAPI(t)
				api.EXPECT().
					DescribeRule(ctx, &eventbridge.DescribeRuleInput{
						Name:         aws.String(ruleName),
						EventBusName: aws.String(eventBusName),
					}).
					Return(nil, errors.New("error on aws"))
				return api
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var arn string
			if tt.expect != nil {
				arn = tt.expect.ARN.String()
			}
			api := tt.mockAPI(context.TODO(), tt.ruleName, tt.eventBusName, arn)
			r, err := Get(context.TODO(), api, tt.ruleName, tt.eventBusName)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, r, tt.expect)
			}
		})
	}
}

func TestRule_Resource(t *testing.T) {
	rule := &Rule{
		Name:         testRuleName,
		EventBusName: testEventBusName,
		ARN: arn.ARN{
			Partition: testPartition,
			Service:   testService,
			Region:    testRegion,
			AccountID: testAccountID,
			Resource:  "rule/" + testEventBusName + "/" + testRuleName,
		},
	}
	assert.Equal(t, rule.Resource(), harness.Resource{Type: "AWS::Events::Rule", PhysicalID: testRuleName, ARN: rule.ARN.String()})
}
