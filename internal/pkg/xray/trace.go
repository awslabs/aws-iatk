package xray

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	"github.com/aws/aws-sdk-go-v2/service/xray/types"
)

func TraceFromApiResponse(trace *types.Trace) (*Trace, error) {
	segments := []*Segment{}
	for _, sm := range trace.Segments {
		doc := aws.ToString(sm.Document)
		segment, err := SegmentFromDocument(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve segment document: %w", err)
		}
		segments = append(segments, segment)
	}
	return &Trace{
		Id:            trace.Id,
		Duration:      trace.Duration,
		LimitExceeded: trace.LimitExceeded,
		Segments:      segments,
	}, nil
}

type xrayClient interface {
	BatchGetTraceAPI
}

type BatchGetTraceAPI interface {
	BatchGetTrace(ctx context.Context, params *xray.BatchGetTracesInput, optFns ...func(*xray.Options)) (xray.BatchGetTracesOutput, error)
}
