// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"iatk/integration/iatk"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/assert"
)

func TestCliErrCases(t *testing.T) {
	cases := map[string]struct {
		input  string
		expect string
	}{
		"not json input": {
			input:  "not json",
			expect: `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"Parse error"}}`,
		},
		"not json rpc input": {
			input:  `{"foo":"bar"}`,
			expect: `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"Parse error"}}`,
		},
		"no method found": {
			input:  `{"jsonrpc": "2.0","id": "42","method": "invalid-hello-world","params": {"name": "Jacob"}}`,
			expect: `{"jsonrpc":"2.0","id":"42","error":{"code":-32601,"message":"Method not found"}}`,
		},
		"valid function but invalid parameters": {
			input:  `{"jsonrpc": "2.0","id": "42","method": "get_physical_id","params": {"LogicalResourceId": "HelloWorldFunction", "StackName": "townhalldemo1","DoesntExist": []}}`,
			expect: `{"jsonrpc":"2.0","id":"42","error":{"code":-32602,"message":"Invalid params"}}`,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var out strings.Builder
			var sErr strings.Builder
			iatk.SInvoke(t, tt.input, &out, &sErr, nil, true)

			actual := strings.Trim(out.String(), "\n")
			assert.Equal(t, tt.expect, actual, fmt.Sprintf("expected: %v, got: %v", tt.expect, actual))
		})
	}

}

func TestCliGetPhysicalIdErrCases(t *testing.T) {
	cases := map[string]struct {
		env    []string
		input  string
		expect string
	}{
		"no creds": {
			env:    []string{"AWS_REGION=us-east-1"},
			input:  `{"jsonrpc": "2.0","id": "42","method": "get_physical_id","params": {"LogicalResourceId": "HelloWorldFunction", "StackName": "ZionDoesntExistStack"}}`,
			expect: `{"jsonrpc":"2.0","id":"42","error":{"code":10,"message":"failed to call service: CloudFormation, operation: DescribeStackResource, error: .*"}}`, // Note(jfuss): depending on setup errors will range. Setting this broad to capture this
		},
		"no creds or region": {
			env:    []string{},
			input:  `{"jsonrpc": "2.0","id": "42","method": "get_physical_id","params": {"LogicalResourceId": "HelloWorldFunction", "StackName": "ZionDoesntExistStack"}}`,
			expect: `{"jsonrpc":"2.0","id":"42","error":{"code":10,"message":"failed to call service: CloudFormation, operation: DescribeStackResource, error: failed to resolve service endpoint, an AWS region is required, but was not found`,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			var out strings.Builder
			var sErr strings.Builder
			iatk.Invoke(t, []byte(tt.input), &out, &sErr, &tt.env)

			re := regexp.MustCompile(tt.expect)
			actual := strings.Trim(out.String(), "\n")
			assert.True(t, re.MatchString(actual), fmt.Sprintf("expected: %v, got: %v", tt.expect, actual))
		})
	}
}

func TestPhysicalIDWithCredsSuite(t *testing.T) {
	suite.Run(t, new(PhysicalIDWithCredsSuite))
}

func (s *PhysicalIDWithCredsSuite) TestCliGetPhysicalIdCredCases() {
	cases := map[string]struct {
		input  string
		expect string
		region string
	}{
		"logical id not found": {
			input:  `{"jsonrpc": "2.0","id": "42","method": "get_physical_id","params": {"LogicalResourceId": "HelloWorldFunction", "StackName": "ZionDoesntExistStack"}}`,
			expect: `{"jsonrpc":"2.0","id":"42","error":{"code":10,"message":"failed to call service: CloudFormation, operation: DescribeStackResource, error: https response error StatusCode: 400, RequestID: [a-z0-9_-]{36}, api error ValidationError: Stack 'ZionDoesntExistStack' does not exist"}}`,
			region: "us-east-1",
		},
		"logical id found": {
			input:  `{"jsonrpc": "2.0","id": "42","method": "get_physical_id","params": {"LogicalResourceId": "SQSQueue", "StackName": "ZionTestStack"}}`,
			expect: `{"jsonrpc":"2.0","id":"42","result":{"output":"https://sqs.us-east-1.amazonaws.com/[0-9]{12}/Zion"}}`,
			region: "us-east-1",
		},
		"Region passed in directly": {
			input:  `{"jsonrpc": "2.0","id": "42","method": "get_physical_id","params": {"LogicalResourceId": "SQSQueue", "StackName": "ZionTestStack", "Region": "us-east-1"}}`,
			expect: `{"jsonrpc":"2.0","id":"42","result":{"output":"https://sqs.us-east-1.amazonaws.com/[0-9]{12}/Zion"}}`,
			region: "us-west-1",
		},
	}

	for name, tt := range cases {
		s.T().Run(name, func(t *testing.T) {
			var out strings.Builder
			var sErr strings.Builder
			tEnv := os.Environ()
			tEnv = append(tEnv, fmt.Sprintf("AWS_REGION=%s", tt.region))
			iatk.Invoke(t, []byte(tt.input), &out, &sErr, &tEnv)

			re := regexp.MustCompile(tt.expect)
			actual := strings.Trim(out.String(), "\n")
			assert.True(t, re.MatchString(actual), fmt.Sprintf("expected: %v, got: %v", tt.expect, actual))
		})
	}
}

type PhysicalIDWithCredsSuite struct {
	suite.Suite
	stackName *string
	region    string
}

func (s *PhysicalIDWithCredsSuite) SetupSuite() {
	stackName := aws.String("ZionTestStack")
	region := "us-east-1"
	s.stackName = aws.String("ZionTestStack")
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	cfnClient := cloudformation.NewFromConfig(cfg)
	templateString := `{"AWSTemplateFormatVersion": "2010-09-09","Description": "simple SQS template","Resources": {"SQSQueue":{"Type" : "AWS::SQS::Queue","Properties":{"QueueName":"Zion"}}},"Outputs": {"QueueURL" : {"Description" : "URL of newly created SQS Queue","Value" : { "Ref" : "SQSQueue" }}}}`

	cfnClient.CreateStack(context.TODO(), &cloudformation.CreateStackInput{
		StackName:    stackName,
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
		StackName: stackName,
	}, *aws.Duration(maxWaitTime))
}

func (s *PhysicalIDWithCredsSuite) TearDownSuite() {
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion(s.region))
	cfnClient := cloudformation.NewFromConfig(cfg)

	cfnClient.DeleteStack(context.TODO(), &cloudformation.DeleteStackInput{
		StackName: s.stackName,
	})

	waiter := cloudformation.NewStackDeleteCompleteWaiter(cfnClient, func(options *cloudformation.StackDeleteCompleteWaiterOptions) {
		options.LogWaitAttempts = false
		options.MaxDelay = 40 * time.Second
	})

	maxWaitTime := 5 * time.Minute
	waiter.Wait(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: s.stackName,
	}, maxWaitTime)
}
