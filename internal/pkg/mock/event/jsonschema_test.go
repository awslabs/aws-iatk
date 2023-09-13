package event

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSuccessfulStringSchema(t *testing.T) {
	var eventMap EventMap
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#",
		"properties": { 
			"test": { "type": "string" }}}`
	eventRef := ""
	schemaType := SchemaType("JSONSchemaDraft4")
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType, EventRef: &eventRef}
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
	eventRef := ""
	schemaType := SchemaType("JSONSchemaDraft4")
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType, EventRef: &eventRef}
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
	eventRef := ""
	schemaType := SchemaType("JSONSchemaDraft4")
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType, EventRef: &eventRef}
	_, err := GenerateJsonschemaEvent(&schema, false)
	assert.Error(t, errors.New(`error compiling schema: json-schema \"temp.json\" compilation failed)`), err)
}

func TestSuccessfulNumberSchema(t *testing.T) {
	var eventMap EventMap
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "number"}}}`
	eventRef := ""
	schemaType := SchemaType("JSONSchemaDraft4")
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType, EventRef: &eventRef}
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
	eventRef := ""
	schemaType := SchemaType("JSONSchemaDraft4")
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType, EventRef: &eventRef}
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
	eventRef := ""
	schemaType := SchemaType("JSONSchemaDraft4")
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType, EventRef: &eventRef}
	event, _ := GenerateEvent(&schema, false)
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
	eventRef := ""
	schemaType := SchemaType("JSONSchemaDraft4")
	schema := Schema{SchemaContent: &schemaString, SchemaType: &schemaType, EventRef: &eventRef}
	event, _ := GenerateEvent(&schema, true)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.Empty(t, eventMap["notRequired"])
	assert.IsType(t, float64(1), eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}
