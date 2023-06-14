package tags

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	tagtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/stretchr/testify/assert"
)

func TestValidateTags(t *testing.T) {
	cases := map[string]struct {
		input     map[string]string
		expectErr error
	}{
		"should succeed": {
			input:     map[string]string{"foo": "bar", "hello": "world"},
			expectErr: nil,
		},
		"should return error if zion:TestHarness:ID is one of the provided keys": {
			input:     map[string]string{"zion:TestHarness:ID": "123"},
			expectErr: errors.New(`reserved tag key "zion:TestHarness:ID" found in provided tags`),
		},
		"should return error if zion:TestHarness:Type is one of the provided keys": {
			input:     map[string]string{"zion:TestHarness:Type": "123"},
			expectErr: errors.New(`reserved tag key "zion:TestHarness:Type" found in provided tags`),
		},
		"should return error if zion:TestHarness:Target is one of the provided keys": {
			input:     map[string]string{"zion:TestHarness:Target": "123"},
			expectErr: errors.New(`reserved tag key "zion:TestHarness:Target" found in provided tags`),
		},
		"should return error if zion:TestHarness:Created is one of the provided keys": {
			input:     map[string]string{"zion:TestHarness:Created": "123"},
			expectErr: errors.New(`reserved tag key "zion:TestHarness:Created" found in provided tags`),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			err := ValidateTags(tt.input)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestGetTargetByTestHarnessID(t *testing.T) {
	cases := map[string]struct {
		mockGetResourcesAPI func(ctx context.Context, testHarnessID string) *MockGetResourcesAPI
		testHarnessID       string
		expect              string
		expectErr           error
	}{
		"success": {
			testHarnessID: "some-id",
			mockGetResourcesAPI: func(ctx context.Context, testHarnessID string) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: []tagtypes.TagFilter{
							{Key: aws.String(string(TestHarnessID)), Values: []string{testHarnessID}},
						},
					}).
					Return(&resourcegroupstaggingapi.GetResourcesOutput{
						ResourceTagMappingList: []tagtypes.ResourceTagMapping{
							{ResourceARN: aws.String("arn1"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessTarget)), Value: aws.String("my-event-bus")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
							{ResourceARN: aws.String("arn2"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessTarget)), Value: aws.String("my-event-bus")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
							{ResourceARN: aws.String("arn3"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessTarget)), Value: aws.String("my-event-bus")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
						},
					}, nil)
				return api
			},
			expect:    "my-event-bus",
			expectErr: nil,
		},
		"api failed": {
			testHarnessID: "some-id",
			mockGetResourcesAPI: func(ctx context.Context, testHarnessID string) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: []tagtypes.TagFilter{
							{Key: aws.String(string(TestHarnessID)), Values: []string{testHarnessID}},
						},
					}).
					Return(nil, errors.New("api failed"))
				return api
			},
			expect:    "",
			expectErr: errors.New("failed to get resources: api failed"),
		},
		"no resource found from api": {
			testHarnessID: "some-id",
			mockGetResourcesAPI: func(ctx context.Context, testHarnessID string) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: []tagtypes.TagFilter{
							{Key: aws.String(string(TestHarnessID)), Values: []string{testHarnessID}},
						},
					}).
					Return(&resourcegroupstaggingapi.GetResourcesOutput{
						ResourceTagMappingList: []tagtypes.ResourceTagMapping{},
					}, nil)
				return api
			},
			expect:    "",
			expectErr: errors.New("no resource found for Test Harness some-id"),
		},
		"more than one target found": {
			testHarnessID: "some-id",
			mockGetResourcesAPI: func(ctx context.Context, testHarnessID string) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: []tagtypes.TagFilter{
							{Key: aws.String(string(TestHarnessID)), Values: []string{testHarnessID}},
						},
					}).
					Return(&resourcegroupstaggingapi.GetResourcesOutput{
						ResourceTagMappingList: []tagtypes.ResourceTagMapping{
							{ResourceARN: aws.String("arn1"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessTarget)), Value: aws.String("my-event-bus")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
							{ResourceARN: aws.String("arn2"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessTarget)), Value: aws.String("my-event-bus-2")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
							{ResourceARN: aws.String("arn3"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessTarget)), Value: aws.String("my-event-bus")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
						},
					}, nil)
				return api
			},
			expect:    "",
			expectErr: errors.New("found multiple targets for Test Harness some-id: [my-event-bus my-event-bus-2]"),
		},
		"no target found": {
			testHarnessID: "some-id",
			mockGetResourcesAPI: func(ctx context.Context, testHarnessID string) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: []tagtypes.TagFilter{
							{Key: aws.String(string(TestHarnessID)), Values: []string{testHarnessID}},
						},
					}).
					Return(&resourcegroupstaggingapi.GetResourcesOutput{
						ResourceTagMappingList: []tagtypes.ResourceTagMapping{
							{ResourceARN: aws.String("arn1"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
							{ResourceARN: aws.String("arn2"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
							{ResourceARN: aws.String("arn3"), Tags: []tagtypes.Tag{
								{Key: aws.String(string(TestHarnessID)), Value: aws.String("some-id")},
								{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
							}},
						},
					}, nil)
				return api
			},
			expect:    "",
			expectErr: errors.New("found zero target for Test Harness some-id"),
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			mockAPI := tt.mockGetResourcesAPI(context.TODO(), tt.testHarnessID)
			target, err := GetTargetByTestHarnessID(context.TODO(), mockAPI, tt.testHarnessID)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, target, tt.expect)
			}
		})
	}
}

func TestGetTestHarnessIDsWithTagFilters(t *testing.T) {
	cases := map[string]struct {
		mockGetResourcesAPI func(ctx context.Context, resourceGroupIDs []string, tagFilters []tagtypes.TagFilter) *MockGetResourcesAPI
		expect              []string
		tagFilters          []tagtypes.TagFilter
		expectErr           error
	}{
		"success": {
			expect: []string{
				"test-harness-id-1",
				"test-harness-id-2",
				"test-harness-id-3",
			},
			expectErr: nil,
			tagFilters: []tagtypes.TagFilter{
				{Key: aws.String("key1"), Values: []string{"Val1"}},
				{Key: aws.String("key2"), Values: []string{"Val1", "Val2"}},
			},
			mockGetResourcesAPI: func(ctx context.Context, resourceGroupIDs []string, tagFilters []tagtypes.TagFilter) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				tagFilters = append(tagFilters, tagtypes.TagFilter{Key: aws.String(string(TestHarnessID))})
				expectResources := []tagtypes.ResourceTagMapping{}
				for i, id := range resourceGroupIDs {
					expectResources = append(expectResources, tagtypes.ResourceTagMapping{
						ResourceARN: aws.String("arn" + fmt.Sprint(i)),
						Tags: []tagtypes.Tag{
							{Key: aws.String(string(TestHarnessID)), Value: aws.String(id)},
							{Key: aws.String(string(TestHarnessTarget)), Value: aws.String("my-event-bus")},
							{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
						},
					})
				}
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: tagFilters,
					}).
					Return(&resourcegroupstaggingapi.GetResourcesOutput{
						ResourceTagMappingList: expectResources,
					}, nil)
				return api
			},
		},
		"already provided system key (zion:TestHarnessID) in tag filter; success": {
			expect: []string{
				"test-harness-id-1",
				"test-harness-id-2",
				"test-harness-id-3",
			},
			expectErr: nil,
			tagFilters: []tagtypes.TagFilter{
				{Key: aws.String("key1"), Values: []string{"Val1"}},
				{Key: aws.String("key2"), Values: []string{"Val1", "Val2"}},
				{Key: aws.String(string(TestHarnessType)), Values: []string{"EventBridge.Listener"}},
			},
			mockGetResourcesAPI: func(ctx context.Context, resourceGroupIDs []string, tagFilters []tagtypes.TagFilter) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				expectResources := []tagtypes.ResourceTagMapping{}
				for i, id := range resourceGroupIDs {
					expectResources = append(expectResources, tagtypes.ResourceTagMapping{
						ResourceARN: aws.String("arn" + fmt.Sprint(i)),
						Tags: []tagtypes.Tag{
							{Key: aws.String(string(TestHarnessID)), Value: aws.String(id)},
							{Key: aws.String(string(TestHarnessTarget)), Value: aws.String("my-event-bus")},
							{Key: aws.String(string(TestHarnessType)), Value: aws.String("EventBridge.Listener")},
						},
					})
				}
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: tagFilters,
					}).
					Return(&resourcegroupstaggingapi.GetResourcesOutput{
						ResourceTagMappingList: expectResources,
					}, nil)
				return api
			},
		},
		"api failed": {
			expect:    nil,
			expectErr: errors.New("failed to get resources: api failed"),
			tagFilters: []tagtypes.TagFilter{
				{Key: aws.String("key1"), Values: []string{"Val1"}},
				{Key: aws.String("key2"), Values: []string{"Val1", "Val2"}},
				{Key: aws.String(string(TestHarnessType)), Values: []string{"EventBridge.Listener"}},
			},
			mockGetResourcesAPI: func(ctx context.Context, resourceGroupIDs []string, tagFilters []tagtypes.TagFilter) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: tagFilters,
					}).
					Return(nil, errors.New("api failed"))
				return api
			},
		},
		"no resource found from api": {
			expect:    []string{},
			expectErr: nil,
			tagFilters: []tagtypes.TagFilter{
				{Key: aws.String("key1"), Values: []string{"Val1"}},
				{Key: aws.String("key2"), Values: []string{"Val1", "Val2"}},
				{Key: aws.String(string(TestHarnessType)), Values: []string{"EventBridge.Listener"}},
			},
			mockGetResourcesAPI: func(ctx context.Context, resourceGroupIDs []string, tagFilters []tagtypes.TagFilter) *MockGetResourcesAPI {
				api := NewMockGetResourcesAPI(t)
				expectResources := []tagtypes.ResourceTagMapping{}
				api.EXPECT().
					GetResources(ctx, &resourcegroupstaggingapi.GetResourcesInput{
						TagFilters: tagFilters,
					}).
					Return(&resourcegroupstaggingapi.GetResourcesOutput{
						ResourceTagMappingList: expectResources,
					}, nil)
				return api
			},
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()
			mockAPI := tt.mockGetResourcesAPI(ctx, tt.expect, tt.tagFilters)
			actual, err := GetTestHarnessIDsWithTagFilters(ctx, mockAPI, tt.tagFilters)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.Equal(t, tt.expect, actual)
			}
		})
	}

}

func Test_hasOneSystemTagKey(t *testing.T) {
	cases := []struct {
		input  []tagtypes.TagFilter
		expect bool
	}{
		{
			input: []tagtypes.TagFilter{
				{Key: aws.String("key1"), Values: []string{"val1", "val2"}},
			},
			expect: false,
		},
		{
			input: []tagtypes.TagFilter{
				{Key: aws.String("zion:TestHarness:ID"), Values: []string{"val1", "val2"}},
			},
			expect: true,
		},
		{
			input: []tagtypes.TagFilter{
				{Key: aws.String("zion:TestHarness:Target"), Values: []string{"val1", "val2"}},
			},
			expect: true,
		},
		{
			input: []tagtypes.TagFilter{
				{Key: aws.String("zion:TestHarness:Type"), Values: []string{"val1", "val2"}},
			},
			expect: true,
		},
		{
			input: []tagtypes.TagFilter{
				{Key: aws.String("zion:TestHarness:Created"), Values: []string{"val1", "val2"}},
			},
			expect: true,
		},
		{
			input: []tagtypes.TagFilter{
				{Key: aws.String("zion:TestHarness:ID"), Values: []string{"val1", "val2"}},
				{Key: aws.String("zion:TestHarness:Target"), Values: []string{"val1", "val2"}},
				{Key: aws.String("zion:TestHarness:Type"), Values: []string{"val1", "val2"}},
				{Key: aws.String("zion:TestHarness:Created"), Values: []string{"val1", "val2"}},
			},
			expect: true,
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			actual := hasOneSystemTagKey(tt.input)
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func Test_extractTestHarnessIDFromTags(t *testing.T) {
	cases := []struct {
		input  []tagtypes.Tag
		expect []string
	}{
		{
			input: []tagtypes.Tag{
				{Key: aws.String("key1"), Value: aws.String("val1")},
				{Key: aws.String("key2"), Value: aws.String("val2")},
				{Key: aws.String("zion:TestHarness:ID"), Value: aws.String("test-harness-id-1")},
			},
			expect: []string{"test-harness-id-1"},
		},
		{
			input: []tagtypes.Tag{
				{Key: aws.String("key1"), Value: aws.String("val1")},
				{Key: aws.String("key2"), Value: aws.String("val2")},
			},
			expect: []string{},
		},
	}

	for i, tt := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			actual := extractTestHarnessIDFromTags(tt.input)
			assert.Equal(t, tt.expect, actual)
		})
	}
}
