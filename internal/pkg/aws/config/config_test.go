// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestConfigRegion(t *testing.T) {
	cfg, err := GetAWSConfig(context.TODO(), "us-west-2", "")

	if err != nil {
		t.Fail()
	}

	assert.Equal(t, "us-west-2", cfg.Region)
}

func TestConfigClientLogMode(t *testing.T) {
	cfg, err := GetAWSConfig(context.TODO(), "us-west-2", "")

	if err != nil {
		t.Fail()
	}

	assert.Equal(t, aws.LogRetries|aws.LogRequest|aws.LogResponse|aws.LogRequest, cfg.ClientLogMode)
}
