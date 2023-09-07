// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"zion/internal/pkg/aws/config"
	schemaregistry "zion/internal/pkg/generate_mock_event/schema_registry"
	"zion/internal/pkg/public-rpc/types"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
)

type GenerateMockEventsParams struct {
	RegistryName  string
	SchemaName    string
	SchemaVersion string

	SchemaFile string

	EventRef     string
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

	var schema *schemaregistry.Schema
	if p.RegistryName != "" {
		client := schemas.NewFromConfig(cfg)
		schema, err = schemaregistry.NewSchemaFromRegistry(
			ctx,
			aws.String(p.RegistryName),
			aws.String(p.SchemaName),
			aws.String(p.SchemaVersion),
			aws.String(p.EventRef),
			client,
		)
	} else {
		schema, err = NewSchemaFromLocalFile()
	}
	if err != nil {
		return nil, fmt.Errorf("error reading schema: %w", err)
	}

	event, err := GenerateMockEvent(schema)
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
	if p.RegistryName == "" && p.SchemaName == "" && p.SchemaFile == "" {
		return errors.New(`missing either "RegistryName and SchemaName" or "SchemaFile"`)
	}

	if p.SchemaFile != "" && (p.SchemaName != "" || p.RegistryName != "") {
		return errors.New(`provide either "RegistryName and SchemaName" or "SchemaFile", not both`)
	}
	if p.SchemaFile == "" && ((p.RegistryName == "" && p.SchemaName != "") || (p.RegistryName != "" && p.SchemaName == "")) {
		return errors.New(`requires both "RegistryName" and "SchemaName"`)
	}

	return nil
}

// TODO: to be replaced by actual implementation (in internal pkg)
func NewSchemaFromLocalFile() (*schemaregistry.Schema, error) {
	return &schemaregistry.Schema{}, nil
}

// TODO: to be replaced by actual implementation (in internal pkg)
func GenerateMockEvent(schema *schemaregistry.Schema) (string, error) {
	return "", nil
}
