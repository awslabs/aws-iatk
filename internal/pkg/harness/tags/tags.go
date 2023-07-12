// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tags

import (
	"context"
	"fmt"

	"zion/internal/pkg/slice"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tagtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
)

type SystemTagKey string

const (
	TestHarnessID      SystemTagKey = "zion:TestHarness:ID"
	TestHarnessType    SystemTagKey = "zion:TestHarness:Type"
	TestHarnessTarget  SystemTagKey = "zion:TestHarness:Target"
	TestHarnessCreated SystemTagKey = "zion:TestHarness:Created"
)

// Validates if a given tags contains any reserved key
func ValidateTags(tags map[string]string) error {
	for _, key := range []SystemTagKey{TestHarnessID, TestHarnessType, TestHarnessTarget, TestHarnessCreated} {
		if _, ok := tags[string(key)]; ok {
			return fmt.Errorf("reserved tag key %q found in provided tags", key)
		}
	}
	return nil
}

// Get Target PhysicalID by TestHarnessID Tag Value
func GetTargetByTestHarnessID(ctx context.Context, api GetResourcesAPI, id string) (string, error) {
	paginator := resourcegroupstaggingapi.NewGetResourcesPaginator(api, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []tagtypes.TagFilter{
			{Key: aws.String(string(TestHarnessID)), Values: []string{id}},
		},
	})

	var resources []tagtypes.ResourceTagMapping

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get resources: %v", err)
		}

		resources = append(resources, output.ResourceTagMappingList...)
	}

	if len(resources) == 0 {
		return "", fmt.Errorf("no resource found for Test Harness %v", id)
	}

	physicalIDs := []string{}
	for _, r := range resources {
		for _, tag := range r.Tags {
			if aws.ToString(tag.Key) == string(TestHarnessTarget) {
				id := aws.ToString(tag.Value)
				physicalIDs = append(physicalIDs, id)
				break
			}
		}
	}
	distinct := slice.Dedup(physicalIDs)

	if len(distinct) > 1 {
		return "", fmt.Errorf("found multiple targets for Test Harness %v: %v", id, distinct)
	}

	if len(distinct) == 0 {
		return "", fmt.Errorf("found zero target for Test Harness %v", id)
	}

	return distinct[0], nil
}

// Find all resources with provided tag filters
func GetTestHarnessIDsWithTagFilters(ctx context.Context, api GetResourcesAPI, tagFilters []tagtypes.TagFilter) ([]string, error) {
	if !hasOneSystemTagKey(tagFilters) {
		// NOTE (hawflau): to make sure only zion-created resources will be found
		tagFilters = append(tagFilters, tagtypes.TagFilter{Key: aws.String(string(TestHarnessID))})
	}

	paginator := resourcegroupstaggingapi.NewGetResourcesPaginator(api, &resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: tagFilters,
	})

	ids := []string{}
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get resources: %v", err)
		}
		for _, r := range output.ResourceTagMappingList {
			ids = append(ids, extractTestHarnessIDFromTags(r.Tags)...)
		}
	}

	return slice.Dedup(ids), nil
}

//go:generate mockery --name GetResourcesAPI
type GetResourcesAPI interface {
	GetResources(context.Context, *resourcegroupstaggingapi.GetResourcesInput, ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error)
}

func hasOneSystemTagKey(l []tagtypes.TagFilter) bool {
	for _, f := range l {
		key := aws.ToString(f.Key)
		switch key {
		case string(TestHarnessID), string(TestHarnessType), string(TestHarnessTarget), string(TestHarnessCreated):
			return true
		}
	}
	return false
}

func extractTestHarnessIDFromTags(tags []tagtypes.Tag) []string {
	l := []string{}
	for _, tag := range tags {
		if aws.ToString(tag.Key) == string(TestHarnessID) {
			l = append(l, aws.ToString(tag.Value))
			break
		}
	}
	return l
}
