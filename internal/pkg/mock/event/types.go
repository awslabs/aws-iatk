package event

type SchemaType string

const (
	OpenApi3         = "OpenApi3"
	JSONSchemaDraft4 = "JSONSchemaDraft4"
)

type Schema struct {
	SchemaContent *string
	SchemaType    *SchemaType
	EventRef      *string
}

type EventMap map[string]interface{}
