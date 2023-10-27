package publicrpc

import (
	"context"
	"ctk/internal/pkg/aws/config"
	ctkcfn "ctk/internal/pkg/cloudformation"
	"ctk/internal/pkg/jsonrpc"
	"ctk/internal/pkg/public-rpc/types"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type GetStackOutputParams struct {
	StackName   string
	OutputNames []string
	Profile     string
	Region      string
}

func (p *GetStackOutputParams) RPCMethod(metadata *jsonrpc.Metadata) (*types.Result, error) {
	cfg, err := config.GetAWSConfig(context.TODO(), p.Region, p.Profile, metadata)

	if err != nil {
		return nil, fmt.Errorf("error when loading AWS config: %v", err)
	}

	cfnClient := cloudformation.NewFromConfig(cfg)

	mOutputKeys, err := ctkcfn.GetStackOuput(p.StackName, p.OutputNames, cfnClient)

	// Fowards id and err to caller for handling
	return &types.Result{
		Output: mOutputKeys,
	}, err
}

func (p *GetStackOutputParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(ctkcfn.GetStackOuput)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}
