// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"zion/internal/pkg/aws/config"
	"zion/internal/pkg/jsonrpc"
	mockevent "zion/internal/pkg/mock/event"
	"zion/internal/pkg/public-rpc/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
)

type GenerateBareboneEventsParams struct {
	RegistryName  string
	SchemaName    string
	SchemaVersion string

	EventRef     string
	SkipOptional bool

	Profile string
	Region  string
}

func (p *GenerateBareboneEventsParams) RPCMethod(metadata *jsonrpc.Metadata) (*types.Result, error) {
	err := p.validateParams()
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	cfg, err := config.GetAWSConfig(ctx, p.Region, p.Profile, metadata)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %w", err)
	}

	var schema *mockevent.Schema
	client := schemas.NewFromConfig(cfg)
	schema, err = mockevent.NewSchemaFromRegistry(
		ctx,
		aws.String(p.RegistryName),
		aws.String(p.SchemaName),
		aws.String(p.SchemaVersion),
		aws.String(p.EventRef),
		client,
	)

	if err != nil {
		return nil, fmt.Errorf("error reading schema: %w", err)
	}

	event, err := mockevent.GenerateMockEvent(schema, p.SkipOptional)
	if err != nil {
		return nil, fmt.Errorf("error generating mock event: %w", err)
	}

	return &types.Result{
		Output: event,
	}, nil
}

func (p *GenerateBareboneEventsParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(mockevent.GenerateMockEvent)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}

func (p *GenerateBareboneEventsParams) validateParams() error {
	if p.RegistryName == "" || p.SchemaName == "" {
		return errors.New(`requires both "RegistryName" and "SchemaName"`)
	}
	return nil
}
