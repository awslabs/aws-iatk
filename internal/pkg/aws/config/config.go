// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"log"
	"os"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/config"
	smithymiddleware "github.com/aws/smithy-go/middleware"
)

func GetAWSConfig(ctx context.Context, region string, profile string, metadata *jsonrpc.Metadata) (aws.Config, error) {
	r := os.Getenv("AWS_REGION")
	log.Printf("Region: %q", r)
	uaVal := "unknown"
	if metadata != nil {
		uaVal = metadata.UserAgentValue()
	}
	args := [](func(*config.LoadOptions) error){
		config.WithAPIOptions([]func(*smithymiddleware.Stack) error{
			awsmiddleware.AddUserAgentKeyValue("aws-zion", uaVal),
		}),
	}
	if profile != "" {
		args = append(args, config.WithSharedConfigProfile(profile))
	}
	if region != "" {
		args = append(args, config.WithRegion(region))
	}
	cfg, err := config.LoadDefaultConfig(
		ctx,
		args...,
	)

	return cfg, err
}
