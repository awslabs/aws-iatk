package publicrpc

import (
	"context"
	"fmt"
	"reflect"

	"iatk/internal/pkg/aws/config"
	iatkcfn "iatk/internal/pkg/cloudformation"
	"iatk/internal/pkg/jsonrpc"
	"iatk/internal/pkg/public-rpc/types"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

type GetPhysicalIdParams struct {
	LogicalResourceId string `json:"LogicalResourceId"`

	StackName string `json:"StackName"`

	Profile string `json:"Profile,omitempty"`

	Region string `json:"Region,omitempty"`
}

func (p *GetPhysicalIdParams) RPCMethod(metadata *jsonrpc.Metadata) (*types.Result, error) {
	cfg, err := config.GetAWSConfig(context.TODO(), p.Region, p.Profile, metadata)

	if err != nil {
		return nil, fmt.Errorf("error when loading AWS config: %v", err)
	}

	cfnClient := cloudformation.NewFromConfig(cfg)
	id, err := iatkcfn.GetPhysicalId(p.StackName, p.LogicalResourceId, cfnClient)

	// Fowards id and err to caller for handling
	return &types.Result{
		Output: id,
	}, err
}

func (p *GetPhysicalIdParams) ReflectOutput() reflect.Value {
	ft := reflect.TypeOf(iatkcfn.GetPhysicalId)
	out0 := ft.Out(0)
	return reflect.New(out0).Elem()
}
