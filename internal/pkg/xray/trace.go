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

func Get(ctx context.Context, api BatchGetTracesAPI, traceIds []string) (map[string]Trace, error) {
	input := &xray.BatchGetTracesInput{
		TraceIds: traceIds,
	}

	paginator := xray.NewBatchGetTracesPaginator(api, input)

	traceMap := make(map[string]Trace)

	for paginator.HasMorePages() {
		resp, err := paginator.NextPage(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to get Traces: %v", err)
		}

		if len(resp.UnprocessedTraceIds) > 0 {
			return nil, fmt.Errorf("the following traces could not be processed: %v", resp.UnprocessedTraceIds)
		}

		for _, trace := range resp.Traces {
			trace_object, err := TraceFromApiResponse(&trace)

			if err != nil {
				return nil, fmt.Errorf("failed to get trace: %v", aws.ToString(trace.Id))
			}

			traceMap[aws.ToString(trace.Id)] = *trace_object
		}
	}
	return traceMap, nil
}

type BatchGetTracesAPI interface {
	BatchGetTraces(context.Context, *xray.BatchGetTracesInput, ...func(*xray.Options)) (*xray.BatchGetTracesOutput, error)
}
