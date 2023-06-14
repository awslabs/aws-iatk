package publicrpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"

	"zion/internal/pkg/aws/config"
	"zion/internal/pkg/harness/eventbridge/listener"
	"zion/internal/pkg/harness/tags"
	"zion/internal/pkg/public-rpc/types"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tagtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
)

type RemoveEbListenersParams struct {
	IDs        []string `json:"Ids"`
	TagFilters []tagtypes.TagFilter
	Profile    string
	Region     string
}

func (p *RemoveEbListenersParams) RPCMethod() (*types.Result, error) {
	if p.IDs != nil && p.TagFilters != nil {
		return nil, errors.New("only one of Ids and TagFilters is needed, not both")
	}

	ctx := context.TODO()

	cfg, err := config.GetAWSConfig(ctx, p.Region, p.Profile)
	if err != nil {
		return nil, fmt.Errorf("error when loading AWS config: %v", err)
	}

	var listenerIDs []string
	if p.TagFilters != nil {
		listenerIDs, err = tags.GetTestHarnessIDsWithTagFilters(ctx, resourcegroupstaggingapi.NewFromConfig(cfg), p.TagFilters)
		if err != nil {
			return nil, fmt.Errorf("unable to find listeners with tag filters: %v", err)
		}
		log.Printf("found listener ids matching tag filters: %v", listenerIDs)
	} else {
		listenerIDs = p.IDs
	}

	err = listener.DestroyMultiple(ctx, listenerIDs, cfg, listener.NewDestroyOptions())

	if err != nil {
		return nil, err
	}

	return &types.Result{
		Output: "success",
	}, nil
}

func (p *RemoveEbListenersParams) ReflectOutput() reflect.Value {
	return reflect.New(reflect.TypeOf(""))
}
