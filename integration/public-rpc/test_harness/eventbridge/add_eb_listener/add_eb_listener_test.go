// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package addeblistener_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
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

	s := new(AddEbListenerSuite)
	s.ebClient = ebClient
	s.sqsClient = sqsClient
	s.eventBusName = "eb-" + xid.New().String()
	s.region = region

	suite.Run(t, s)
}

type AddEbListenerSuite struct {
	suite.Suite
	eventBusName string
	region       string
	ebClient     *eventbridge.Client
	sqsClient    *sqs.Client
	queueURLs    []string
	ruleNames    []string
}

func (s *AddEbListenerSuite) SetupSuite() {
	s.T().Log("setup suite start")
	_, err := s.ebClient.CreateEventBus(context.TODO(), &eventbridge.CreateEventBusInput{
		Name: aws.String(s.eventBusName),
	})
	s.Require().NoErrorf(err, "failed to create event bus: %v", err)
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

	err := deleteEventBus(s.ebClient, s.eventBusName)
	s.Require().NoErrorf(err, "failed to delete bus %v", s.eventBusName)
	s.T().Log("tear down suite complete")
}

func (s *AddEbListenerSuite) TestAddEbListenerNoInputTransformation() {
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
						"EventPattern": `{"source":[{"prefix":""}]}`,
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
						"EventPattern": `{"source":[{"prefix":""}]}`,
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
			// t := s.T()
			req := tt.request(tt.tags)
			res := s.invoke(req)
			s.Require().Nilf(res.Error, "expect error to be nil, actual: %v", res.Error)
			queueURL, queueARN, ruleName, ruleARN, expectTags := s.assertOutput(res, tt.tags)
			s.queueURLs = append(s.queueURLs, queueURL)
			s.ruleNames = append(s.ruleNames, ruleName)

			s.assertQueue(queueURL, queueARN, ruleARN, expectTags)
			s.assertRule(ruleName, s.eventBusName, expectTags)
			s.assertRuleTarget(ruleName, s.eventBusName, "", "", nil)
		})
	}
}

func (s *AddEbListenerSuite) TestAddEbListenerWithInputTransformation() {
	cases := []struct {
		testname         string
		request          func(input, inputPath string, inputTransformer map[string]any) []byte
		input            string
		inputPath        string
		inputTransformer map[string]any
	}{
		{
			testname: "add listener with Input",
			request: func(input, inputPath string, inputTransformer map[string]any) []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"EventPattern": `{"source":[{"prefix":"com.test"}]}`,
						"Input":        input,
						"Region":       s.region,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			input: `{"message": "hello, world!"}`,
		},
		{
			testname: "add listener with InputPath",
			request: func(input, inputPath string, inputTransformer map[string]any) []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"EventPattern": `{"source":[{"prefix":"com.test"}]}`,
						"InputPath":    inputPath,
						"Region":       s.region,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			inputPath: "$.detail.id",
		},
		{
			testname: "add listener with InputTransformer",
			request: func(input, inputPath string, inputTransformer map[string]any) []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName":     s.eventBusName,
						"EventPattern":     `{"source":[{"prefix":"com.test"}]}`,
						"InputTransformer": inputTransformer,
						"Region":           s.region,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			inputTransformer: map[string]any{
				"InputTemplate": `{"id": "<id>", "foo": "<foo>"}`,
				"InputPathsMap": map[string]string{
					"foo": "$.detail.foo",
					"id":  "$.detail.id",
				},
			},
		},
	}

	for _, tt := range cases {
		s.Run(tt.testname, func() {
			// t := s.T()
			req := tt.request(tt.input, tt.inputPath, tt.inputTransformer)
			res := s.invoke(req)
			s.Require().Nilf(res.Error, "expect error to be nil, actual: %v", res.Error)
			queueURL, queueARN, ruleName, ruleARN, expectTags := s.assertOutput(res, nil)
			s.queueURLs = append(s.queueURLs, queueURL)
			s.ruleNames = append(s.ruleNames, ruleName)

			s.assertQueue(queueURL, queueARN, ruleARN, expectTags)
			s.assertRule(ruleName, s.eventBusName, expectTags)
			s.assertRuleTarget(ruleName, s.eventBusName, tt.input, tt.inputPath, tt.inputTransformer)
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
						"EventPattern": "{}",
						"Region":       s.region,
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			expectErrCode: 10,
			expectErrMsg:  `missing required param "EventBusName"`,
		},
		{
			testname: "missing event pattern",
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
			expectErrMsg:  `missing required param "EventPattern"`,
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
						"EventPattern": `{"source":[{"prefix":"com.test"}]}`,
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
			testname: "provides both input and inputpath",
			request: func() []byte {
				r := map[string]any{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  method,
					"params": map[string]any{
						"EventBusName": s.eventBusName,
						"EventPattern": `{"source":[{"prefix":"com.test"}]}`,
						"Region":       s.region,
						"Input":        "{}",
						"InputPath":    "$.detail-type",
					},
				}
				out, _ := json.Marshal(r)
				return out
			},
			expectErrCode: 10,
			expectErrMsg:  `provide only one of "Input", "InputPath" and "InputTransformer"`,
		},
	}

	for _, tt := range cases {
		s.Run(tt.testname, func() {
			req := tt.request()
			res := s.invoke(req)
			s.Require().NotNil(res.Error)
			s.Equal(tt.expectErrCode, res.Error.Code)
			s.Contains(tt.expectErrMsg, res.Error.Message)
		})
	}
}

func (s *AddEbListenerSuite) invoke(req []byte) jsonrpc.Response {
	var out strings.Builder
	err := invoke(req, &out)
	s.Require().NoError(err)

	s.T().Logf("response: %v", out.String())
	var res jsonrpc.Response
	err = json.Unmarshal([]byte(out.String()), &res)
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

func (s *AddEbListenerSuite) assertRuleTarget(ruleName, eventBusName, input, inputPath string, inputTransformer map[string]any) {
	targets, err := s.ebClient.ListTargetsByRule(context.TODO(), &eventbridge.ListTargetsByRuleInput{
		Rule:         aws.String(ruleName),
		EventBusName: aws.String(eventBusName),
	})
	s.Require().NoError(err)

	s.Len(targets.Targets, 1, "expect to have 1 target only")
	target := targets.Targets[0]

	if input != "" {
		s.Require().Equal(input, aws.ToString(target.Input))
	}

	if inputPath != "" {
		s.Require().Equal(inputPath, aws.ToString(target.InputPath))
	}

	if inputTransformer != nil {
		s.Require().Equal(inputTransformer["InputTemplate"].(string), aws.ToString(target.InputTransformer.InputTemplate))
		s.Require().Equal(inputTransformer["InputPathsMap"].(map[string]string), target.InputTransformer.InputPathsMap)
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

func invoke(in []byte, out *strings.Builder) error {
	cmd := exec.Command("../../../../../bin/zion")
	cmd.Stdin = bytes.NewReader(in)
	cmd.Stdout = out
	err := cmd.Run()
	return err
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
