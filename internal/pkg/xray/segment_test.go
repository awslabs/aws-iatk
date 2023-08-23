package xray

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegmentFromDocumentSuccess(t *testing.T) {
	cases := []struct {
		document  string
		id        string
		traceId   string
		origin    string
		url       string
		startTime float64
	}{
		{
			document:  "{\"id\":\"2b4e0ef3499cd781\",\"name\":\"some-app/Prod\",\"start_time\":1.692291184114E9,\"trace_id\":\"1-64de5070-34ff53fe0d2e1fbe64b28bef\",\"end_time\":1.692291184602E9,\"http\":{\"request\":{\"url\":\"https://xxx.execute-api.us-east-1.amazonaws.com/Prod/order\",\"method\":\"POST\",\"user_agent\":\"Python/3.10 aiohttp/3.8.5\",\"client_ip\":\"a.b.c.d\",\"x_forwarded_for\":true},\"response\":{\"status\":200,\"content_length\":0}},\"aws\":{\"api_gateway\":{\"account_id\":\"123456789012\",\"rest_api_id\":\"xxx\",\"stage\":\"Prod\",\"request_id\":\"a8722fe5-a391-4c9f-bd89-024d8a6f595e\"}},\"annotations\":{\"aws:api_id\":\"xxx\",\"aws:api_stage\":\"Prod\"},\"metadata\":{\"default\":{\"extended_request_id\":\"J0GBkHklIAMF8qA=\",\"request_id\":\"a8722fe5-a391-4c9f-bd89-024d8a6f595e\"}},\"origin\":\"AWS::ApiGateway::Stage\",\"resource_arn\":\"arn:aws:apigateway:us-east-1::/restapis/xxx/stages/Prod\",\"subsegments\":[{\"id\":\"23a449852a482e1d\",\"name\":\"Lambda\",\"start_time\":1.69229118412E9,\"end_time\":1.692291184602E9,\"http\":{\"request\":{\"url\":\"https://lambda.us-east-1.amazonaws.com/2015-03-31/functions/arn:aws:lambda:us-east-1:123456789012:function:my-function/invocations\",\"method\":\"POST\"},\"response\":{\"status\":200,\"content_length\":155}},\"aws\":{\"function_name\":\"my-function\",\"region\":\"us-east-1\",\"operation\":\"Invoke\",\"resource_names\":[\"my-function\"]},\"namespace\":\"aws\"}]}",
			id:        "2b4e0ef3499cd781",
			traceId:   "1-64de5070-34ff53fe0d2e1fbe64b28bef",
			origin:    "AWS::ApiGateway::Stage",
			url:       "https://xxx.execute-api.us-east-1.amazonaws.com/Prod/order",
			startTime: 1.692291184114e+09,
		},
		{
			document:  "{\"id\":\"4174aefb49d05470\",\"name\":\"my-function\",\"start_time\":1.692291184757E9,\"trace_id\":\"1-64de5070-34ff53fe0d2e1fbe64b28bef\",\"end_time\":1.692291184772E9,\"parent_id\":\"28793e67ea7098e3\",\"http\":{\"response\":{\"status\":202}},\"aws\":{\"request_id\":\"63ea79ab-e338-4c12-a534-91a2dc43e2b0\"},\"origin\":\"AWS::Lambda\",\"resource_arn\":\"arn:aws:lambda:us-east-1:012345678901:function:my-function\",\"subsegments\":[{\"id\":\"7f27e935214d0df0\",\"name\":\"Dwell Time\",\"start_time\":1.692291184757E9,\"end_time\":1.692291184802E9},{\"id\":\"695ea167ab53c29f\",\"name\":\"Attempt #1\",\"start_time\":1.692291184802E9,\"end_time\":1.692291185307E9,\"http\":{\"response\":{\"status\":200}}}]}",
			id:        "4174aefb49d05470",
			traceId:   "1-64de5070-34ff53fe0d2e1fbe64b28bef",
			origin:    "AWS::Lambda",
			startTime: 1.692291184757e+09,
		},
		{
			document:  "{\"id\":\"224acacf31b69977\",\"name\":\"Events\",\"start_time\":1.692293788194E9,\"trace_id\":\"1-64de5a99-5d09aa705e56bbd0152548cb\",\"end_time\":1.692293788694E9,\"parent_id\":\"aabeb336dc5bce90\",\"inferred\":true,\"http\":{\"response\":{\"status\":200,\"content_length\":85}},\"aws\":{\"retries\":1,\"region\":\"us-east-1\",\"operation\":\"PutEvents\",\"request_id\":\"e957a46b-60da-47be-b4af-8b947507c9ab\"},\"origin\":\"AWS::Events\"}",
			id:        "224acacf31b69977",
			traceId:   "1-64de5a99-5d09aa705e56bbd0152548cb",
			origin:    "AWS::Events",
			startTime: 1.692293788194e+09,
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			segment, err := SegmentFromDocument(tt.document)
			assert.Nil(t, err)
			assert.Equal(t, tt.id, *segment.Id)
			assert.Equal(t, tt.traceId, *segment.TraceId)
			assert.Equal(t, tt.origin, *segment.Origin)
			assert.Equal(t, tt.startTime, *segment.StartTime)
			if tt.url != "" {
				assert.Equal(t, tt.url, *segment.Http.Request.Url)
			}
		})
	}
}

func TestSegmentFromDocumentFailure(t *testing.T) {
	cases := []struct {
		document string
		err      error
	}{
		{
			document: "",
			err:      errors.New("failed to decode segment document: unexpected end of JSON input"),
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			segment, err := SegmentFromDocument(tt.document)
			assert.Nil(t, segment)
			assert.EqualError(t, err, tt.err.Error())
		})
	}
}
