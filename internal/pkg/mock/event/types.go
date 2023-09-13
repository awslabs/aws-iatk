package event

type Schema struct {
	SchemaContent *string
	SchemaType    *string
	EventRef      *string
}

type EventMap map[string]interface{}
