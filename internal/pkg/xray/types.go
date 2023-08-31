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
	Id        *string       `json:"id,omitempty"`
	Message   *string       `json:"message,omitempty"`
	Type      *string       `json:"type,omitempty"`
	Remote    *bool         `json:"remote,omitempty"`
	Truncated *int          `json:"truncated,omitempty"`
	Skipped   *int          `json:"skipped,omitempty"`
	Cause     *string       `json:"cause,omitempty"`
	Stack     []*StackFrame `json:"stack,omitempty"`
}

type StackFrame struct {
	Path  *string `json:"path,omitempty"`
	Line  *int    `json:"line,omitempty"`
	Label *string `json:"label,omitempty"`
}

type Http struct {
	Request  *Request  `json:"request,omitempty"`
	Response *Response `json:"response,omitempty"`
}

type Request struct {
	Method        *string `json:"method,omitempty"`
	ClientIp      *string `json:"client_ip,omitempty"`
	Url           *string `json:"url,omitempty"`
	UserAgent     *string `json:"user_agent,omitempty"`
	XForwardedFor *bool   `json:"x_forwarded_for,omitempty"`
	Traced        *bool   `json:"traced,omitempty"`
}

type Response struct {
	Status        *int `json:"status,omitempty"`
	ContentLength *int `json:"content_length,omitempty"`
}

type Sql struct {
	ConnectionString *string `json:"connection_string,omitempty"`
	Url              *string `json:"url,omitempty"`
	SanitizedQuery   *string `json:"sanitized_query,omitempty"`
	DatabaseType     *string `json:"database_type,omitempty"`
	DatabaseVersion  *string `json:"database_version,omitempty"`
	DriverVersion    *string `json:"driver_version,omitempty"`
	User             *string `json:"user,omitempty"`
	Preparation      *string `json:"preparation,omitempty"`
}

type ReferenceType string

const (
	ReferenceTypeParent ReferenceType = "parent"
	ReferenceTypeChild  ReferenceType = "child"
)

type Link struct {
	TraceId    *string         `json:"trace_id"`
	Id         *string         `json:"id"`
	Attributes *LinkAttributes `json:"attributes"`
}

type LinkAttributes struct {
	ReferenceType *ReferenceType `json:"aws.xray.reserved.reference_type"`
}

type Tree struct {
	Root        *Segment     `json:"root"`
	Paths       [][]*Segment `json:"paths"`
	SourceTrace *Trace       `json:"source_trace"`
}
