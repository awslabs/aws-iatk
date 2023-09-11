package event

import (
	"encoding/json"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v4"
	"golang.org/x/exp/slices"
	"log"
	"strings"
	"time"
)

type Map map[string]interface{}

func GenerateEventObject(schema *jsonschema.Schema, generateRequiredOnly bool, maxDepth int) (Map, error) {
	event := Map{}
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
					event[key], _ = GenerateEventObject(schema.Properties[key], generateRequiredOnly, maxDepth-1)
				}
			case "array":
				event[key] = []string{}
			default:
				return nil, fmt.Errorf(
					"invalid or unsupported property type %q found for property %q", elementType, key,
				)

			}
		} else if len(elementType) > 1 {
			return nil, fmt.Errorf("cannot handle mmultiple type declaration")
		}
		if schema.Properties[key].Ref != nil && maxDepth > 0 {
			event[key], _ = GenerateEventObject(schema.Properties[key].Ref, generateRequiredOnly, maxDepth-1)
		} else if schema.Properties[key].Ref != nil {
			event[key] = nil
		}
	}
	return event, nil
}

func GenerateEvent(schemaString string, generateRequiredOnly bool) ([]byte, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("temp.json", strings.NewReader(schemaString)); err != nil {
		log.Fatal(err)
	}
	schema, err := compiler.Compile("temp.json")
	if err != nil {
		return nil, fmt.Errorf("error compiling schema: %w", err)
	}

	eventMap, err := GenerateEventObject(schema, generateRequiredOnly, 3)
	if err != nil {
		return nil, fmt.Errorf("error generating event: %w", err)
	}
	return json.Marshal(eventMap)
}
