package event

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"golang.org/x/exp/slices"
)

// Function to load openapi schema and generate a mock event
func GenerateOpenapiEvent(schema *Schema, skipOptional bool) ([]byte, error) {

	if *schema.EventRef == "" {
		return nil, fmt.Errorf("no eventRef specified to generate a mock event")
	}

	var schemaByteArray []byte
	if schema.SchemaContent == nil {
		return nil, fmt.Errorf("failed while loading schema due to error: invalid schema provided")
	}

	schemaByteArray = []byte(*schema.SchemaContent)

	openApiSchema, err := openapi3.NewLoader().LoadFromData(schemaByteArray)
	if err != nil {
		return nil, fmt.Errorf("failed while loading schema due to error: %v", err)
	}

	schemaComponents := openApiSchema.Components
	if schemaComponents == nil {
		return nil, fmt.Errorf("failed to generate a mock event, no components found in schema")
	}
	schemaComponentsSchemas := schemaComponents.Schemas
	if schemaComponentsSchemas == nil {
		return nil, fmt.Errorf("failed to generate a mock event, no schemas found under components in schema")
	}

	eventComponentRef, ok := schemaComponentsSchemas[*schema.EventRef]
	if !ok {
		return nil, fmt.Errorf("provided eventRef %q not found in the schema", *schema.EventRef)
	}

	eventMap, err := ConstructOpenapiEvent(eventComponentRef.Value, skipOptional, 20)
	if err != nil {
		return nil, fmt.Errorf("failed to generate a mock event for the provided schema due to error: %v", err)
	}

	return json.Marshal(eventMap)
}

// Recursive function to iterate the properties and construct the event
func ConstructOpenapiEvent(schema *openapi3.Schema, skipOptional bool, maxDepth int) (EventMap, error) {
	event := EventMap{}
	for propertyName, propertyRef := range schema.Properties {
		if schema.Required != nil && skipOptional && !slices.Contains(schema.Required, propertyName) {
			continue
		}
		propertyVal := propertyRef.Value
		if len(propertyVal.Enum) > 0 {
			event[propertyName] = propertyVal.Enum[0]
			continue
		}
		switch propertyVal.Type {
		case "string":
			if propertyVal.Format == "date-time" {
				event[propertyName] = time.Now()
			} else {
				event[propertyName] = ""
			}
		case "number", "integer":
			event[propertyName] = 0
		case "boolean":
			event[propertyName] = false
		case "array":
			event[propertyName] = []string{}
		case "object":
			var err error
			if maxDepth == 0 {
				event[propertyName] = nil
			} else {
				event[propertyName], err = ConstructOpenapiEvent(propertyVal, skipOptional, maxDepth-1)
				if err != nil {
					return nil, err
				}
			}
		default:
			return nil, fmt.Errorf(
				"invalid or unsupported property type %q found for property %q", propertyVal.Type, propertyName,
			)
		}
	}

	return event, nil
}
