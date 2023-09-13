package event

import (
	"encoding/json"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v4"
	"golang.org/x/exp/slices"
	"time"
)

func ConstructJsonschemaEvent(schema *jsonschema.Schema, generateRequiredOnly bool, maxDepth int) (EventMap, error) {
	event := EventMap{}
	for key, element := range schema.Properties {
		if schema.Required != nil && generateRequiredOnly && !slices.Contains(schema.Required, key) {
			continue
		}
		elementType := element.Types
		if len(elementType) == 1 {
			if len(element.Enum) > 0 {
				event[key] = element.Enum[0]
				continue
			}
			switch elementType[0] {
			case "string":
				if element.Format == "date-time" {
					event[key] = time.Now()
				} else {
					event[key] = ""
				}
			case "number", "integer":
				event[key] = 0
			case "null":
				event[key] = nil
			case "bool":
				event[key] = false
			case "object":
				if maxDepth == 0 {
					event[key] = nil
				} else {
					var err error
					event[key], err = ConstructJsonschemaEvent(schema.Properties[key], generateRequiredOnly, maxDepth-1)
					if err != nil {
						return nil, fmt.Errorf("error generated event object: %w", err)
					}
				}
			case "array":
				event[key] = []string{}
			default:
				return nil, fmt.Errorf(
					"invalid or unsupported property type %q found for property %q", elementType, key,
				)

			}
		} else if len(elementType) > 1 {
			return nil, fmt.Errorf("cannot handle multiple type declaration")
		}
		if schema.Properties[key].Ref != nil && maxDepth > 0 {
			event[key], _ = ConstructJsonschemaEvent(schema.Properties[key].Ref, generateRequiredOnly, maxDepth-1)
		} else if schema.Properties[key].Ref != nil {
			event[key] = nil
		}
	}
	return event, nil
}

func GenerateJsonschemaEvent(rawSchema *Schema, skipOptional bool) ([]byte, error) {
	if len(*rawSchema.SchemaContent) < 1 {
		return nil, fmt.Errorf("failed while loading schema due to error: invalid schema provided")
	}
	if len(*rawSchema.EventRef) > 0 {
		return nil, fmt.Errorf("failed while loading schema due to error: event ref not supported for json schemas")
	}
	schema, err := jsonschema.CompileString("temp.json", *rawSchema.SchemaContent)
	if err != nil {
		return nil, fmt.Errorf("error compiling schema: %w", err)
	}

	eventMap, err := ConstructJsonschemaEvent(schema, skipOptional, 20)
	if err != nil {
		return nil, fmt.Errorf("error generating event: %w", err)
	}
	return json.Marshal(eventMap)
}
