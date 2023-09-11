package validate

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSuccessfulValidateString(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": { "type": "string" }}}`
	event := []byte(`{"test": "testing"}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.True(t, valid)
}

func TestSuccessfulFailureString(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": { "type": "string" }}}`
	event := []byte(`{"test": 123}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.False(t, valid)
}

func TestSuccessfulFailureInteger(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": { "type": "integer" }}}`
	event := []byte(`{"test": "testing"}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.False(t, valid)
}

func TestSuccessfulValidateNumber(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": { "type": "number" }}}`
	event := []byte(`{"test": 10123}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.True(t, valid)
}

func TestSuccessfulValidateBoolean(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": { 
			"test": { "type": "boolean" }}}`
	event := []byte(`{"test": false}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.True(t, valid)
}

func TestSuccessfulValidateArray(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": {
			"test": {
				"items": {
				  "type": "string"
				},
				"type": "array"
			}}}`
	event := []byte(`{"test":["test1", "test2"]}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.True(t, valid)
}

func TestSuccessfulValidateObject(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": {
			"testObject": {
				"properties": {
				  "test": {
           			 "type": "string"
					}},
				"type": "object"
			}}}`
	event := []byte(`{"testObject":{"test":"testString"}}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.True(t, valid)
}

func TestSuccessfulValidateNestedArray(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"properties": {
			"test": {
				"items": {
				  "type": "array",
				  "items": {
						"type": "object",
						"properties": {
							"test1": {
								"type": "string"
							}
						}
					}
				},
				"type": "array"
			}}}`
	event := []byte(`{"test:[[{"test1":"testString"}]]"}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.True(t, valid)
}

func TestSuccessfulValidateRef(t *testing.T) {
	schemaString := `{ "$schema": "http://json-schema.org/draft-04/schema#", 
		"definitions": { 
			"RoomStateChange": {
				"properties": { 
					"testRef": { "type": "string" }}}}, 
		"properties": { "test": { "type": "string" }, 
		"detail": { "$ref": "#/definitions/RoomStateChange" 
		}}}`
	event := []byte(`{"test":"test1", "detail": {"testRef": "test2"}}`)
	valid, _ := ValidateEvent(schemaString, event)
	assert.True(t, valid)
}
