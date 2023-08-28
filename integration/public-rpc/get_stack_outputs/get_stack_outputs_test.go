// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package getstackoutputs_test

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"
	cfn "zion/integration/cloudformation"
	"zion/integration/zion"
	"zion/internal/pkg/aws/config"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
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
	err := cfn.Deploy(s.T(), cfnClient, s.stackName, "./test_stack.json", nil)
	s.Require().NoErrorf(err, "failed to create stack")
}

func (s *GetStackOutputsSuite) TearDownSuite() {
	cfnClient := cloudformation.NewFromConfig(s.cfg)
	err := cfn.Destroy(s.T(), cfnClient, s.stackName)
	s.Require().NoErrorf(err, "failed to destroy stack")
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
			var sErr strings.Builder
			zion.Invoke(t, input, &out, &sErr, nil)
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
			var sErr strings.Builder
			zion.Invoke(t, input, &out, &sErr, nil)
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
