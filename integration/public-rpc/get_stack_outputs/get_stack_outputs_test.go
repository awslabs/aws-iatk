// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package getstackoutputs_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"os/exec"
	"strings"
	"testing"
	"time"
	"zion/internal/pkg/aws/config"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestGetStackOutputs(t *testing.T) {
	s := &GetStackOutputsSuite{
		stackName: "test-stack-" + uuid.NewString(),
		region:    "us-west-2",
	}
	suite.Run(t, s)
}

type GetStackOutputsSuite struct {
	suite.Suite
	stackName string
	region    string
	cfg       aws.Config
}

func (s *GetStackOutputsSuite) setAWSConfig() {
	cfg, err := config.GetAWSConfig(context.TODO(), s.region, "")
	if err != nil {
		s.T().Fatalf("failed to get aws config: %v", err)
	}
	s.cfg = cfg
}

func (s *GetStackOutputsSuite) SetupSuite() {
	s.setAWSConfig()
	cfnClient := cloudformation.NewFromConfig(s.cfg)
	templateString := `{"AWSTemplateFormatVersion": "2010-09-09","Description": "simple SQS template","Resources": {"SQSQueue":{"Type" : "AWS::SQS::Queue"}},"Outputs": {"QueueURL" : {"Description" : "URL of newly created SQS Queue","Value" : { "Ref" : "SQSQueue" }}, "QueueURLFromGetAtt": {"Description": "Queue URL", "Value": {"Fn::GetAtt": ["SQSQueue", "QueueUrl"]}}, "QueueArn": {"Description": "Queue ARN", "Value": {"Fn::GetAtt": ["SQSQueue", "Arn"]}}}}`

	log.Printf("create stack %q", s.stackName)
	cfnClient.CreateStack(context.TODO(), &cloudformation.CreateStackInput{
		StackName:    aws.String(s.stackName),
		TemplateBody: aws.String(templateString),
	})

	createRetryable := func(
		ctx context.Context,
		params *cloudformation.DescribeStacksInput,
		output *cloudformation.DescribeStacksOutput,
		err error,
	) (bool, error) {
		if output.Stacks != nil {
			for _, stack := range output.Stacks {
				switch stack.StackStatus {
				case types.StackStatusCreateInProgress:
					return true, nil
				case types.StackStatusCreateFailed:
					return false, errors.New(*stack.StackStatusReason)
				case types.StackStatusCreateComplete:
					return false, nil
				default:
					return false, nil
				}
			}
		}
		return false, err
	}

	maxWaitTime := 5 * time.Minute
	waiter := cloudformation.NewStackCreateCompleteWaiter(cfnClient, func(o *cloudformation.StackCreateCompleteWaiterOptions) {
		o.Retryable = createRetryable
	})

	waiter.Wait(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackName),
	}, *aws.Duration(maxWaitTime))
	log.Printf("completed create stack %q", s.stackName)
}

func (s *GetStackOutputsSuite) TearDownSuite() {
	cfnClient := cloudformation.NewFromConfig(s.cfg)

	log.Printf("destroy stack %q", s.stackName)
	cfnClient.DeleteStack(context.TODO(), &cloudformation.DeleteStackInput{
		StackName: aws.String(s.stackName),
	})

	waiter := cloudformation.NewStackDeleteCompleteWaiter(cfnClient, func(options *cloudformation.StackDeleteCompleteWaiterOptions) {
		options.LogWaitAttempts = false
		options.MaxDelay = 40 * time.Second
	})

	maxWaitTime := 5 * time.Minute
	waiter.Wait(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: aws.String(s.stackName),
	}, maxWaitTime)
	log.Printf("completed destroy stack %q", s.stackName)
}

func (s *GetStackOutputsSuite) TestGetStackOutputs() {
	cases := map[string]struct {
		outputs []string
		input   func(outputs []string) []byte
	}{
		"should succeed with all outputs": {
			outputs: []string{"QueueURL", "QueueURLFromGetAtt", "QueueArn"},
			input: func(outputs []string) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  "get_stack_outputs",
					"params": map[string]interface{}{
						"StackName":   s.stackName,
						"OutputNames": outputs,
						"Region":      s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
		"should succeed with some outputs": {
			outputs: []string{"QueueURL", "QueueURLFromGetAtt"},
			input: func(outputs []string) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  "get_stack_outputs",
					"params": map[string]interface{}{
						"StackName":   s.stackName,
						"OutputNames": outputs,
						"Region":      s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
		"should succeed with duplicate keys": {
			outputs: []string{"QueueURL", "QueueURL", "QueueArn", "QueueArn"},
			input: func(outputs []string) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  "get_stack_outputs",
					"params": map[string]interface{}{
						"StackName":   s.stackName,
						"OutputNames": outputs,
						"Region":      s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
		"should succeed with empty inputs": {
			outputs: []string{},
			input: func(outputs []string) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  "get_stack_outputs",
					"params": map[string]interface{}{
						"StackName":   s.stackName,
						"OutputNames": outputs,
						"Region":      s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
	}

	for name, tt := range cases {
		s.T().Run(name, func(t *testing.T) {
			input := tt.input(tt.outputs)
			log.Printf("request: %v", string(input))
			var out strings.Builder
			invoke(t, input, &out)
			log.Printf("response: %v", out.String())
			var actual jsonrpc.Response
			json.Unmarshal([]byte(out.String()), &actual)
			result := actual.Result
			if result == nil {
				assert.FailNow(t, "expect result to be not nil")
			}
			output := result.(map[string]interface{})["output"]
			for _, key := range tt.outputs {
				_, ok := output.(map[string]interface{})[key]
				assert.True(t, ok)
			}
		})
	}
}

func (s *GetStackOutputsSuite) TestErrGetStackOutputs() {
	cases := map[string]struct {
		outputs      []string
		expectErrMsg string
		input        func(outputs []string) []byte
	}{
		"should fail due to not all output keys found": {
			expectErrMsg: "Not all output keys found",
			outputs:      []string{"QueueURL", "Foo"},
			input: func(outputs []string) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  "get_stack_outputs",
					"params": map[string]interface{}{
						"StackName":   s.stackName,
						"OutputNames": outputs,
						"Region":      s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
	}

	for name, tt := range cases {
		s.T().Run(name, func(t *testing.T) {
			input := tt.input(tt.outputs)
			log.Printf("request: %v", string(input))
			var out strings.Builder
			invoke(t, input, &out)
			log.Printf("response: %v", out.String())
			var actual jsonrpc.Response
			json.Unmarshal([]byte(out.String()), &actual)
			if actual.Error == nil {
				assert.FailNow(t, "expect error to be not nil")
			}
			if tt.expectErrMsg != "" {
				assert.Equal(t, tt.expectErrMsg, actual.Error.Message)
			}
		})
	}
}

func invoke(t *testing.T, in []byte, out *strings.Builder) {
	cmd := exec.Command("../../../bin/zion")
	cmd.Stdin = bytes.NewReader(in)
	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		t.Fatalf("command run failed: %v", err)
	}
}
