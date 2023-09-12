package event

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOverrideString(t *testing.T) {
	var eventMap EventMap
	eventInput := EventMap{"test": "before"}
	overrideInput := EventMap{"test": "after"}
	eventEncoded, _ := json.Marshal(eventInput)
	overrideEncoded, _ := json.Marshal(overrideInput)
	event, _ := EventOverrides(overrideEncoded, eventEncoded)
	_ = json.Unmarshal(event, &eventMap)
	assert.Equal(t, "after", eventMap["test"])
}

func TestOverrideObject(t *testing.T) {
	var eventMap EventMap
	eventInput := EventMap{"test": EventMap{"test1": "before"}}
	overrideInput := EventMap{"test": "after"}
	eventEncoded, _ := json.Marshal(eventInput)
	overrideEncoded, _ := json.Marshal(overrideInput)
	event, _ := EventOverrides(overrideEncoded, eventEncoded)
	_ = json.Unmarshal(event, &eventMap)
	assert.Equal(t, "after", eventMap["test"])
}

func TestOverrideObjectDefault(t *testing.T) {
	var eventMap EventMap
	eventInput := EventMap{}
	overrideInput := EventMap{"test": "after", "test1": "after"}
	eventEncoded, _ := json.Marshal(eventInput)
	overrideEncoded, _ := json.Marshal(overrideInput)
	event, _ := EventOverrides(overrideEncoded, eventEncoded)
	_ = json.Unmarshal(event, &eventMap)
	assert.Equal(t, "after", eventMap["test"])
	assert.Equal(t, "after", eventMap["test1"])
}

func TestOverrideInNestedObject(t *testing.T) {
	var eventMap EventMap
	eventInput := EventMap{"test1": EventMap{"test": "before"}}
	overrideInput := EventMap{"test": "after"}
	eventEncoded, _ := json.Marshal(eventInput)
	overrideEncoded, _ := json.Marshal(overrideInput)
	event, _ := EventOverrides(overrideEncoded, eventEncoded)
	_ = json.Unmarshal(event, &eventMap)
	assert.Equal(t, "after", eventMap["test1"].(map[string]interface{})["test"])
}

func TestOverrideInNestedArray(t *testing.T) {
	var eventMap EventMap
	eventInput := EventMap{"test1": []any{EventMap{"test": "before"}}}
	overrideInput := EventMap{"test": "after"}
	eventEncoded, _ := json.Marshal(eventInput)
	overrideEncoded, _ := json.Marshal(overrideInput)
	event, _ := EventOverrides(overrideEncoded, eventEncoded)
	_ = json.Unmarshal(event, &eventMap)
	assert.Equal(t, "after", eventMap["test1"].([]any)[0].(map[string]interface{})["test"])
}
