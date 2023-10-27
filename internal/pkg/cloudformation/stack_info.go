// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cloudformation

import (
	"context"
	"errors"

	"ctk/internal/pkg/slice"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"golang.org/x/exp/slices"
)

type DescribeStackResourceAPI interface {
	DescribeStackResource(ctx context.Context, params *cloudformation.DescribeStackResourceInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStackResourceOutput, error)
}

type DescribeStacksAPI interface {
	DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
}

func GetPhysicalId(stackName string, logicalID string, api DescribeStackResourceAPI) (string, error) {
	output, err := api.DescribeStackResource(context.TODO(), &cloudformation.DescribeStackResourceInput{
		LogicalResourceId: aws.String(logicalID),
		StackName:         aws.String(stackName),
	})

	if err != nil {
		return "", err
	}

	return aws.ToString(output.StackResourceDetail.PhysicalResourceId), nil
}

func GetStackOuput(stackName string, outputKeys []string, api DescribeStacksAPI) (map[string]string, error) {
	distinct := slice.Dedup(outputKeys)
	p := cloudformation.NewDescribeStacksPaginator(api, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})

	m := make(map[string]string)

	for p.HasMorePages() {
		resp, err := p.NextPage(context.TODO())

		if err != nil {
			return nil, err
		}

		for _, stack := range resp.Stacks {
			if aws.ToString(stack.StackName) != stackName {
				break
			}

			for _, sInfo := range stack.Outputs {
				if slices.Contains(distinct, *sInfo.OutputKey) {
					m[aws.ToString(sInfo.OutputKey)] = aws.ToString(sInfo.OutputValue)
				}
			}
		}
	}

	if len(m) != len(distinct) {
		return nil, errors.New("Not all output keys found")
	}

	return m, nil
}
