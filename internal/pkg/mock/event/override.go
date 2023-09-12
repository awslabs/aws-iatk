package event

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type EventMap map[string]interface{}

func EventOverridesMap(overrideMap EventMap, eventMap EventMap) {
	for key, element := range eventMap {
		if override, ok := overrideMap[key]; ok {
			eventMap[key] = override
			delete(overrideMap, key)
		} else if reflect.ValueOf(element).Kind() == reflect.Map {
			EventOverridesMap(overrideMap, element.(map[string]interface{}))
		} else if reflect.ValueOf(element).Kind() == reflect.Slice || reflect.ValueOf(element).Kind() == reflect.Array {
			EventOverridesArray(overrideMap, element.([]any))
		}
	}
}

func EventOverridesArray(overrideMap EventMap, eventArray []any) {
	for i := range eventArray {
		if reflect.ValueOf(eventArray[i]).Kind() == reflect.Map {
			EventOverridesMap(overrideMap, eventArray[i].(map[string]interface{}))
		} else if reflect.ValueOf(eventArray[i]).Kind() == reflect.Slice || reflect.ValueOf(eventArray[i]).Kind() == reflect.Array {
			EventOverridesArray(overrideMap, eventArray[i].([]any))
		}
	}
}

func EventOverrides(overrides []byte, event []byte) ([]byte, error) {
	eventMap := EventMap{}
	overridesMap := EventMap{}
	err := json.Unmarshal(event, &eventMap)
	if err != nil {
		return nil, fmt.Errorf("invalid event: can not decode: %w", err)
	}
	err = json.Unmarshal(overrides, &overridesMap)
	if err != nil {
		return nil, fmt.Errorf("invalid overrides: can not decode: %w", err)
	}
	EventOverridesMap(overridesMap, eventMap)
	for key, element := range overridesMap {
		eventMap[key] = element
	}
	overrideEvent, err := json.Marshal(eventMap)
	if err != nil {
		return nil, fmt.Errorf("cannot encode event into json: %w", err)
	}
	return overrideEvent, nil
}
