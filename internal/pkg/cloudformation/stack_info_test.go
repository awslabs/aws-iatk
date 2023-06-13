package cloudformation

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/assert"
)

type mockDescribeStackResourceAPI func(ctx context.Context, params *cloudformation.DescribeStackResourceInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackResourceOutput, error)

func (m mockDescribeStackResourceAPI) DescribeStackResource(ctx context.Context, params *cloudformation.DescribeStackResourceInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackResourceOutput, error) {
	return m(ctx, params, optFns...)
}

type mockDescribeStacksAPI func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)

func (m mockDescribeStacksAPI) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return m(ctx, params, optFns...)
}

func TestGetPhysicalId(t *testing.T) {
	cases := []struct {
		client    func(t *testing.T) DescribeStackResourceAPI
		stackName string
		logicalID string
		expect    string
	}{
		{
			client: func(t *testing.T) DescribeStackResourceAPI {
				return mockDescribeStackResourceAPI(func(ctx context.Context, params *cloudformation.DescribeStackResourceInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackResourceOutput, error) {
					t.Helper()
					if params.LogicalResourceId == nil {
						t.Fatal("expect LogicalResourceId to not be nil")
					}
					if e, a := "LogicalId", *params.LogicalResourceId; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}
					if params.StackName == nil {
						t.Fatal("expect StackName to not be nil")
					}
					if e, a := "StackName", *params.StackName; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}

					return &cloudformation.DescribeStackResourceOutput{
						StackResourceDetail: &types.StackResourceDetail{
							PhysicalResourceId: aws.String("LogicalResourceId"),
						},
					}, nil
				})
			},
			logicalID: "LogicalId",
			stackName: "StackName",
			expect:    "LogicalResourceId",
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			id, err := GetPhysicalId(tt.stackName, tt.logicalID, tt.client(t))

			assert.Nil(t, err, "Expected err to be nil, got %v", err)
			assert.Equal(t, tt.expect, id)
		})
	}
}

func TestErrGetPhysicalId(t *testing.T) {
	cases := []struct {
		client    func(t *testing.T) DescribeStackResourceAPI
		stackName string
		logicalID string
		expect    string
	}{
		{
			client: func(t *testing.T) DescribeStackResourceAPI {
				return mockDescribeStackResourceAPI(func(ctx context.Context, params *cloudformation.DescribeStackResourceInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackResourceOutput, error) {
					t.Helper()
					if params.LogicalResourceId == nil {
						t.Fatal("expect LogicalResourceId to not be nil")
					}
					if e, a := "LogicalId", *params.LogicalResourceId; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}
					if params.StackName == nil {
						t.Fatal("expect StackName to not be nil")
					}
					if e, a := "StackName", *params.StackName; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}

					return nil, errors.New("This failed")
				})
			},
			logicalID: "LogicalId",
			stackName: "StackName",
			expect:    "This failed",
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			_, err := GetPhysicalId(tt.stackName, tt.logicalID, tt.client(t))

			assert.Equal(t, tt.expect, err.Error())
		})
	}
}

func TestGetStackOutput(t *testing.T) {
	cases := map[string]struct {
		client     func(t *testing.T) DescribeStacksAPI
		stackName  string
		outputKeys []string
		expect     map[string]string
	}{
		"Sucess": {
			client: func(t *testing.T) DescribeStacksAPI {
				return mockDescribeStacksAPI(func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					t.Helper()
					if params.StackName == nil {
						t.Fatal("expect StackName to not be nil")
					}
					if e, a := "StackName", *params.StackName; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}

					return &cloudformation.DescribeStacksOutput{
						NextToken: nil,
						Stacks: []types.Stack{
							{
								StackName: params.StackName,
								Outputs: []types.Output{
									{
										OutputKey:   aws.String("Queue"),
										OutputValue: aws.String("value1"),
									},
									{
										OutputKey:   aws.String("Function"),
										OutputValue: aws.String("value2"),
									},
								},
							},
						},
					}, nil
				})
			},
			stackName:  "StackName",
			outputKeys: []string{"Queue", "Function"},
			expect:     map[string]string{"Queue": "value1", "Function": "value2"},
		},
		"Sucess with duplicate OuputKeys": {
			client: func(t *testing.T) DescribeStacksAPI {
				return mockDescribeStacksAPI(func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					t.Helper()
					if params.StackName == nil {
						t.Fatal("expect StackName to not be nil")
					}
					if e, a := "StackName", *params.StackName; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}

					return &cloudformation.DescribeStacksOutput{
						NextToken: nil,
						Stacks: []types.Stack{
							{
								StackName: params.StackName,
								Outputs: []types.Output{
									{
										OutputKey:   aws.String("Queue"),
										OutputValue: aws.String("value1"),
									},
									{
										OutputKey:   aws.String("Function"),
										OutputValue: aws.String("value2"),
									},
								},
							},
						},
					}, nil
				})
			},
			stackName:  "StackName",
			outputKeys: []string{"Queue", "Function", "Function"},
			expect:     map[string]string{"Queue": "value1", "Function": "value2"},
		},
		"Sucess with empty OuputKeys": {
			client: func(t *testing.T) DescribeStacksAPI {
				return mockDescribeStacksAPI(func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					t.Helper()
					if params.StackName == nil {
						t.Fatal("expect StackName to not be nil")
					}
					if e, a := "StackName", *params.StackName; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}

					return &cloudformation.DescribeStacksOutput{
						NextToken: nil,
						Stacks: []types.Stack{
							{
								StackName: params.StackName,
								Outputs: []types.Output{
									{
										OutputKey:   aws.String("Queue"),
										OutputValue: aws.String("value1"),
									},
									{
										OutputKey:   aws.String("Function"),
										OutputValue: aws.String("value2"),
									},
								},
							},
						},
					}, nil
				})
			},
			stackName:  "StackName",
			outputKeys: []string{},
			expect:     map[string]string{},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			r, _ := GetStackOuput(tt.stackName, tt.outputKeys, tt.client(t))

			assert.Equal(t, tt.expect, r)
		})
	}
}

func TestErrGetStackOutput(t *testing.T) {
	cases := map[string]struct {
		client     func(t *testing.T) DescribeStacksAPI
		stackName  string
		outputKeys []string
		expect     string
	}{
		"Err: Not all keys found": {
			client: func(t *testing.T) DescribeStacksAPI {
				return mockDescribeStacksAPI(func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					t.Helper()
					if params.StackName == nil {
						t.Fatal("expect StackName to not be nil")
					}
					if e, a := "StackName", *params.StackName; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}

					return &cloudformation.DescribeStacksOutput{
						NextToken: nil,
						Stacks: []types.Stack{
							{
								StackName: params.StackName,
								Outputs: []types.Output{
									{
										OutputKey:   aws.String("Queue"),
										OutputValue: aws.String("value1"),
									},
								},
							},
						},
					}, nil
				})
			},
			stackName:  "StackName",
			outputKeys: []string{"Queue", "Function"},
			expect:     "Not all output keys found",
		},
		"Err: DescribeStacks response with err": {
			client: func(t *testing.T) DescribeStacksAPI {
				return mockDescribeStacksAPI(func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					t.Helper()
					if params.StackName == nil {
						t.Fatal("expect StackName to not be nil")
					}
					if e, a := "StackName", *params.StackName; e != a {
						t.Errorf("expect %v, got %v", e, a)
					}

					return nil, errors.New("API Error")
				})
			},
			stackName:  "StackName",
			outputKeys: []string{"Queue"},
			expect:     "API Error",
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := GetStackOuput(tt.stackName, tt.outputKeys, tt.client(t))

			assert.Equal(t, tt.expect, err.Error())
		})
	}
}
