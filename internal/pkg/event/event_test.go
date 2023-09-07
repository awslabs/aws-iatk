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
	event, _ := GenerateEventString(schemaString)
	_ = json.Unmarshal(event, &eventMap)
	assert.NotEmpty(t, eventMap["test"])
	assert.IsType(t, "", eventMap["test"])
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
	event, _ := GenerateEventString(schemaString)
	_ = json.Unmarshal(event, &eventMap)
	assert.NotEmpty(t, eventMap["detail"])
	refMap := eventMap["detail"].(map[string]interface{})
	assert.NotEmpty(t, refMap["testRef"])
	assert.IsType(t, eventMap["test"], "")
	assert.Equal(t, 2, len(eventMap))
}

func TestCompilationError(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "random"}}}`
	_, err := GenerateEventString(schemaString)
	assert.Error(t, errors.New(`error compiling schema: json-schema \"temp.json\" compilation failed)`), err)
}

func TestSuccessfulNumberSchema(t *testing.T) {
	var eventMap Map
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": {"type": "number"}}}`
	event, _ := GenerateEventString(schemaString)
	_ = json.Unmarshal(event, &eventMap)
	assert.NotEmpty(t, eventMap["test"])
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
	event, _ := GenerateEventString(schemaString)
	_ = json.Unmarshal(event, &eventMap)
	assert.NotEmpty(t, eventMap["test"])
	arr := eventMap["test"].([]interface{})
	assert.IsType(t, []interface{}{}, arr)
	assert.IsType(t, "", arr[0])
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
	event, _ := GenerateEventString(schemaString)
	_ = json.Unmarshal(event, &eventMap)
	assert.NotEmpty(t, eventMap["testObject"])
	assert.IsType(t, map[string]any{}, eventMap["testObject"])
	objMap := eventMap["testObject"].(map[string]interface{})
	assert.NotEmpty(t, objMap["test"])
	assert.Equal(t, 1, len(eventMap))
}
