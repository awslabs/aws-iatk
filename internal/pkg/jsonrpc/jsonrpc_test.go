package jsonrpc

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata_UserAgentValue(t *testing.T) {
	cases := []struct {
		name   string
		m      *Metadata
		expect string
	}{
		{
			name: "success",
			m: &Metadata{
				Client:  "python",
				Version: "0.0.1",
				Caller:  "poll_events",
			},
			expect: "python#0.0.1#poll_events",
		},
		{
			name: "invalid client",
			m: &Metadata{
				Client:  "xxxxx",
				Version: "0.0.1",
				Caller:  "add_listener",
			},
			expect: "unknown",
		},
		{
			name: "invalid version",
			m: &Metadata{
				Client:  "python",
				Version: "x.0.3",
				Caller:  "retry_get_trace_tree_until",
			},
			expect: "unknown",
		},
		{
			name: "version with beta tag",
			m: &Metadata{
				Client:  "python",
				Version: "1.0.0-beta",
				Caller:  "retry_get_trace_tree_until",
			},
			expect: "unknown",
		},
		{
			name: "invalid caller",
			m: &Metadata{
				Client:  "python",
				Version: "0.0.2",
				Caller:  "with space",
			},
			expect: "unknown",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			log.Println(tt.m.Client)
			log.Println(tt.m.Version)
			log.Println(tt.m.Caller)
			log.Printf("expect: %v\n", tt.expect)
			actual := tt.m.UserAgentValue()
			log.Printf("actual: %v\n", actual)
			log.Println("")
			assert.Equal(t, tt.expect, actual)
		})
	}
}
