package xray

import (
	"context"
)

func NewTree(ctx context.Context, api BatchGetTracesAPI, sourceTraceId string, fetchChildLinkedTraces bool) (*Tree, error) {
	return &Tree{}, nil
}
