// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	smithymiddleware "github.com/aws/smithy-go/middleware"
)

func GetAWSConfig(ctx context.Context, region string, profile string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequest|aws.LogResponse|aws.LogRequest),
		config.WithAPIOptions([]func(*smithymiddleware.Stack) error{
			// (@jfuss) Feels like we may want to know what client language the customer is using. Making
			// A note for later
			awsmiddleware.AddUserAgentKeyValue("aws-zion", "0.1"),
		}),
	)

	return cfg, err
}
