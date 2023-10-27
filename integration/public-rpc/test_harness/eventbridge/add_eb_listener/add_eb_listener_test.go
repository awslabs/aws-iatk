// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package addeblistener_test

import (
	"context"
	"ctk/integration/ctk"
	"ctk/internal/pkg/jsonrpc"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/rs/xid"
	"github.com/stretchr/testify/suite"
)

const method string = "test_harness.eventbridge.add_listener"

func TestAddEbListener(t *testing.T) {
	region := "us-west-2"
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		t.Fatalf("failed to get aws config: %v", err)
	}
	ebClient := eventbridge.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)
	snsClient := sns.NewFromConfig(cfg)

	s := new(AddEbListenerSuite)
	s.ebClient = ebClient
	s.sqsClient = sqsClient
	s.snsClient = snsClient
	s.eventBusName = "eb-" + xid.New().String()
	s.eventBusRule = "eb-testrule"
	s.eventBusTarget = "eb-testtarget"
	s.targetInputPathMaps = map[string]string{"detail-type": "$.detail-type"}
	s.targetInputTemplate = "\"This event was of <detail-type> type.\""
	s.region = region

	suite.Run(t, s)
}

type AddEbListenerSuite struct {
	suite.Suite
	eventBusName        string
	eventBusRule        string
	eventBusTarget      string
	region              string
	snsTopicArn         *string
	targetInputTemplate string
	targetInputPathMaps map[string]string
	ebClient            *eventbridge.Client
	sqsClient           *sqs.Client
	snsClient           *sns.Client
	queueURLs           []string
	ruleNames           []string
}

func (s *AddEbListenerSuite) SetupSuite() {
	s.T().Log("setup suite start")
	topic, err := s.snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String("MySnsTopic"),
	})
	s.Require().NoErrorf(err, "failed to create sns topic: %v", err)
	s.snsTopicArn = topic.TopicArn

	_, err = s.ebClient.CreateEventBus(context.TODO(), &eventbridge.CreateEventBusInput{
		Name: aws.String(s.eventBusName),
	})
	s.Require().NoErrorf(err, "failed to create event bus: %v", err)
	_, err = s.ebClient.PutRule(context.TODO(), &eventbridge.PutRuleInput{
		Name:         aws.String(s.eventBusRule),
		EventBusName: aws.String(s.eventBusName),
		EventPattern: aws.String("{\"detail-type\": [\"customerCreated\"], \"source\": [\"aws.events\"]}"),
	})
	s.Require().NoErrorf(err, "failed to create eventbridge rule: %v", err)

	_, err = s.ebClient.PutTargets(context.TODO(), &eventbridge.PutTargetsInput{
		EventBusName: aws.String(s.eventBusName),
		Rule:         aws.String(s.eventBusRule),
		Targets: []ebtypes.Target{{
			Id:  aws.String(s.eventBusTarget),
			Arn: topic.TopicArn,
			InputTransformer: &ebtypes.InputTransformer{
				InputPathsMap: s.targetInputPathMaps,
				InputTemplate: aws.String(s.targetInputTemplate),
			}}},
	})

	s.Require().NoErrorf(err, "failed to create eventbridge target: %v", err)
	s.T().Log("setup suite complete")
}

func (s *AddEbListenerSuite) TearDownSuite() {
	s.T().Log("tear down suite start")
	for _, ruleName := range s.ruleNames {
		err := deleteRuleAndTargets(s.ebClient, ruleName, s.eventBusName)
		s.Require().NoErrorf(err, "failed to delete rule %v/%v", s.eventBusName, ruleName)
	}

	for _, queueURL := range s.queueURLs {
		err := deleteQueue(s.sqsClient, queueURL)
		s.Require().NoErrorf(err, "failed to delete queue %v", queueURL)
	}

	_, err := s.ebClient.RemoveTargets(context.TODO(), &eventbridge.RemoveTargetsInput{
		Ids:          []string{s.eventBusTarget},
		Rule:         aws.String(s.eventBusRule),
		EventBusName: aws.String(s.eventBusName),
	})
	s.Require().NoErrorf(err, "failed to remove targets %v", s.eventBusTarget)

	_, err = s.ebClient.DeleteRule(context.TODO(), &eventbridge.DeleteRuleInput{
		Name:         aws.String(s.eventBusRule),
		EventBusName: aws.String(s.eventBusName),
	})
	s.Require().NoErrorf(err, "failed to delete rule %v", s.eventBusRule)

	_, err = s.snsClient.DeleteTopic(context.TODO(), &sns.DeleteTopicInput{
		TopicArn: s.snsTopicArn,
	})
	s.Require().NoErrorf(err, "failed to delete sns topic %v", s.snsTopicArn)

	err = deleteEventBus(s.ebClient, s.eventBusName)
	s.Require().NoErrorf(err, "failed to delete bus %v", s.eventBusName)
	s.T().Log("tear down suite complete")
}

func (s *AddEbListenerSuite) TestAddEbListenerNoTarget() {
	cases := []struct {
		testname string
		request  func(tags map[string]string) []byte
		tags     map[string]string
	}{
		{
			testname: "add listener without custom tags",
			request: func(tags map[string]string) []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"RuleName":     s.eventBusRule,
						"Region":       s.region,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
		},
		{
			testname: "add listener with custom tags",
			request: func(tags map[string]string) []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"RuleName":     s.eventBusRule,
						"Region":       s.region,
						"Tags":         tags,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			tags: map[string]string{
				"stage": "test",
			},
		},
	}

	for _, tt := range cases {
		s.Run(tt.testname, func() {
			req := tt.request(tt.tags)
			res := s.invoke(req)
			s.Require().Nilf(res.Error, "expect error to be nil, actual: %v", res.Error)
			queueURL, queueARN, ruleName, ruleARN, expectTags := s.assertOutput(res, tt.tags)
			s.queueURLs = append(s.queueURLs, queueURL)
			s.ruleNames = append(s.ruleNames, ruleName)

			s.assertQueue(queueURL, queueARN, ruleARN, expectTags)
			s.assertRule(ruleName, s.eventBusName, expectTags)
			s.assertRuleTarget(ruleName, s.eventBusName, false)
		})
	}
}

func (s *AddEbListenerSuite) TestAddEbListenerWithTarget() {
	cases := []struct {
		testname string
		request  func() []byte
	}{
		{
			testname: "add listener with TargetId.InputTransformer",
			request: func() []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"TargetId":     s.eventBusTarget,
						"RuleName":     s.eventBusRule,
						"Region":       s.region,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
		},
	}

	for _, tt := range cases {
		s.Run(tt.testname, func() {
			req := tt.request()
			res := s.invoke(req)
			s.Require().Nilf(res.Error, "expect error to be nil, actual: %v", res.Error)
			queueURL, queueARN, ruleName, ruleARN, expectTags := s.assertOutput(res, nil)
			s.queueURLs = append(s.queueURLs, queueURL)
			s.ruleNames = append(s.ruleNames, ruleName)

			s.assertQueue(queueURL, queueARN, ruleARN, expectTags)
			s.assertRule(ruleName, s.eventBusName, expectTags)
			s.assertRuleTarget(ruleName, s.eventBusName, true)
		})
	}
}

func (s *AddEbListenerSuite) TestErrors() {
	cases := []struct {
		testname      string
		request       func() []byte
		expectErrCode int
		expectErrMsg  string
	}{
		{
			testname: "missing event bus name",
			request: func() []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"Region":   s.region,
						"RuleName": s.eventBusRule,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			expectErrCode: 10,
			expectErrMsg:  `missing required param "EventBusName"`,
		},
		{
			testname: "missing rule name",
			request: func() []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"Region":       s.region,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			expectErrCode: 10,
			expectErrMsg:  `missing required param "RuleName"`,
		},
		{
			testname: "custom tags contain reserved key",
			request: func() []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"RuleName":     s.eventBusRule,
						"Region":       s.region,
						"Tags": map[string]string{
							"zion:TestHarness:Created": "12345",
						},
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			expectErrCode: 10,
			expectErrMsg:  `invalid tags: reserved tag key "zion:TestHarness:Created" found in provided tags`,
		},
		{
			testname: "provides invalid targetid",
			request: func() []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"Region":       s.region,
						"TargetId":     "DoesNotExistTarget",
						"RuleName":     s.eventBusRule,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			expectErrCode: 10,
			expectErrMsg:  `failed to locate test target: failed to create resource group: TagetId: DoesNotExistTarget was not found on eb-testrule Rule`,
		},
		{
			testname: "provides invalid ruleName",
			request: func() []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"Region":       s.region,
						"TargetId":     s.eventBusTarget,
						"RuleName":     "DoesNotExistTarget",
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			expectErrCode: 10,
			expectErrMsg:  `failed to locate test target: RuleName "DoesNotExistTarget" was provided but not found`,
		},
	}

	for _, tt := range cases {
		s.Run(tt.testname, func() {
			req := tt.request()
			res := s.invoke(req)
			s.Require().NotNil(res.Error)
			s.Equal(tt.expectErrCode, res.Error.Code)
			s.Contains(res.Error.Message, tt.expectErrMsg)
		})
	}
}

func (s *AddEbListenerSuite) invoke(req []byte) jsonrpc.Response {
	var out strings.Builder
	var sErr strings.Builder
	test := s.T()
	ctk.Invoke(test, req, &out, &sErr, nil)

	test.Logf("response: %v", out.String())
	test.Logf("err: %v", sErr.String())

	var res jsonrpc.Response
	err := json.Unmarshal([]byte(out.String()), &res)
	s.Require().NoError(err, "cannot unmarshal response")
	return res
}

func (s *AddEbListenerSuite) assertOutput(response jsonrpc.Response, customTags map[string]string) (string, string, string, string, map[string]string) {
	s.Require().Contains(response.Result.(map[string]any), "output", "response result must contain output")
	output := response.Result.(map[string]any)["output"].(map[string]any)
	s.Require().Contains(output, "TargetUnderTest")
	s.Require().Contains(output, "Components")
	components := output["Components"].([]any)
	s.Require().Len(components, 2)
	queueURL := components[0].(map[string]any)["PhysicalID"].(string)
	queueARN := components[0].(map[string]any)["ARN"].(string)
	ruleName := components[1].(map[string]any)["PhysicalID"].(string)
	ruleARN := components[1].(map[string]any)["ARN"].(string)

	id := output["Id"].(string)
	ebARN := output["TargetUnderTest"].(map[string]any)["ARN"].(string)
	expectTags := map[string]string{
		"zion:TestHarness:ID":      id,
		"zion:TestHarness:Target":  ebARN,
		"zion:TestHarness:Type":    "EventBridge.Listener",
		"zion:TestHarness:Created": "",
	}
	for key, val := range customTags {
		if _, ok := expectTags[key]; !ok {
			expectTags[key] = val
		}
	}
	return queueURL, queueARN, ruleName, ruleARN, expectTags
}

func (s *AddEbListenerSuite) assertQueue(queueURL, queueARN, ruleARN string, expectTags map[string]string) {
	out, err := s.sqsClient.GetQueueAttributes(context.TODO(), &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueURL),
		AttributeNames: []sqstypes.QueueAttributeName{
			sqstypes.QueueAttributeNameAll,
		},
	})
	s.Require().NoError(err)
	queuePolicy := out.Attributes["Policy"]
	s.Require().Contains(queuePolicy, fmt.Sprintf(`"Resource":%q`, queueARN), "expected queue policy to contain queue arn")
	s.Require().Contains(queuePolicy, fmt.Sprintf(`"aws:SourceArn":%q`, ruleARN), "expected queue policy to contain rule arn")

	tags, err := s.sqsClient.ListQueueTags(context.TODO(), &sqs.ListQueueTagsInput{
		QueueUrl: aws.String(queueURL),
	})
	s.Require().NoError(err)
	s.assertTags(expectTags, tags.Tags)
}

func (s *AddEbListenerSuite) assertRule(ruleName, eventBusName string, expectTags map[string]string) {
	r, err := s.ebClient.DescribeRule(context.TODO(), &eventbridge.DescribeRuleInput{
		Name:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	})
	s.Require().NoError(err)
	s.Equal(ebtypes.RuleStateEnabled, r.State)

	tags, err := s.ebClient.ListTagsForResource(context.TODO(), &eventbridge.ListTagsForResourceInput{
		ResourceARN: r.Arn,
	})
	s.Require().NoError(err)
	actualTags := make(map[string]string)
	for _, tag := range tags.Tags {
		actualTags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	s.assertTags(expectTags, actualTags)
}

func (s *AddEbListenerSuite) assertRuleTarget(ruleName, eventBusName string, hasTarget bool) {
	targets, err := s.ebClient.ListTargetsByRule(context.TODO(), &eventbridge.ListTargetsByRuleInput{
		Rule:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	})
	s.Require().NoError(err)

	s.Len(targets.Targets, 1, "expect to have 1 target only")
	target := targets.Targets[0]

	if hasTarget {
		s.Require().Equal("\"This event was of <detail-type> type.\"", aws.ToString(target.InputTransformer.InputTemplate))
		s.Require().Equal(map[string]string{"detail-type": "$.detail-type"}, target.InputTransformer.InputPathsMap)
	}
}

func (s *AddEbListenerSuite) assertTags(expectTags, actualTags map[string]string) {
	for key, val := range expectTags {
		s.Contains(actualTags, key)
		if key != "zion:TestHarness:Created" {
			s.Equal(val, actualTags[key])
		}
	}
}

func deleteQueue(client *sqs.Client, queueURL string) error {
	_, err := client.DeleteQueue(context.TODO(), &sqs.DeleteQueueInput{
		QueueUrl: aws.String(queueURL),
	})
	return err
}

func deleteRuleAndTargets(client *eventbridge.Client, ruleName, eventBusName string) error {
	targets, err := client.ListTargetsByRule(context.TODO(), &eventbridge.ListTargetsByRuleInput{
		Rule:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	})
	if err != nil {
		return err
	}
	var ids []string
	for _, target := range targets.Targets {
		ids = append(ids, aws.ToString(target.Id))
	}

	if len(ids) > 0 {
		_, err = client.RemoveTargets(context.TODO(), &eventbridge.RemoveTargetsInput{
			Ids:          ids,
			Rule:         aws.String(ruleName),
			EventBusName: aws.String(eventBusName),
		})
		if err != nil {
			return err
		}
	}

	_, err = client.DeleteRule(context.TODO(), &eventbridge.DeleteRuleInput{
		Name:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	})
	return err
}

func deleteEventBus(client *eventbridge.Client, eventBusName string) error {
	_, err := client.DeleteEventBus(context.TODO(), &eventbridge.DeleteEventBusInput{
		Name: aws.String(eventBusName),
	})
	return err
}
