// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package gettracetree_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
	cfn "zion/integration/cloudformation"
	"zion/integration/zion"
	zioncfn "zion/internal/pkg/cloudformation"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	awshttp "github.com/aws/smithy-go/transport/http"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	s.lambdaClient = lambda.NewFromConfig(cfg)
	s.sfnClient = sfn.NewFromConfig(cfg)
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
	sfnClient    *sfn.Client

	producerFunctionName string
	stateMachineArn      string
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
		[]string{"ProducerFunctionName", "StateMachineArn"},
		s.cfnClient,
	)
	s.Require().Contains(output, "ProducerFunctionName")
	s.Require().Contains(output, "StateMachineArn")
	s.Require().NotZero(output["ProducerFunctionName"])
	s.Require().NotZero(output["StateMachineArn"])
	s.producerFunctionName = output["ProducerFunctionName"]
	s.stateMachineArn = output["StateMachineArn"]
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

func (s *GetTraceTreeSuite) TestInvokeAndGetTraceTree() {
	cases := []struct {
		testname                     string
		invoke                       func(*testing.T) string
		sleep                        int
		expectSourceTraceNumSegments int
		expectNumPaths               int
	}{
		{
			testname: "invoke lambda",
			invoke: func(t *testing.T) string {
				t.Logf("invoke lambda function %q", s.producerFunctionName)
				invokeLambdaOut, err := s.lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
					FunctionName: aws.String(s.producerFunctionName),
					Payload:      []byte("{}"),
				})
				require.NoError(t, err, "failed to invoke producer lambda function")
				rawResponse := middleware.GetRawResponse(invokeLambdaOut.ResultMetadata).(*awshttp.Response)
				tracingHeader := rawResponse.Header["X-Amzn-Trace-Id"][0]
				return tracingHeader

			},
			sleep:                        10,
			expectNumPaths:               1,
			expectSourceTraceNumSegments: 2,
		},
		{
			testname: "invoke state machine",
			invoke: func(t *testing.T) string {
				t.Logf("invoke state machine %q", s.stateMachineArn)
				startExecutionOut, err := s.sfnClient.StartExecution(context.TODO(), &sfn.StartExecutionInput{
					StateMachineArn: aws.String(s.stateMachineArn),
					Input:           aws.String("{}"),
					TraceHeader:     aws.String("Sampled=1"),
				})
				require.NoError(t, err, "failed to start state machine execution")
				describeOut, err := s.sfnClient.DescribeExecution(context.TODO(), &sfn.DescribeExecutionInput{
					ExecutionArn: startExecutionOut.ExecutionArn,
				})
				require.NoError(t, err, "failed to describe state machine execution")
				tracingHeader := aws.ToString(describeOut.TraceHeader)
				return tracingHeader
			},
			sleep:                        10,
			expectNumPaths:               2,
			expectSourceTraceNumSegments: 5,
		},
	}

	for _, tt := range cases {
		s.T().Run(tt.testname, func(t *testing.T) {
			tracingHeader := tt.invoke(t)
			t.Logf("tracing header: %v", tracingHeader)
			t.Logf("sleeping for %v seconds", tt.sleep)
			time.Sleep(time.Duration(tt.sleep) * time.Second)

			// Get Trace Tree
			tree := s.assertAndReturnTraceTree(tracingHeader)
			paths := tree["paths"].([]any)
			assert.Equal(t, tt.expectNumPaths, len(paths), "expected num paths is different than actual")
			sourceTrace := tree["source_trace"].(map[string]any)
			require.Contains(t, sourceTrace, "segments")
			segments := sourceTrace["segments"].([]any)
			assert.Equal(t, tt.expectSourceTraceNumSegments, len(segments))
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
						"Region": %q
					}
				}`, test_method, s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  "invalid tracing header provided",
		},
	}

	for _, tt := range cases {
		s.T().Run(tt.testname, func(t *testing.T) {
			req := tt.request()
			res := s.invoke(req)
			require.NotNil(t, res.Error)
			assert.Equal(t, tt.expectErrCode, res.Error.Code)
			assert.Contains(t, res.Error.Message, tt.expectErrMsg)
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

func (s *GetTraceTreeSuite) assertAndReturnTraceTree(tracingHeader string) map[string]any {
	req := []byte(fmt.Sprintf(`
	{
		"jsonrpc": "2.0",
		"id": "999999",
		"method": %q,
		"params": {
			"TracingHeader": %q,
			"Region": %q
		}
	}`, test_method, tracingHeader, s.region))
	t := s.T()
	t.Log("get trace tree")
	res := s.invoke(req)
	require.Nilf(t, res.Error, "failed to get trace tree: %w", res.Error)
	require.NotNil(t, res.Result)
	output, ok := res.Result.(map[string]any)["output"].(map[string]any)
	require.True(t, ok, "output of get_trace_tree must be a map")
	require.Contains(t, output, "root")
	require.Contains(t, output, "paths")
	require.Contains(t, output, "source_trace")
	return output
}
