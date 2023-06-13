package publicrpc

import (
	"context"
	"fmt"
	"reflect"
	"zion/internal/pkg/aws/config"
	zioncfn "zion/internal/pkg/cloudformation"
	"zion/internal/pkg/public-rpc/types"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type GetStackOutputParams struct {
	StackName   string
	OutputNames []string
	Profile     string
	Region      string
}

func (p *GetStackOutputParams) RPCMethod() (*types.Result, error) {
	cfg, err := config.GetAWSConfig(context.TODO(), p.Region, p.Profile)

	if err != nil {
		return nil, fmt.Errorf("error when loading AWS config: %v", err)
	}

	cfnClient := cloudformation.NewFromConfig(cfg)

	mOutputKeys, err := zioncfn.GetStackOuput(p.StackName, p.OutputNames, cfnClient)

	// Fowards id and err to caller for handling
	return &types.Result{
		Output: mOutputKeys,
	}, err
}

func (p *GetStackOutputParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(zioncfn.GetStackOuput)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}
