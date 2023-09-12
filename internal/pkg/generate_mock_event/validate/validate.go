package validate

import (
	"encoding/json"
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v4"
	"log"
	"strings"
)

func ValidateEvent(schemaString string, event []byte) (bool, error) {
	var eventTest interface{}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("temp.json", strings.NewReader(schemaString)); err != nil {
		log.Fatal(err)
	}
	schema, err := compiler.Compile("temp.json")
	if err != nil {
		return false, fmt.Errorf("error compiling schema: %w", err)
	}
	_ = json.Unmarshal(event, &eventTest)
	err = schema.ValidateInterface(eventTest)
	if err != nil {
		return false, fmt.Errorf("validation error: %w", err)
	}
	return true, nil
}
