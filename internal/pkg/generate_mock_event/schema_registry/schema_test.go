package schemaregistry

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/schemas"
	"github.com/stretchr/testify/assert"
)

func TestNewSchemaFromRegistry(t *testing.T) {
	cases := map[string]struct {
		mockDescribeSchemaAPI func(ctx context.Context, registryName, schemaName, schemaVersion, eventRef *string, schemaType, schemaContent string) *MockDescribeSchemaAPI
		registryName          string
		schemaName            string
		schemaVersion         string
		eventRef              string
		schemaType            string
		schemaContent         string
		expectErr             error
	}{
		"success": {
			registryName:  "mock-registry-name",
			schemaName:    "mock-schema-name",
			schemaVersion: "1",
			schemaType:    "JSONSchemaDraft4",
			schemaContent: "{\n  \"$id\": \"https://example.com/person.schema.json\",\n  \"$schema\": \"https://json-schema.org/draft/2020-12/schema\",\n  \"title\": \"Person\",\n  \"type\": \"object\",\n  \"properties\": {\n    \"firstName\": {\n      \"type\": \"string\",\n      \"description\": \"The person's first name.\"\n    },\n    \"lastName\": {\n      \"type\": \"string\",\n      \"description\": \"The person's last name.\"\n    },\n    \"age\": {\n      \"description\": \"Age in years which must be equal to or greater than zero.\",\n      \"type\": \"integer\",\n      \"minimum\": 0\n    }\n  }\n}",
			eventRef:      "#/definitions/Person",
			expectErr:     nil,
			mockDescribeSchemaAPI: func(ctx context.Context, registryName, schemaName, schemaVersion, eventRef *string, schemaType, schemaContent string) *MockDescribeSchemaAPI {
				api := NewMockDescribeSchemaAPI(t)
				api.EXPECT().
					DescribeSchema(ctx, &schemas.DescribeSchemaInput{
						RegistryName:  registryName,
						SchemaName:    schemaName,
						SchemaVersion: schemaVersion,
					}).
					Return(&schemas.DescribeSchemaOutput{
						Content:       &schemaContent,
						SchemaName:    schemaName,
						SchemaVersion: schemaVersion,
						Type:          &schemaType,
					}, nil)
				return api
			},
		},
		"success when no schemaVersion provided": {
			registryName:  "mock-registry-name",
			schemaName:    "mock-schema-name",
			schemaType:    "OpenApi3",
			schemaContent: "{\"openapi\":\"3.0.0\",\"info\":{\"version\":\"1.0.0\",\"title\":\"SomeAwesomeSchema\"},\"paths\":{},\"components\":{\"schemas\":{\"Some Awesome Schema\":{\"type\":\"object\",\"required\":[\"foo\",\"bar\",\"baz\"],\"properties\":{\"foo\":{\"type\":\"string\"},\"bar\":{\"type\":\"string\"},\"baz\":{\"type\":\"string\"}}}}}} ",
			eventRef:      "#/components/schemas/Person",
			expectErr:     nil,
			mockDescribeSchemaAPI: func(ctx context.Context, registryName, schemaName, schemaVersion, eventRef *string, schemaType, schemaContent string) *MockDescribeSchemaAPI {
				api := NewMockDescribeSchemaAPI(t)
				api.EXPECT().
					DescribeSchema(ctx, &schemas.DescribeSchemaInput{
						RegistryName:  registryName,
						SchemaName:    schemaName,
						SchemaVersion: nil,
					}).
					Return(&schemas.DescribeSchemaOutput{
						Content:    &schemaContent,
						SchemaName: schemaName,
						Type:       &schemaType,
					}, nil)
				return api
			},
		},
		"describe schema api failed": {
			registryName:  "mock-registry-name",
			schemaName:    "mock-schema-name",
			schemaVersion: "1",
			eventRef:      "",
			mockDescribeSchemaAPI: func(ctx context.Context, registryName, schemaName, schemaVersion, eventRef *string, schemaType, schemaContent string) *MockDescribeSchemaAPI {
				api := NewMockDescribeSchemaAPI(t)
				api.EXPECT().
					DescribeSchema(ctx, &schemas.DescribeSchemaInput{
						RegistryName:  registryName,
						SchemaName:    schemaName,
						SchemaVersion: schemaVersion,
					}).
					Return(nil, errors.New("schema not found"))
				return api
			},
			expectErr: errors.New("failed to get the schema mock-schema-name: schema not found"),
		},
	}
	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()

			var schemaVersion *string
			if tt.schemaVersion == "" {
				schemaVersion = nil
			} else {
				schemaVersion = &tt.schemaVersion
			}
			outputSchema, err := NewSchemaFromRegistry(
				ctx, &tt.registryName, &tt.schemaName, schemaVersion, &tt.eventRef,
				tt.mockDescribeSchemaAPI(
					ctx, &tt.registryName, &tt.schemaName, schemaVersion, &tt.eventRef, tt.schemaType, tt.schemaContent,
				),
			)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, tt.schemaType, *outputSchema.SchemaType)
				assert.Equal(t, tt.eventRef, *outputSchema.EventRef)
				assert.Equal(t, tt.schemaContent, *outputSchema.SchemaContent)
			}
		})
	}
}
