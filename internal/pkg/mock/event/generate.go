package event

import (
	"fmt"
)

func GenerateMockEvent(schema *Schema, skipOptional bool) (string, error) {

	var generatedEvent []byte
	var err error
	if *schema.SchemaType == "OpenApi3" {
		generatedEvent, err = GenerateOpenapiEvent(schema, skipOptional)
	} else if *schema.SchemaType == "JSONSchemaDraft4" {
		generatedEvent, err = GenerateJsonschemaEvent(schema, skipOptional)
	} else {
		return "", fmt.Errorf("error generating mock event: unsupported schema type found")
	}

	if err != nil {
		return "", fmt.Errorf("error generating mock event: %w", err)
	}

	return string(generatedEvent), nil
}
