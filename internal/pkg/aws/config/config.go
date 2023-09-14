// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	smithymiddleware "github.com/aws/smithy-go/middleware"
)

func GetAWSConfig(ctx context.Context, region string, profile string, metadata *jsonrpc.Metadata) (aws.Config, error) {
	uaVal := "unknown"
	if metadata != nil {
		uaVal = metadata.UserAgentValue()
	}
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequest|aws.LogResponse|aws.LogRequest),
		config.WithAPIOptions([]func(*smithymiddleware.Stack) error{
			awsmiddleware.AddUserAgentKeyValue("aws-zion", uaVal),
		}),
	)

	return cfg, err
}
