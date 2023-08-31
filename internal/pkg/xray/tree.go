package xray

import (
	"context"
)

func NewTree(ctx context.Context, api BatchGetTracesAPI, sourceTraceId string) (*Tree, error) {
	return &Tree{}, nil
}
