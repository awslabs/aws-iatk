package jsonrpc

import (
	"log"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMetadata_UserAgentValue(t *testing.T) {
	id := uuid.NewString()
	cases := []struct {
		name   string
		m      *Metadata
		expect string
	}{
		{
			name: "success",
			m: &Metadata{
				Client:        "python",
				ClientVersion: "3.10.9",
				Version:       "0.0.1",
				Caller:        "poll_events",
				DedupKey:      id,
			},
			expect: "python#3.10.9#0.0.1#poll_events#" + id,
		},
		{
			name: "invalid client",
			m: &Metadata{
				Client:        "xxxxx",
				ClientVersion: "3.10.9",
				Version:       "0.0.1",
				Caller:        "add_listener",
				DedupKey:      id,
			},
			expect: "unknown",
		},
		{
			name: "invalid version",
			m: &Metadata{
				Client:        "python",
				ClientVersion: "3.10.9",
				Version:       "x.0.3",
				Caller:        "retry_get_trace_tree_until",
				DedupKey:      id,
			},
			expect: "unknown",
		},
		{
			name: "version with beta tag",
			m: &Metadata{
				Client:        "python",
				ClientVersion: "3.10.9",
				Version:       "1.0.0-beta",
				Caller:        "retry_get_trace_tree_until",
				DedupKey:      id,
			},
			expect: "python#3.10.9#1.0.0-beta#retry_get_trace_tree_until#" + id,
		},
		{
			name: "invalid caller",
			m: &Metadata{
				Client:        "python",
				ClientVersion: "3.10.9",
				Version:       "0.0.2",
				Caller:        "with space",
				DedupKey:      id,
			},
			expect: "unknown",
		},
		{
			name: "invalid request_id",
			m: &Metadata{
				Client:        "python",
				ClientVersion: "3.10.9",
				Version:       "0.0.2",
				Caller:        "with space",
				DedupKey:      "invalid uuid format",
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
