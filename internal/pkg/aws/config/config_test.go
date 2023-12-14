// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"iatk/internal/pkg/jsonrpc"
	"reflect"
	"runtime"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigRegion(t *testing.T) {
	cfg, err := GetAWSConfig(context.TODO(), "us-west-2", "", nil)

	if err != nil {
		t.Fail()
	}

	assert.Equal(t, "us-west-2", cfg.Region)
}

func TestConfigClientLogMode(t *testing.T) {
	cfg, err := GetAWSConfig(context.TODO(), "us-west-2", "", nil)

	if err != nil {
		t.Fail()
	}

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
	expectFuncName := runtime.FuncForPC(reflect.ValueOf(expectFunc).Pointer())
	actualFuncName := runtime.FuncForPC(reflect.ValueOf(cfg.APIOptions[0]).Pointer())
	assert.Equal(t, expectFuncName, actualFuncName)
}
