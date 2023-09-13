// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"zion/internal/pkg/aws/config"
	mockevent "zion/internal/pkg/mock/event"
	"zion/internal/pkg/public-rpc/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
	"golang.org/x/exp/slices"
)

type GenerateMockEventsParams struct {
	RegistryName  string
	SchemaName    string
	SchemaVersion string

	EventRef     string
	Context      []string
	Overrides    string
	SkipOptional bool

	Profile string
	Region  string
}

func (p *GenerateMockEventsParams) RPCMethod() (*types.Result, error) {
	err := p.validateParams()
	if err != nil {
		return nil, err
	}

	ctx := context.TODO()
	cfg, err := config.GetAWSConfig(ctx, p.Region, p.Profile)
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

	event, err := GenerateMockEvent(schema, p.SkipOptional)
	if err != nil {
		return nil, fmt.Errorf("error generating mock event: %w", err)
	}

	return &types.Result{
		Output: event,
	}, nil
}

func (p *GenerateMockEventsParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(GenerateMockEvent)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}

func (p *GenerateMockEventsParams) validateParams() error {
	if p.RegistryName == "" || p.SchemaName == "" {
		return errors.New(`requires both "RegistryName" and "SchemaName"`)
	}

	supportedContext := []string{
		"eventbridge.v0",
	}
	if p.Context != nil && len(p.Context) > 0 {
		for _, c := range p.Context {
			if !slices.Contains(supportedContext, c) {
				return fmt.Errorf("%q is not a supported context. supported context: %v", c, supportedContext)
			}
		}
	}

	return nil
}

// TODO: to be replaced by actual implementation (in internal pkg)
func GenerateMockEvent(schema *mockevent.Schema, skipOptional bool) (string, error) {

	// Commenting out as the integ tests fail, will be moved to internal pkg in a separate PR
	// generatedEvent, err := mockevent.GenerateOpenapiEvent(schema, false)

	// if err != nil {
	// 	return "", fmt.Errorf("error generating mock event: %w", err)
	// }

	// return string(generatedEvent), nil

	return "", nil
}
