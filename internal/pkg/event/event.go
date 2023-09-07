package event

import "C"
import (
	"encoding/json"
	"errors"
	"github.com/santhosh-tekuri/jsonschema/v4"
	"log"
	"reflect"
	"strings"
)

type Map map[string]interface{}

func GenerateEvent(schema *jsonschema.Schema) (Map, error) {
	event := Map{}
	for key, element := range schema.Properties {
		elementType := element.Types
		if len(elementType) == 1 {
			switch elementType[0] {
			case "string":
				event[key] = "string"
			case "number":
				event[key] = 1
			case "object":
				event[key], _ = GenerateEvent(schema.Properties[key])
			case "array":
				arr := make([]any, 1)
				if reflect.TypeOf(schema.Properties[key].Items) == reflect.TypeOf(schema) {
					arrSchema := schema.Properties[key].Items.(*jsonschema.Schema)
					if len(arrSchema.Types) == 1 {
						switch arrSchema.Types[0] {
						case "string":
							arr[0] = "string"
						case "number":
							arr[0] = 1
						case "object":
							arr[0], _ = GenerateEvent(schema.Properties[key])
						}
					}
					event[key] = arr
				}
			}
		}
		if schema.Properties[key].Ref != nil {
			event[key], _ = GenerateEvent(schema.Properties[key].Ref)
		}
	}
	return event, nil
}

func GenerateEventString(schemaString string) ([]byte, error) {
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("temp.json", strings.NewReader(schemaString)); err != nil {
		log.Fatal(err)
	}
	schema, err := compiler.Compile("temp.json")
	if err != nil {
		return nil, errors.New("error compiling schema: " + err.Error())
	}

	eventMap, _ := GenerateEvent(schema)
	return json.Marshal(eventMap)
}
