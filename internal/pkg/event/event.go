package event

import (
	"encoding/json"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v4"
	"golang.org/x/exp/slices"
	"reflect"
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
func GenerateRefSchema(schema *jsonschema.Schema, eventRef string, maxDepth int) (jsonschema.Schema, error) {
	queueSchema := []jsonschema.Schema{*schema}
	queueMaxDepth := []int{maxDepth}
	for len(queueSchema) > 0 {
		curSchema := queueSchema[len(queueSchema)-1]
		curMaxDepth := queueMaxDepth[len(queueMaxDepth)-1]
		queueSchema = queueSchema[:len(queueSchema)-1]
		queueMaxDepth = queueMaxDepth[:len(queueMaxDepth)-1]
		if curSchema.Ref != nil && curSchema.Ref.Ptr == eventRef {
			return *curSchema.Ref, nil
		}
		if curMaxDepth > 0 {
			for _, element := range curSchema.Properties {
				queueSchema = append(queueSchema, *element)
				queueMaxDepth = append(queueMaxDepth, curMaxDepth-1)
			}
			if reflect.TypeOf(curSchema.Items) == reflect.TypeOf([]*jsonschema.Schema{}) {
				items := curSchema.Items.([]*jsonschema.Schema)
				for _, element := range items {
					queueSchema = append(queueSchema, *element)
					queueMaxDepth = append(queueMaxDepth, curMaxDepth-1)
				}
			} else if reflect.TypeOf(curSchema.Items) == reflect.TypeOf(schema) {
				item := curSchema.Items.(*jsonschema.Schema)
				queueSchema = append(queueSchema, *item)
				queueMaxDepth = append(queueMaxDepth, curMaxDepth-1)
			}
			if schema.Ref != nil {
				queueSchema = append(queueSchema, *schema.Ref)
				queueMaxDepth = append(queueMaxDepth, curMaxDepth-1)
			}
		}
	}
	return jsonschema.Schema{}, fmt.Errorf("reference not found")
}

func GenerateEvent(schemaString string, generateRequiredOnly bool, eventRef string) ([]byte, error) {
	schema, err := jsonschema.CompileString("temp.json", schemaString)
	if len(eventRef) > 1 {
		*schema, err = GenerateRefSchema(schema, eventRef, 20)
	}
	if err != nil {
		return nil, fmt.Errorf("error compiling schema: %w", err)
	}

	eventMap, err := GenerateEventObject(schema, generateRequiredOnly, 20)
	fmt.Println(eventMap)
	if err != nil {
		return nil, fmt.Errorf("error generating event: %w", err)
	}
	return json.Marshal(eventMap)
}
