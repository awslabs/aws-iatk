package event

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateOpenApiEventErrors(t *testing.T) {
	cases := map[string]struct {
		schemaContent string
		schemaType    string
		eventRef      string
		expectErr     error
	}{
		"no eventRef provided": {
			schemaContent: `{"openapi": "3.0.0"}`,
			eventRef:      "",
			expectErr:     errors.New("no eventRef specified to generate a mock event"),
		},
		"missing SchemaContent": {
			schemaContent: "",
			eventRef:      "RandomEvent",
			expectErr:     errors.New("failed while loading schema due to error: invalid schema provided"),
		},
		"error while loading schema": {
			schemaContent: "invalid json content",
			eventRef:      "RandomEvent",
			expectErr:     errors.New("failed while loading schema due to error: invalid schema provided"),
		},
		"missing components in schema": {
			schemaContent: `{"openapi": "3.0.0"}`,
			eventRef:      "RandomEvent",
			expectErr:     errors.New("failed to generate a mock event, no components found in schema"),
		},
		"missing schemas in schema": {
			schemaContent: `{
				"openapi": "3.0.0",
				"components": {}}`,
			eventRef:  "RandomEvent",
			expectErr: errors.New("failed to generate a mock event, no schemas found under components in schema"),
		},
		"eventRef not found in schema": {
			schemaContent: `{
				"openapi": "3.0.0",
				"components": {
					"schemas": {
						"RandomName": {
							"type": "object",
							"required": ["name"],
							"properties": {"name": {"type": "string"}}
							}}}}`,
			eventRef:  "RandomEvent",
			expectErr: errors.New("provided eventRef \"RandomEvent\" not found in the schema"),
		},
		"error while generating mock event": {
			schemaContent: `{
				"openapi": "3.0.0",
				"info": {
					"version": "1.0.0",
					"title": "Event"
				},
				"components": {
					"schemas": {
						"RandomEvent": {
							"type": "object",
							"required": ["name"],
							"properties": {"name": {"type": "unknown-type"}}
							}}}}`,
			eventRef: "RandomEvent",
			expectErr: errors.New(
				"failed to generate a mock event for the provided schema due to error: invalid or unsupported property type \"unknown-type\" found for property \"name\"",
			),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			schemaType := "OpenApi3"
			var schemaContent *string
			if tt.schemaContent == "" {
				schemaContent = nil
			} else {
				schemaContent = &tt.schemaContent
			}
			schema := &Schema{
				SchemaContent: schemaContent,
				SchemaType:    &schemaType,
				EventRef:      &tt.eventRef,
			}

			generatedEvent, err := GenerateOpenapiEvent(schema, false)
			assert.EqualError(t, err, tt.expectErr.Error())
			assert.Nil(t, generatedEvent)
		})
	}
}

func TestSuccessfullStringSchema(t *testing.T) {
	var eventMap EventMap
	schemaType := "OpenApi3"
	EventRef := "RandomEvent"
	schemaContent := `{
		"openapi": "3.0.0",
		"components": {
			"schemas": {
				"RandomEvent": {
					"type": "object",
					"required": ["name"],
					"properties": {"name": {"type": "string"}}
	}}}}`
	schema := &Schema{SchemaContent: &schemaContent, SchemaType: &schemaType, EventRef: &EventRef}

	event, err := GenerateOpenapiEvent(schema, false)
	_ = json.Unmarshal(event, &eventMap)

	assert.Nil(t, err)
	assert.Contains(t, eventMap, "name")
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfullEnumSchema(t *testing.T) {
	var eventMap EventMap
	schemaType := "OpenApi3"
	EventRef := "RandomEvent"
	schemaContent := `{
		"openapi": "3.0.0",
		"components": {
			"schemas": {
				"RandomEvent": {
					"type": "object",
					"required": ["day"],
					"properties": 
					{"day": {"type": "string", "enum": ["Sunday", "Monday"]}}
	}}}}`
	schema := &Schema{SchemaContent: &schemaContent, SchemaType: &schemaType, EventRef: &EventRef}

	event, err := GenerateOpenapiEvent(schema, false)
	_ = json.Unmarshal(event, &eventMap)

	assert.Nil(t, err)
	assert.Contains(t, eventMap, "day")
	assert.Equal(t, "Sunday", eventMap["day"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfullNumberSchema(t *testing.T) {
	var eventMap EventMap
	schemaType := "OpenApi3"
	EventRef := "RandomEvent"
	schemaContent := `{
		"openapi": "3.0.0",
		"components": {
			"schemas": {
				"RandomEvent": {
					"type": "object",
					"required": ["test"],
					"properties": 
					{"test": {"type": "number"}}
	}}}}`
	schema := &Schema{SchemaContent: &schemaContent, SchemaType: &schemaType, EventRef: &EventRef}

	event, err := GenerateOpenapiEvent(schema, false)
	_ = json.Unmarshal(event, &eventMap)

	assert.Nil(t, err)
	assert.Contains(t, eventMap, "test")
	assert.Equal(t, float64(0), eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfullArraySchema(t *testing.T) {
	var eventMap EventMap
	schemaType := "OpenApi3"
	EventRef := "RandomEvent"
	schemaContent := `{
		"openapi": "3.0.0",
		"components": {
			"schemas": {
				"RandomEvent": {
					"type": "object",
					"required": ["test"],
					"properties": 
					{"test": {"type": "array", items": {"type": "string"}}}
	}}}}`
	schema := &Schema{SchemaContent: &schemaContent, SchemaType: &schemaType, EventRef: &EventRef}

	event, err := GenerateOpenapiEvent(schema, false)
	_ = json.Unmarshal(event, &eventMap)

	assert.Nil(t, err)
	assert.Contains(t, eventMap, "test")
	assert.IsType(t, []interface{}{}, eventMap["test"])
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfullObjectSchema(t *testing.T) {
	var eventMap EventMap
	schemaType := "OpenApi3"
	EventRef := "RandomEvent"
	schemaContent := `{
		"openapi": "3.0.0",
		"components": {
			"schemas": {
				"RandomEvent": {
					"type": "object",
					"required": ["test"],
					"properties": 
					{"testObject": {"type": "object", "properties": {"test": {"type": "string"}}}}
	}}}}`
	schema := &Schema{SchemaContent: &schemaContent, SchemaType: &schemaType, EventRef: &EventRef}

	event, err := GenerateOpenapiEvent(schema, false)
	_ = json.Unmarshal(event, &eventMap)

	assert.Nil(t, err)
	assert.Contains(t, eventMap, "testObject")
	assert.NotEmpty(t, eventMap["testObject"])
	assert.IsType(t, map[string]any{}, eventMap["testObject"])

	objMap := eventMap["testObject"].(map[string]interface{})
	assert.Contains(t, objMap, "test")
	assert.Equal(t, 1, len(eventMap))
}

func TestSuccessfullRefObjectSchema(t *testing.T) {
	var eventMap EventMap
	schemaType := "OpenApi3"
	EventRef := "RandomEvent"
	schemaContent := `{
		"openapi": "3.0.0",
		"components": {
			"schemas": {
				"RandomEvent": {
					"type": "object",
					"required": ["name", "workPlace"],
					"properties": {"name": {"type": "string"}, "workPlace": {"$ref": "#/components/schemas/Office"}}
				},
				"Office": {
					"type": "object",
					"required": ["company"],
					"properties": {"company": {"type": "string"}}
				},
	}}}`
	schema := &Schema{SchemaContent: &schemaContent, SchemaType: &schemaType, EventRef: &EventRef}

	event, err := GenerateOpenapiEvent(schema, false)
	_ = json.Unmarshal(event, &eventMap)

	assert.Nil(t, err)
	assert.Contains(t, eventMap, "name")
	assert.Contains(t, eventMap, "workPlace")
	assert.Equal(t, 2, len(eventMap))

	objMap := eventMap["workPlace"].(map[string]interface{})
	assert.Contains(t, objMap, "company")
	assert.Equal(t, 1, len(objMap))
}

func TestSuccessfullRequiredOnlySchema(t *testing.T) {
	var eventMapAllProperties EventMap
	schemaType := "OpenApi3"
	EventRef := "RandomEvent"
	schemaContent := `{
		"openapi": "3.0.0",
		"components": {
			"schemas": {
				"RandomEvent": {
					"type": "object",
					"required": ["name"],
					"properties": {"name": {"type": "string"}, "age": {"type": "integer"}, "dob": {"type": "string", "format": "date"}}
	}}}}`
	schema := &Schema{SchemaContent: &schemaContent, SchemaType: &schemaType, EventRef: &EventRef}

	eventAllProperties, err := GenerateOpenapiEvent(schema, false)
	_ = json.Unmarshal(eventAllProperties, &eventMapAllProperties)

	assert.Nil(t, err)
	assert.Contains(t, eventMapAllProperties, "name")
	assert.Contains(t, eventMapAllProperties, "age")
	assert.Contains(t, eventMapAllProperties, "dob")
	assert.Equal(t, 3, len(eventMapAllProperties))

	var eventMap EventMap
	event, err := GenerateOpenapiEvent(schema, true)
	_ = json.Unmarshal(event, &eventMap)

	assert.Nil(t, err)
	assert.Contains(t, eventMap, "name")
	assert.Equal(t, 1, len(eventMap))
}
