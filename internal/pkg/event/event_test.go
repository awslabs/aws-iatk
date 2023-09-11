package event

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessfulStringSchema(t *testing.T) {
	var eventMap Map
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#",
		"properties": { 
			"test": { "type": "string" }}}`
	event, _ := GenerateEvent(schemaString, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, "", eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulEnumSchema(t *testing.T) {
	var eventMap Map
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#",
		"properties": { 
			"test": { "type": "string",
  					  "enum": ["testEnum"]}}}`
	event, _ := GenerateEvent(schemaString, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, "", eventMap["test"])
	assert.Equal(t, "testEnum", eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulRefSchema(t *testing.T) {
	var eventMap Map
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"definitions": { 
			"RoomStateChange": {
				"properties": { 
					"testRef": { "type": "string" }}}}, 
		"properties": { "test": { "type": "string" }, 
		"detail": { "$ref": "#/definitions/RoomStateChange" 
		}}}`
	event, _ := GenerateEvent(schemaString, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "detail")
	refMap := eventMap["detail"].(map[string]interface{})
	assert.Contains(t, refMap, "testRef")
	assert.IsType(t, eventMap["test"], "")
	assert.Equal(t, 2, len(eventMap))
}

func TestCompilationError(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "random"}}}`
	_, err := GenerateEvent(schemaString, false)
	assert.Error(t, errors.New(`error compiling schema: json-schema \"temp.json\" compilation failed)`), err)
}

func TestSuccessfulNumberSchema(t *testing.T) {
	var eventMap Map
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "number"}}}`
	event, _ := GenerateEvent(schemaString, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, float64(1), eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulArraySchema(t *testing.T) {
	var eventMap Map
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": {
			"test": {
				"items": {
				  "type": "string"
				},
				"type": "array"
			}}}`
	event, _ := GenerateEvent(schemaString, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, []interface{}{}, eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulObjectSchema(t *testing.T) {
	var eventMap Map
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": {
			"testObject": {
				"properties": {
				  "test": {
           			 "type": "string"
					}},
				"type": "object"
			}}}`
	event, _ := GenerateEvent(schemaString, false)
	_ = json.Unmarshal(event, &eventMap)
	assert.NotEmpty(t, eventMap["testObject"])
	assert.IsType(t, map[string]any{}, eventMap["testObject"])
	objMap := eventMap["testObject"].(map[string]interface{})
	assert.Contains(t, objMap, "test")
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfulRequiredOnlySchema(t *testing.T) {
	var eventMap Map
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "number"},
			"notRequired": {"type": "string"}},
			"required": ["test"]}`
	event, _ := GenerateEvent(schemaString, true)
	_ = json.Unmarshal(event, &eventMap)
	assert.Contains(t, eventMap, "test")
	assert.Empty(t, eventMap["notRequired"])
	assert.IsType(t, float64(1), eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}
