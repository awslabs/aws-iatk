package xray

type Trace struct {
	Id            *string    `json:"id"`
	Duration      *float64   `json:"duration"`
	LimitExceeded *bool      `json:"limitExceeded"`
	Segments      []*Segment `json:"segments"`
}

type Segment struct {
	Id          *string                `json:"id"`
	Name        *string                `json:"name"`
	TraceId     *string                `json:"trace_id"`
	StartTime   *float64               `json:"start_time"`
	EndTime     *float64               `json:"end_time,omitempty"`
	InProgress  *bool                  `json:"in_progress,omitempty"`
	Service     *Service               `json:"service,omitempty"`
	User        *string                `json:"user,omitempty"`
	Origin      *string                `json:"origin,omitempty"`
	ParentId    *string                `json:"parent_id,omitempty"`
	Http        *Http                  `json:"http,omitempty"`
	Error       *bool                  `json:"error,omitempty"`
	Throttle    *bool                  `json:"throttle,omitempty"`
	Fault       *bool                  `json:"fault,omitempty"`
	Cause       interface{}            `json:"cause,omitempty"`
	Aws         interface{}            `json:"aws,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"` // only accepts string, number or boolean
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Subsegments []*Subsegment          `json:"subsegments,omitempty"`
	Links       []*Link                `json:"links,omitempty"`
}

type Subsegment struct {
	Id           *string                `json:"id"`
	Name         *string                `json:"name"`
	StartTime    *float64               `json:"start_time"`
	TraceId      *string                `json:"trace_id,omitempty"`
	EndTime      *float64               `json:"end_time,omitempty"`
	InProgress   *bool                  `json:"in_progress,omitempty"`
	Error        *bool                  `json:"error,omitempty"`
	Throttle     *bool                  `json:"throttle,omitempty"`
	Fault        *bool                  `json:"fault,omitempty"`
	Http         *Http                  `json:"http,omitempty"`
	Sql          *Sql                   `json:"sql,omitempty"`
	Namespace    *string                `json:"namespace,omitempty"`
	ParentId     *string                `json:"parent_id,omitempty"`
	Traced       *bool                  `json:"traced,omitempty"`
	PrecursorIds []*string              `json:"precursor_ids,omitempty"`
	Cause        interface{}            `json:"cause,omitempty"`
	Aws          interface{}            `json:"aws,omitempty"`
	Annotations  map[string]interface{} `json:"annotations,omitempty"` // only accepts string, number or boolean
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Type         *string                `json:"type,omitempty"`
	Subsegments  []*Subsegment          `json:"subsegments,omitempty"`
	Links        []*Link                `json:"links,omitempty"`
}

type Service struct {
	Type *string `json:"type"`
}

type Cause struct {
	WorkingDirectory *string      `json:"working_directory"`
	Paths            []*string    `json:"paths"`
	Exceptions       []*Exception `json:"exceptions"`
}

type Exception struct {
	Id        *string       `json:"id"`
	Message   *string       `json:"message"`
	Type      *string       `json:"type"`
	Remote    *bool         `json:"remote"`
	Truncated *int          `json:"truncated"`
	Skipped   *int          `json:"skipped"`
	Cause     *string       `json:"cause"`
	Stack     []*StackFrame `json:"stack"`
}

type StackFrame struct {
	Path  *string `json:"path"`
	Line  *int    `json:"line"`
	Label *string `json:"label"`
}

type Http struct {
	Request  *Request  `json:"request"`
	Response *Response `json:"response"`
}

type Request struct {
	Method        *string `json:"method"`
	ClientIp      *string `json:"client_ip"`
	Url           *string `json:"url"`
	UserAgent     *string `json:"user_agent"`
	XForwardedFor *bool   `json:"x_forwarded_for"`
}

type Response struct {
	Status        *int `json:"status"`
	ContentLength *int `json:"content_length"`
}

type Sql struct {
	ConnectionString *string `json:"connection_string"`
	Url              *string `json:"url"`
	SanitizedQuery   *string `json:"sanitized_query"`
	DatabaseType     *string `json:"database_type"`
	DatabaseVersion  *string `json:"database_version"`
	DriverVersion    *string `json:"driver_version"`
	User             *string `json:"user"`
	Preparation      *string `json:"preparation"`
}

type ReferenceType string

const (
	ReferenceTypeParent ReferenceType = "parent"
	ReferenceTypeChild  ReferenceType = "child"
)

type Link struct {
	TraceId    *string        `json:"trace_id"`
	Id         *string        `json:"id"`
	Attributes *LinkAttribute `json:"attributes"`
}

type LinkAttribute struct {
	ReferenceType *ReferenceType `json:"aws.xray.reserved.reference_type"`
}

type Tree struct {
	Root        *Segment          `json:"root"`
	Paths       [][]*Segment      `json:"paths"`
	SourceTrace *Trace            `json:"source_trace"`
	ChildTraces map[string]*Trace `json:"child_traces"`
}
