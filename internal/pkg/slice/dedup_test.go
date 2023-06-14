package slice_test

import (
	"testing"
	"zion/internal/pkg/slice"

	"github.com/stretchr/testify/assert"
)

func TestDedup(t *testing.T) {
	cases := map[string]struct {
		input  []string
		expect []string
	}{
		"Success: returns same slice": {
			input:  []string{"a", "b"},
			expect: []string{"a", "b"},
		},
		"Success: dedup": {
			input:  []string{"a", "b", "a", "b"},
			expect: []string{"a", "b"},
		},
		"Success: empty slice": {
			input:  []string{},
			expect: []string{},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			r := slice.Dedup(tt.input)

			assert.Equal(t, tt.expect, r)
		})
	}
}
