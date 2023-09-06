package schemaregistry

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/schemas"
)

func GetRegistrySchema(ctx context.Context, registryName, schemaName, schemaVersion, eventRef *string, api DescribeSchemaAPI) (*Schema, error) {

	input := &schemas.DescribeSchemaInput{
		RegistryName:  registryName,
		SchemaName:    schemaName,
		SchemaVersion: schemaVersion,
	}
	describeSchemaOutput, err := api.DescribeSchema(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("failed to get the schema %v: %w", aws.ToString(schemaName), err)
	}

	return &Schema{
		EventRef:      eventRef,
		SchemaContent: describeSchemaOutput.Content,
		SchemaType:    describeSchemaOutput.Type,
	}, nil
}

//go:generate mockery --name DescribeSchemaAPI
type DescribeSchemaAPI interface {
	DescribeSchema(ctx context.Context, params *schemas.DescribeSchemaInput, optFns ...func(*schemas.Options)) (*schemas.DescribeSchemaOutput, error)
}
