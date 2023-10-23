// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"reflect"
	"runtime"
	"testing"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigRegion(t *testing.T) {
	cases := []struct {
		name   string
		region string
		env    string
	}{
		{
			name:   "Region provided",
			region: "us-west-2",
			env:    "us-east-1",
		},
		{
			name:   "Region not provided",
			region: "",
			env:    "us-east-1",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("AWS_REGION", tt.env)
			cfg, err := GetAWSConfig(context.TODO(), tt.region, "", nil)

			assert.NoError(t, err)

			if tt.region != "" {
				assert.Equal(t, tt.region, cfg.Region)
			} else {
				assert.Equal(t, tt.env, cfg.Region)
			}
			os.Unsetenv("AWS_REGION")
		})
	}
}

func TestConfigClientLogMode(t *testing.T) {
	cfg, err := GetAWSConfig(context.TODO(), "us-west-2", "", nil)

	assert.NoError(t, err)

	// should not log anything from AWS SDK
	var logMode aws.ClientLogMode
	assert.Equal(t, logMode, cfg.ClientLogMode)
}

func TestConfigClientUserAgentIsAdded(t *testing.T) {
	metadata := &jsonrpc.Metadata{
		Client:  "python",
		Version: "0.0.3",
		Caller:  "wait_until_event_matched",
	}
	cfg, err := GetAWSConfig(context.TODO(), "us-west-2", "", metadata)
	require.NoError(t, err)
	assert.Len(t, cfg.APIOptions, 1)
	expectFunc := awsmiddleware.AddUserAgentKeyValue("does not matter", "does not matter")
	expectFuncName := runtime.FuncForPC(reflect.ValueOf(expectFunc).Pointer()).Name()
	actualFuncName := runtime.FuncForPC(reflect.ValueOf(cfg.APIOptions[0]).Pointer()).Name()
	assert.Equal(t, actualFuncName, expectFuncName)
}
