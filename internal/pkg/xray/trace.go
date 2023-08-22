package xray

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/xray"
)

type BatchGetTracesAPI interface {
	BatchGetTraces(context.Context, *xray.BatchGetTracesInput, ...func(*xray.Options)) (*xray.BatchGetTracesOutput, error)
}
