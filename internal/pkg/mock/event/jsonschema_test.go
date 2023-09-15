package event

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessfulStringSchema(t *testing.T) {
	var eventMap EventMap
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#",
		"properties": { 
			"test": { "type": "string" }}}`
	schemaType := "JSONSchemaDraft4"
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType}
	event, _ := GenerateJsonschemaEvent(&schema, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, "", eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulEnumSchema(t *testing.T) {
	var eventMap EventMap
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#",
		"properties": { 
			"test": { "type": "string",
  					  "enum": ["testEnum"]}}}`
	schemaType := "JSONSchemaDraft4"
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType}
	event, _ := GenerateJsonschemaEvent(&schema, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, "", eventMap["test"])
	assert.Equal(t, "testEnum", eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestCompilationError(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "random"}}}`
	schemaType := "JSONSchemaDraft4"
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType}
	_, err := GenerateJsonschemaEvent(&schema, false)
	assert.Error(t, errors.New(`error compiling schema: json-schema \"temp.json\" compilation failed)`), err)
}

func TestSuccessfulNumberSchema(t *testing.T) {
	var eventMap EventMap
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "number"}}}`
	schemaType := "JSONSchemaDraft4"
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType}
	event, _ := GenerateJsonschemaEvent(&schema, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, float64(1), eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulArraySchema(t *testing.T) {
	var eventMap EventMap
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": {
			"test": {
				"items": {
				  "type": "string"
				},
				"type": "array"
			}}}`
	schemaType := "JSONSchemaDraft4"
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType}
	event, _ := GenerateJsonschemaEvent(&schema, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, []interface{}{}, eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulObjectSchema(t *testing.T) {
	var eventMap EventMap
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": {
			"testObject": {
				"properties": {
				  "test": {
           			 "type": "string"
					}},
				"type": "object"
			}}}`
	schemaType := "JSONSchemaDraft4"
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType}
	event, _ := GenerateJsonschemaEvent(&schema, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.NotEmpty(t, eventMap["testObject"])
	assert.IsType(t, map[string]any{}, eventMap["testObject"])
	objMap := eventMap["testObject"].(map[string]interface{})
	assert.Contains(t, objMap, "test")
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulRequiredOnlySchema(t *testing.T) {
	var eventMap EventMap
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "number"},
			"notRequired": {"type": "string"}},
			"required": ["test"]}`
	schemaType := "JSONSchemaDraft4"
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType}
	event, _ := GenerateJsonschemaEvent(&schema, true)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.Empty(t, eventMap["notRequired"])
	assert.IsType(t, float64(1), eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}
