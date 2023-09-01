package xray

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/xray"
	"github.com/aws/aws-sdk-go-v2/service/xray/types"
	"github.com/stretchr/testify/assert"
)

func TestTraceFromApiResponseSuccess(t *testing.T) {
	cases := []struct {
		trace         types.Trace
		traceId       string
		duration      float64
		limitExceeded bool
		numSegments   int
	}{
		{
			trace: types.Trace{
				Id:            aws.String("1-64de5a99-5d09aa705e56bbd0152548cb"),
				Duration:      aws.Float64(2.783),
				LimitExceeded: aws.Bool(false),
				Segments: []types.Segment{
					{
						Id:       aws.String("3705e19822f3db45"),
						Document: aws.String("{\"id\":\"3705e19822f3db45\",\"name\":\"sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\",\"start_time\":1.692293788127201E9,\"trace_id\":\"1-64de5a99-5d09aa705e56bbd0152548cb\",\"end_time\":1.6922937887753212E9,\"parent_id\":\"60af780ecc5b639d\",\"aws\":{\"account_id\":\"012345678901\",\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\",\"resource_names\":[\"sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"]},\"origin\":\"AWS::Lambda::Function\",\"subsegments\":[{\"id\":\"d42ac4bfc7906b53\",\"name\":\"Initialization\",\"start_time\":1.6922937877737608E9,\"end_time\":1.6922937881258464E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"83f9591e84c79024\",\"name\":\"Overhead\",\"start_time\":1.6922937887534742E9,\"end_time\":1.6922937887750967E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"5640de914527757d\",\"name\":\"Invocation\",\"start_time\":1.6922937881274729E9,\"end_time\":1.6922937887534282E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"},\"subsegments\":[{\"id\":\"aabeb336dc5bce90\",\"name\":\"Events\",\"start_time\":1.692293788194E9,\"end_time\":1.692293788694E9,\"http\":{\"response\":{\"status\":200,\"content_length\":85}},\"aws\":{\"retries\":1,\"region\":\"us-east-1\",\"operation\":\"PutEvents\",\"request_id\":\"e957a46b-60da-47be-b4af-8b947507c9ab\"},\"namespace\":\"aws\"}]}]}"),
					},
				},
			},
			traceId:       "1-64de5a99-5d09aa705e56bbd0152548cb",
			duration:      2.783,
			limitExceeded: false,
			numSegments:   1,
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			trace, err := TraceFromApiResponse(&tt.trace)
			assert.Nil(t, err)
			assert.Equal(t, tt.traceId, *trace.Id)
			assert.Equal(t, tt.duration, *trace.Duration)
			assert.Equal(t, tt.limitExceeded, *trace.LimitExceeded)
			assert.Len(t, trace.Segments, tt.numSegments)
		})
	}
}

func TestTraceFromApiResponseFailure(t *testing.T) {
	cases := []struct {
		trace types.Trace
		err   error
	}{
		{
			trace: types.Trace{
				Id:            aws.String("1-64de5a99-5d09aa705e56bbd0152548cb"),
				Duration:      aws.Float64(2.783),
				LimitExceeded: aws.Bool(false),
				Segments: []types.Segment{
					{
						Id:       aws.String("3705e19822f3db45"),
						Document: aws.String(""),
					},
				},
			},
			err: errors.New("failed to resolve segment document: failed to decode segment document: unexpected end of JSON input"),
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			trace, err := TraceFromApiResponse(&tt.trace)
			assert.Nil(t, trace)
			assert.EqualError(t, err, tt.err.Error())
		})
	}
}

func TestGetTraces(t *testing.T) {
	cases := map[string]struct {
		mockBatchGetTracesAPI func(ctx context.Context, traceIds []string) *MockBatchGetTracesAPI
		traceIds              []string
		expectErr             error
	}{
		"success": {
			traceIds: []string{"1-64de5a99-5d09aa705e56bbd0152548cb", "1-12de5a99-5d09hh705e56aad0152548cb"},
			mockBatchGetTracesAPI: func(ctx context.Context, traceIds []string) *MockBatchGetTracesAPI {
				api := NewMockBatchGetTracesAPI(t)
				api.EXPECT().
					BatchGetTraces(ctx, &xray.BatchGetTracesInput{
						TraceIds: traceIds,
					}).
					Return(&xray.BatchGetTracesOutput{
						Traces: []types.Trace{
							{
								Id:            aws.String(traceIds[0]),
								Duration:      aws.Float64(2.783),
								LimitExceeded: aws.Bool(false),
								Segments: []types.Segment{
									{
										Id:       aws.String("3705e19822f3db45"),
										Document: aws.String("{\"id\":\"3705e19822f3db45\",\"name\":\"sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\",\"start_time\":1.692293788127201E9,\"trace_id\":\"1-64de5a99-5d09aa705e56bbd0152548cb\",\"end_time\":1.6922937887753212E9,\"parent_id\":\"60af780ecc5b639d\",\"aws\":{\"account_id\":\"012345678901\",\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\",\"resource_names\":[\"sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"]},\"origin\":\"AWS::Lambda::Function\",\"subsegments\":[{\"id\":\"d42ac4bfc7906b53\",\"name\":\"Initialization\",\"start_time\":1.6922937877737608E9,\"end_time\":1.6922937881258464E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"83f9591e84c79024\",\"name\":\"Overhead\",\"start_time\":1.6922937887534742E9,\"end_time\":1.6922937887750967E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"5640de914527757d\",\"name\":\"Invocation\",\"start_time\":1.6922937881274729E9,\"end_time\":1.6922937887534282E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"},\"subsegments\":[{\"id\":\"aabeb336dc5bce90\",\"name\":\"Events\",\"start_time\":1.692293788194E9,\"end_time\":1.692293788694E9,\"http\":{\"response\":{\"status\":200,\"content_length\":85}},\"aws\":{\"retries\":1,\"region\":\"us-east-1\",\"operation\":\"PutEvents\",\"request_id\":\"e957a46b-60da-47be-b4af-8b947507c9ab\"},\"namespace\":\"aws\"}]}]}"),
									},
								},
							},
							{
								Id:            aws.String(traceIds[1]),
								Duration:      aws.Float64(2.783),
								LimitExceeded: aws.Bool(false),
								Segments: []types.Segment{
									{
										Id:       aws.String("2405e19822f3db45"),
										Document: aws.String("{\"id\":\"2405e19822f3db45\",\"name\":\"sam-ts-app-z-NewOrderConsumerFunction-Adv4wKsAxhOi\",\"start_time\":1.682293788127201E9,\"trace_id\":\"1-12de5a99-5d09hh705e56aad0152548cb\",\"end_time\":1.6922937887753212E9,\"parent_id\":\"60af780ecc5b639d\",\"aws\":{\"account_id\":\"012345678901\",\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\",\"resource_names\":[\"sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"]},\"origin\":\"AWS::Lambda::Function\",\"subsegments\":[{\"id\":\"d42ac4bfc7906b53\",\"name\":\"Initialization\",\"start_time\":1.6922937877737608E9,\"end_time\":1.6922937881258464E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"83f9591e84c79024\",\"name\":\"Overhead\",\"start_time\":1.6922937887534742E9,\"end_time\":1.6922937887750967E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"5640de914527757d\",\"name\":\"Invocation\",\"start_time\":1.6922937881274729E9,\"end_time\":1.6922937887534282E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"},\"subsegments\":[{\"id\":\"aabeb336dc5bce90\",\"name\":\"Events\",\"start_time\":1.692293788194E9,\"end_time\":1.692293788694E9,\"http\":{\"response\":{\"status\":200,\"content_length\":85}},\"aws\":{\"retries\":1,\"region\":\"us-east-1\",\"operation\":\"PutEvents\",\"request_id\":\"e957a46b-60da-47be-b4af-8b947507c9ab\"},\"namespace\":\"aws\"}]}]}"),
									},
								},
							},
						},
					}, nil)
				return api
			},
			expectErr: nil,
		},
		"api failed": {
			traceIds: []string{"1-64de5a99-5d09aa705e56bbd0152548cb", "1-12de5a99-5d09hh705e56aad0152548cb"},
			mockBatchGetTracesAPI: func(ctx context.Context, traceIds []string) *MockBatchGetTracesAPI {
				api := NewMockBatchGetTracesAPI(t)
				api.EXPECT().
					BatchGetTraces(ctx, &xray.BatchGetTracesInput{
						TraceIds: traceIds,
					}).
					Return(nil, errors.New("api failed"))
				return api
			},
			expectErr: errors.New("failed to get Traces: api failed"),
		},
		"load trace failed": {
			traceIds: []string{"1-64de5a99-5d09aa705e56bbd0152548cb"},
			mockBatchGetTracesAPI: func(ctx context.Context, traceIds []string) *MockBatchGetTracesAPI {
				api := NewMockBatchGetTracesAPI(t)
				api.EXPECT().
					BatchGetTraces(ctx, &xray.BatchGetTracesInput{
						TraceIds: traceIds,
					}).
					Return(&xray.BatchGetTracesOutput{
						Traces: []types.Trace{
							{
								Id:            aws.String("1-64de5a99-5d09aa705e56bbd0152548cb"),
								Duration:      aws.Float64(2.783),
								LimitExceeded: aws.Bool(false),
								Segments: []types.Segment{
									{
										Id:       aws.String("3705e19822f3db45"),
										Document: aws.String(""),
									},
								}},
						},
					}, nil)
				return api
			},
			expectErr: errors.New("failed to load trace details for trace id: 1-64de5a99-5d09aa705e56bbd0152548cb"),
		},
		"found unprocessed traces": {
			traceIds: []string{"1-64de5a99-5d09aa705e56bbd0152548cb", "1-12de5a99-5d09hh705e56aad0152548cb"},
			mockBatchGetTracesAPI: func(ctx context.Context, traceIds []string) *MockBatchGetTracesAPI {
				api := NewMockBatchGetTracesAPI(t)
				api.EXPECT().
					BatchGetTraces(ctx, &xray.BatchGetTracesInput{
						TraceIds: traceIds,
					}).
					Return(&xray.BatchGetTracesOutput{
						Traces: []types.Trace{
							{
								Id:            aws.String(traceIds[1]),
								Duration:      aws.Float64(2.783),
								LimitExceeded: aws.Bool(false),
								Segments: []types.Segment{
									{
										Id:       aws.String("2405e19822f3db45"),
										Document: aws.String("{\"id\":\"2405e19822f3db45\",\"name\":\"sam-ts-app-z-NewOrderConsumerFunction-Adv4wKsAxhOi\",\"start_time\":1.682293788127201E9,\"trace_id\":\"1-12de5a99-5d09hh705e56aad0152548cb\",\"end_time\":1.6922937887753212E9,\"parent_id\":\"60af780ecc5b639d\",\"aws\":{\"account_id\":\"012345678901\",\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\",\"resource_names\":[\"sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"]},\"origin\":\"AWS::Lambda::Function\",\"subsegments\":[{\"id\":\"d42ac4bfc7906b53\",\"name\":\"Initialization\",\"start_time\":1.6922937877737608E9,\"end_time\":1.6922937881258464E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"83f9591e84c79024\",\"name\":\"Overhead\",\"start_time\":1.6922937887534742E9,\"end_time\":1.6922937887750967E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"5640de914527757d\",\"name\":\"Invocation\",\"start_time\":1.6922937881274729E9,\"end_time\":1.6922937887534282E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"},\"subsegments\":[{\"id\":\"aabeb336dc5bce90\",\"name\":\"Events\",\"start_time\":1.692293788194E9,\"end_time\":1.692293788694E9,\"http\":{\"response\":{\"status\":200,\"content_length\":85}},\"aws\":{\"retries\":1,\"region\":\"us-east-1\",\"operation\":\"PutEvents\",\"request_id\":\"e957a46b-60da-47be-b4af-8b947507c9ab\"},\"namespace\":\"aws\"}]}]}"),
									},
								},
							},
						},
						UnprocessedTraceIds: []string{"1-64de5a99-5d09aa705e56bbd0152548cb"},
					}, nil).Once()

				api.EXPECT().
					BatchGetTraces(ctx, &xray.BatchGetTracesInput{
						TraceIds: traceIds,
					}).
					Return(&xray.BatchGetTracesOutput{
						Traces: []types.Trace{
							{
								Id:            aws.String(traceIds[0]),
								Duration:      aws.Float64(2.783),
								LimitExceeded: aws.Bool(false),
								Segments: []types.Segment{
									{
										Id:       aws.String("3705e19822f3db45"),
										Document: aws.String("{\"id\":\"3705e19822f3db45\",\"name\":\"sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\",\"start_time\":1.692293788127201E9,\"trace_id\":\"1-64de5a99-5d09aa705e56bbd0152548cb\",\"end_time\":1.6922937887753212E9,\"parent_id\":\"60af780ecc5b639d\",\"aws\":{\"account_id\":\"012345678901\",\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\",\"resource_names\":[\"sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"]},\"origin\":\"AWS::Lambda::Function\",\"subsegments\":[{\"id\":\"d42ac4bfc7906b53\",\"name\":\"Initialization\",\"start_time\":1.6922937877737608E9,\"end_time\":1.6922937881258464E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"83f9591e84c79024\",\"name\":\"Overhead\",\"start_time\":1.6922937887534742E9,\"end_time\":1.6922937887750967E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"}},{\"id\":\"5640de914527757d\",\"name\":\"Invocation\",\"start_time\":1.6922937881274729E9,\"end_time\":1.6922937887534282E9,\"aws\":{\"function_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:sam-ts-app-z-NewOrderConsumerFunction-Qqv4wKsAxhOi\"},\"subsegments\":[{\"id\":\"aabeb336dc5bce90\",\"name\":\"Events\",\"start_time\":1.692293788194E9,\"end_time\":1.692293788694E9,\"http\":{\"response\":{\"status\":200,\"content_length\":85}},\"aws\":{\"retries\":1,\"region\":\"us-east-1\",\"operation\":\"PutEvents\",\"request_id\":\"e957a46b-60da-47be-b4af-8b947507c9ab\"},\"namespace\":\"aws\"}]}]}"),
									},
								},
							},
						},
					}, nil).Once()
				return api
			},
			expectErr: nil,
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			mockAPI := tt.mockBatchGetTracesAPI(ctx, tt.traceIds)
			traceMap, err := GetTraces(ctx, mockAPI, tt.traceIds)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				for _, traceId := range tt.traceIds {
					assert.Equal(t, traceId, aws.ToString(traceMap[traceId].Id))
					assert.Len(t, traceMap[traceId].Segments, 1)
					assert.Nil(t, err)
				}
			}
		})
	}
}
