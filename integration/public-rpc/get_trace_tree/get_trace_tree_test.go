// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package gettracetree_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	cfn "zion/integration/cloudformation"
	"zion/integration/zion"
	zioncfn "zion/internal/pkg/cloudformation"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/stepfunctions"
	"github.com/rs/xid"
	"github.com/stretchr/testify/suite"
)

const test_method = "get_trace_tree"

func TestGetTraceTree(t *testing.T) {
	region := "ap-southeast-1"
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		t.Fatalf("failed to get aws config: %v", err)
	}
	cfnClient := cloudformation.NewFromConfig(cfg)
	s := new(GetTraceTreeSuite)
	s.cfnClient = cfnClient
	s.stackName = "test-stack-" + xid.New().String()
	s.region = region
	suite.Run(t, s)
}

type GetTraceTreeSuite struct {
	suite.Suite

	stackName string
	region    string

	cfnClient    *cloudformation.Client
	lambdaClient *lambda.Client
	sfnClient    *stepfunctions.Client

	producerFunctionName string
	stateMachineArn      string
	apiEndpoint          string
}

func (s *GetTraceTreeSuite) SetupSuite() {
	s.T().Log("setup suite start")
	err := cfn.Deploy(
		s.T(),
		s.cfnClient,
		s.stackName,
		"./template.yaml",
		[]types.Capability{
			types.CapabilityCapabilityIam,
			types.CapabilityCapabilityAutoExpand,
		})
	s.Require().NoError(err, "failed to create stack")
	output, _ := zioncfn.GetStackOuput(
		s.stackName,
		[]string{"ProducerFunctionName", "ApiEndpoint", "StateMachineArn"},
		s.cfnClient,
	)
	s.producerFunctionName = output["ProducerFunctionName"]
	s.stateMachineArn = output["StateMachineArn"]
	s.apiEndpoint = output["ApiEndpoint"]
	s.T().Log("setup suite complete")
}

func (s *GetTraceTreeSuite) TearDownSuite() {
	s.T().Log("teardown suite start")
	err := cfn.Destroy(s.T(), s.cfnClient, s.stackName)
	s.Require().NoError(err, "failed to destroy stack")
	s.T().Log("teardown suite complete")
}

func (s *GetTraceTreeSuite) TestPass() {
	s.Assert().Equal(1, 1)
}

func (s *GetTraceTreeSuite) TestInvokeLambda() {
	cases := []struct {
		testname                     string
		invoke                       func(*testing.T) string
		fetchChildTraces             bool
		expectHasChildTrace          bool
		expectSourceTraceNumSegments int
		expectNumPaths               int
	}{
		{
			testname: "invoke lambda and fetch child traces",
			invoke: func(t *testing.T) string {
				t.Logf("invoke lambda function %q", s.producerFunctionName)
				out, err := s.lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
					FunctionName: aws.String(s.producerFunctionName),
					Payload:      []byte("{}"),
				})
				s.Require().NoError(err, "failed to invoke producer lambda function")

				// retrieve tracing header
				tracingHeader := ""
				return tracingHeader

			},
			fetchChildTraces:             true,
			expectNumPaths:               1,
			expectHasChildTrace:          true,
			expectSourceTraceNumSegments: 2,
		},
		{
			testname: "invoke lambda and do not fetch child traces",
			invoke: func(t *testing.T) string {
				t.Logf("invoke lambda function %q", s.producerFunctionName)
				out, err := s.lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
					FunctionName: aws.String(s.producerFunctionName),
					Payload:      []byte("{}"),
				})
				s.Require().NoError(err, "failed to invoke producer lambda function")

				// retrieve tracing header
				tracingHeader := ""
				return tracingHeader

			},
			fetchChildTraces:             false,
			expectNumPaths:               1,
			expectHasChildTrace:          false,
			expectSourceTraceNumSegments: 2,
		},
		{
			testname: "invoke state machine and fetch child traces",
			invoke: func(t *testing.T) string {
				// Invoke StateMachine
				// startExecutionOut, err := s.sfnClient.
				// describe execution
				tracingHeader := ""
				return tracingHeader
			},
			fetchChildTraces:             true,
			expectNumPaths:               2,
			expectHasChildTrace:          false,
			expectSourceTraceNumSegments: 5,
		},
		{
			testname: "invoke state machine and do not fetch child traces",
			invoke: func(t *testing.T) string {
				// Invoke StateMachine
				// startExecutionOut, err := s.sfnClient.
				// describe execution
				tracingHeader := ""
				return tracingHeader
			},
			fetchChildTraces:             false,
			expectNumPaths:               2,
			expectHasChildTrace:          false,
			expectSourceTraceNumSegments: 5,
		},
		{
			testname: "invoke api endpoint and do not fetch child traces",
			invoke: func(t *testing.T) string {
				// Invoke StateMachine
				// startExecutionOut, err := s.sfnClient.
				// describe execution
				tracingHeader := ""
				return tracingHeader
			},
			fetchChildTraces:             false,
			expectNumPaths:               2,
			expectHasChildTrace:          false,
			expectSourceTraceNumSegments: 5,
		},
	}

	for _, tt := range cases {
		s.T().Run(tt.testname, func(t *testing.T) {
			tracingHeader := tt.invoke(t)

			// Get Trace Tree
			tree := s.assertAndReturnTraceTree(tracingHeader, tt.fetchChildTraces)
			if tt.expectHasChildTrace {
				s.Require().Contains(tree, "child_traces")
			} else {
				s.Require().NotContains(tree, "child_traces")
			}
			paths := tree["paths"].([]any)
			s.Equal(tt.expectNumPaths, len(paths), "expected num paths is different than actual")
			sourceTrace := tree["source_trace"].(map[string]any)
			s.Require().Contains(sourceTrace, "segments")
			segments := sourceTrace["segment"].([]any)
			s.Equal(tt.expectSourceTraceNumSegments, len(segments))
		})
	}
}

func (s *GetTraceTreeSuite) TestErrors() {
	cases := []struct {
		testname      string
		request       func() []byte
		expectErrCode int
		expectErrMsg  string
	}{
		{
			testname: "missing Tracing Header",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "42",
					"method": %q,
					"params": {
						"Region": %q
					}
				}`, test_method, s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  `missing required param "TracingHeader"`,
		},
		{
			testname: "invalid tracing header",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "30",
					"method": %q,
					"params": {
						"TracingHeader": "Root=;",
						"FetchChildLinkedTraces": true,
						"Region": %q
					}
				}`, test_method, s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  "invalid TracingHeader",
		},
		{
			testname: "invalid fetch child linked traces",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "020",
					"method": %q,
					"params": {
						"TracingHeader": "Root=1-64e92e5a-10a48563226aa2bd6e930b8c",
						"FetchChildLinkedTraces": 123,
						"Region": %q
					}
				}`, test_method, s.region))
			},
			expectErrCode: -32602,
			expectErrMsg:  "Invalid params",
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

func (s *GetTraceTreeSuite) invoke(req []byte) jsonrpc.Response {
	var stdout strings.Builder
	var stderr strings.Builder
	test := s.T()
	test.Logf("request: %v", string(req))
	zion.Invoke(test, req, &stdout, &stderr, nil)

	test.Logf("response: %v", stdout.String())
	var res jsonrpc.Response
	err := json.Unmarshal([]byte(stdout.String()), &res)
	s.Require().NoError(err, "cannot unmarshal response")
	return res
}

func (s *GetTraceTreeSuite) assertAndReturnTraceTree(tracingHeader string, fetchChildTraces bool) map[string]any {
	req := []byte(fmt.Sprintf(`
	{
		"jsonrpc": "2.0",
		"id": "999999",
		"method": %q,
		"params": {
			"TracingHeader": %q,
			"FetchChildLinkedTraces": %v,
			"Region": %q
		}
	}`, test_method, tracingHeader, fetchChildTraces, s.region))
	s.T().Log("get trace tree")
	res := s.invoke(req)
	s.Require().Nilf(res.Error, "failed to get trace tree: %w", res.Error)
	s.Require().NotNil(res.Result)
	output, ok := res.Result.(map[string]any)["output"].(map[string]any)
	s.Require().True(ok, "output of get_trace_tree must be a map")
	s.Require().Contains(output, "root")
	s.Require().Contains(output, "paths")
	s.Require().Contains(output, "source_traces")
	return output
}
