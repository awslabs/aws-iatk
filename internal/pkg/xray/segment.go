package xray

import (
	"encoding/json"
	"fmt"
)

func SegmentFromDocument(doc string) (*Segment, error) {
	var decoded Segment
	err := json.Unmarshal([]byte(doc), &decoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode segment document: %w", err)
	}
	return &decoded, nil
}
