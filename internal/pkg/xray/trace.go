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

func GetTraces(ctx context.Context, api BatchGetTracesAPI, traceIds []string) (map[string]*Trace, error) {
	input := &xray.BatchGetTracesInput{
		TraceIds: traceIds,
	}

	unprocessedTraceIds := make([]string, len(traceIds))
	copy(unprocessedTraceIds, traceIds)

	traceMap := make(map[string]*Trace)

	// Loop until all provided traces are processed
	for len(unprocessedTraceIds) > 0 {
		paginator := xray.NewBatchGetTracesPaginator(api, input)

		for paginator.HasMorePages() {
			resp, err := paginator.NextPage(ctx)

			if err != nil {
				return nil, fmt.Errorf("failed to get Traces: %v", err)
			}

			unprocessedTraceIds = resp.UnprocessedTraceIds

			// Create Trace and add to map if processed
			for _, trace := range resp.Traces {
				if _, ok := traceMap[aws.ToString(trace.Id)]; !ok {
					trace_object, err := TraceFromApiResponse(&trace)

					if err != nil {
						return nil, fmt.Errorf("failed to load trace details for trace id: %v", aws.ToString(trace.Id))
					}

					traceMap[aws.ToString(trace.Id)] = trace_object
				}
			}
		}
	}
	return traceMap, nil
}

//go:generate mockery --name BatchGetTracesAPI
type BatchGetTracesAPI interface {
	BatchGetTraces(context.Context, *xray.BatchGetTracesInput, ...func(*xray.Options)) (*xray.BatchGetTracesOutput, error)
}

//go:generate mockery --name getTracesFunc
type getTracesFunc func(ctx context.Context, api BatchGetTracesAPI, traceIds []string) (map[string]*Trace, error)
