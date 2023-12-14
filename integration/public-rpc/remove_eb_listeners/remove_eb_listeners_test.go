// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package destroytestingresources_test

import (
	"context"
	"encoding/json"
	"iatk/integration/iatk"
	"iatk/internal/pkg/harness/eventbridge/listener"
	"iatk/internal/pkg/jsonrpc"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const test_method = "test_harness.eventbridge.remove_listeners"

func TestRemoveListeners(t *testing.T) {
	s := &EventBusDestroyTestingResourcesSuite{
		eventBusName: uuid.NewString(),
		eventBusRule: "eb-testrule",
		region:       "us-west-2",
	}
	s.setAWSConfig()
	suite.Run(t, s)
}

type EventBusDestroyTestingResourcesSuite struct {
	suite.Suite
	eventBusName string
	eventBusRule string
	region       string
	cfg          aws.Config
}

func (s *EventBusDestroyTestingResourcesSuite) setAWSConfig() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(s.region))
	if err != nil {
		s.T().Fatalf("failed to get aws config: %v", err)
	}
	s.cfg = cfg
}

func (s *EventBusDestroyTestingResourcesSuite) SetupSuite() {
	ebClient := eventbridge.NewFromConfig(s.cfg)
	_, err := ebClient.CreateEventBus(context.TODO(), &eventbridge.CreateEventBusInput{
		Name: aws.String(s.eventBusName),
	})
	if err != nil {
		s.T().Fatalf("failed to create event bus: %v", err)
	}

	_, err = ebClient.PutRule(context.TODO(), &eventbridge.PutRuleInput{
		Name:         aws.String(s.eventBusRule),
		EventBusName: aws.String(s.eventBusName),
		EventPattern: aws.String("{\"detail-type\": [\"customerCreated\"], \"source\": [\"aws.events\"]}"),
	})
	s.Require().NoErrorf(err, "failed to create eventbridge rule: %v", err)
	s.T().Log("setup suite complete")
}

func (s *EventBusDestroyTestingResourcesSuite) TearDownSuite() {
	ebClient := eventbridge.NewFromConfig(s.cfg)

	_, err := ebClient.DeleteRule(context.TODO(), &eventbridge.DeleteRuleInput{
		Name:         aws.String(s.eventBusRule),
		EventBusName: aws.String(s.eventBusName),
	})
	s.Require().NoErrorf(err, "failed to delete rule %v", s.eventBusRule)

	deleteEventBus(s.T(), ebClient, s.eventBusName)
}

func (s *EventBusDestroyTestingResourcesSuite) TestRemoveListenersByResourceGroupIDs() {
	cases := map[string]struct {
		numResourceGroups    int
		createResourceGroups func(num int) []string
		input                func(resourceGroupIDs []string) []byte
		expectErr            bool
	}{
		"should sucessfully destroy": {
			expectErr:         false,
			numResourceGroups: 5,
			createResourceGroups: func(count int) []string {
				resourceGroupIDs := []string{}
				for i := 0; i < count; i++ {
					rgid := createTestingResources(s.T(), s.cfg, s.eventBusName, s.eventBusRule, s.region, nil)
					resourceGroupIDs = append(resourceGroupIDs, rgid)
				}
				return resourceGroupIDs
			},
			input: func(resourceGroupIDs []string) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  test_method,
					"params": map[string]interface{}{
						"Ids":    resourceGroupIDs,
						"Region": s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
		"should succeed with empty resource group ids in input": {
			expectErr:         false,
			numResourceGroups: 0,
			createResourceGroups: func(count int) []string {
				return []string{}
			},
			input: func(resourceGroupIDs []string) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  test_method,
					"params": map[string]interface{}{
						"Ids":    []string{},
						"Region": s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
		"should succeed with duplicated resource group ids in input": {
			expectErr:         false,
			numResourceGroups: 2,
			createResourceGroups: func(count int) []string {
				resourceGroupIDs := []string{}
				for i := 0; i < count; i++ {
					rgid := createTestingResources(s.T(), s.cfg, s.eventBusName, s.eventBusRule, s.region, nil)
					resourceGroupIDs = append(resourceGroupIDs, rgid)
				}
				return resourceGroupIDs
			},
			input: func(resourceGroupIDs []string) []byte {
				dups := []string{}
				dups = append(dups, resourceGroupIDs...)
				dups = append(dups, resourceGroupIDs...)
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  test_method,
					"params": map[string]interface{}{
						"Ids":    dups,
						"Region": s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
		"should fail with resource group id containing no resources": {
			expectErr:         true,
			numResourceGroups: 0,
			createResourceGroups: func(count int) []string {
				return []string{}
			},
			input: func(resourceGroupIDs []string) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  test_method,
					"params": map[string]interface{}{
						"Ids": []string{
							"eb-ffffffff-ffff-ffff-ffff-ffffffffffff",
						},
						"Region": s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
	}

	for name, tt := range cases {
		s.T().Run(name, func(t *testing.T) {
			resourceGroupIDs := tt.createResourceGroups(tt.numResourceGroups)

			input := tt.input(resourceGroupIDs)
			var out strings.Builder
			var sErr strings.Builder
			iatk.Invoke(t, input, &out, &sErr, nil)
			log.Printf("response: %v", out.String())
			var actual jsonrpc.Response
			json.Unmarshal([]byte(out.String()), &actual)
			if !tt.expectErr {
				output := actual.Result.(map[string]interface{})["output"].(string)
				assert.Equal(s.T(), "success", output)
			} else {
				assert.NotNil(t, actual.Error)
			}
			assertResourceGroupsAreDestoryed(s.T(), resourceGroupIDs, s.cfg)
		})
	}
}

func (s *EventBusDestroyTestingResourcesSuite) TestRemoveListenersByTagFilters() {
	cases := map[string]struct {
		createResourceGroups func() []string
		tagFilters           []types.TagFilter
		input                func(tagFilters []types.TagFilter) []byte
		expectErr            bool
	}{
		"should succeed": {
			expectErr: false,
			tagFilters: []types.TagFilter{
				{Key: aws.String("foo"), Values: []string{"bar"}},
			},
			createResourceGroups: func() []string {
				count := 5
				resourceGroupIDs := []string{}
				for i := 0; i < count; i++ {
					rgid := createTestingResources(s.T(), s.cfg, s.eventBusName, s.eventBusRule, s.region, map[string]string{
						"foo": "bar",
					})
					resourceGroupIDs = append(resourceGroupIDs, rgid)
				}
				return resourceGroupIDs
			},
			input: func(tagFilters []types.TagFilter) []byte {
				rJson := map[string]interface{}{
					"jsonrpc": "2.0",
					"id":      "42",
					"method":  test_method,
					"params": map[string]interface{}{
						"TagFilters": tagFilters,
						"Region":     s.region,
					},
				}
				out, _ := json.Marshal(rJson)
				return out
			},
		},
	}

	for name, tt := range cases {
		s.T().Run(name, func(t *testing.T) {
			resourceGroupIDs := tt.createResourceGroups()

			input := tt.input(tt.tagFilters)
			log.Printf("request: %v", string(input))
			var out strings.Builder
			var sErr strings.Builder
			iatk.Invoke(t, input, &out, &sErr, nil)
			log.Printf("response: %v", out.String())
			var actual jsonrpc.Response
			json.Unmarshal([]byte(out.String()), &actual)
			if !tt.expectErr {
				output := actual.Result.(map[string]interface{})["output"].(string)
				assert.Equal(s.T(), "success", output)
			} else {
				assert.NotNil(t, actual.Error)
			}
			assertResourceGroupsAreDestoryed(s.T(), resourceGroupIDs, s.cfg)
		})
	}
}

func createTestingResources(t *testing.T, cfg aws.Config, eventBusName, eventBusRule, region string, tags map[string]string) string {
	rJson := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "42",
		"method":  "test_harness.eventbridge.add_listener",
		"params": map[string]interface{}{
			"EventBusName": eventBusName,
			"RuleName":     eventBusRule,
			"Region":       region,
			"Tags":         tags,
		},
	}
	resquest, _ := json.Marshal(rJson)
	var out strings.Builder
	var sErr strings.Builder
	iatk.Invoke(t, resquest, &out, &sErr, nil)
	var response jsonrpc.Response
	json.Unmarshal([]byte(out.String()), &response)
	t.Log(response.Error)
	output := response.Result.(map[string]interface{})["output"].(map[string]interface{})
	return output["Id"].(string)
}

func deleteEventBus(t *testing.T, client *eventbridge.Client, eventBusName string) {
	_, err := client.DeleteEventBus(context.TODO(), &eventbridge.DeleteEventBusInput{
		Name: aws.String(eventBusName),
	})
	if err != nil {
		t.Fatalf("failed to delete event bus: %v", err)
	}
}

func assertResourceGroupsAreDestoryed(t *testing.T, resourceGroupIDs []string, cfg aws.Config) {
	for _, id := range resourceGroupIDs {
		_, err := listener.Get(context.TODO(), id, listener.NewOptions(cfg))
		assert.NotNil(t, err)
	}
}
