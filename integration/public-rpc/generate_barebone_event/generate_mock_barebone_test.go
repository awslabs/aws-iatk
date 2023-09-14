package generatebareboneevent_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	cfn "zion/integration/cloudformation"
	"zion/integration/zion"
	zioncfn "zion/internal/pkg/cloudformation"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/rs/xid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const test_method = "mock.generate_barebone_event"

func TestGenerateMockEvent(t *testing.T) {
	region := "us-east-1"

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		t.Fatalf("failed to get aws config: %v", err)
	}
	cfnClient := cloudformation.NewFromConfig(cfg)
	s := new(GenerateMockEventSuite)
	s.cfnClient = cfnClient
	s.stackName = "test-stack-" + xid.New().String()
	s.region = region
	suite.Run(t, s)
}

type GenerateMockEventSuite struct {
	suite.Suite

	stackName string
	region    string

	cfnClient *cloudformation.Client

	registry          string
	openapiName       string
	openapiVersion    string
	jsonschemaName    string
	jsonschemaVersion string
}

func (s *GenerateMockEventSuite) SetupSuite() {
	s.T().Log("setup suite start")
	err := cfn.Deploy(
		s.T(),
		s.cfnClient,
		s.stackName,
		"./test_stack.yaml",
		[]types.Capability{})
	s.Require().NoError(err, "failed to create stack")
	output, err := zioncfn.GetStackOuput(
		s.stackName,
		[]string{
			"TestSchemaRegistryName",
			"TestEBEventSchemaOpenAPIName",
			"TestEBEventSchemaOpenAPIVersion",
			"TestEBEventSchemaJSONSchemaName",
			"TestEBEventSchemaJSONSchemaVersion",
		},
		s.cfnClient,
	)
	s.Require().NoError(err, "failed to get stack outputs")
	s.registry = output["TestSchemaRegistryName"]
	s.openapiName = output["TestEBEventSchemaOpenAPIName"]
	s.openapiVersion = output["TestEBEventSchemaOpenAPIVersion"]
	s.jsonschemaName = output["TestEBEventSchemaJSONSchemaName"]
	s.jsonschemaVersion = output["TestEBEventSchemaJSONSchemaVersion"]
	s.Require().NotEmpty(s.registry)
	s.Require().NotEmpty(s.openapiName)
	s.Require().NotEmpty(s.openapiVersion)
	s.Require().NotEmpty(s.jsonschemaName)
	s.Require().NotEmpty(s.jsonschemaVersion)
	s.T().Log("setup suite complete")
}

func (s *GenerateMockEventSuite) TearDownSuite() {
	s.T().Log("teardown suite start")
	err := cfn.Destroy(s.T(), s.cfnClient, s.stackName)
	s.Require().NoError(err, "failed to destroy stack")
	s.T().Log("teardown suite complete")
}

func (s *GenerateMockEventSuite) TestGenerateMockEvent() {
	cases := []struct {
		testname string
		request  func() []byte
	}{
		{
			testname: "openapi, with version",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "43",
					"method": %q,
					"params": {
						"Region": %q,
						"RegistryName": %q,
						"SchemaName": %q,
						"SchemaVersion": %q,
						"EventRef": "MyEvent"
					}
				}`, test_method, s.region, s.registry, s.openapiName, s.openapiVersion))
			},
		},
	}

	for _, tt := range cases {
		s.T().Run(tt.testname, func(t *testing.T) {
			req := tt.request()
			res := s.invoke(req)
			require.Nilf(t, res.Error, "failed to generate mock event: %w", res.Error)
			require.NotNil(t, res.Result)
			_, ok := res.Result.(map[string]any)["output"].(string)
			require.True(t, ok)
		})
	}
}

func (s *GenerateMockEventSuite) TestInvalidParams() {
	cases := []struct {
		testname      string
		request       func() []byte
		expectErrCode int
		expectErrMsg  string
	}{
		{
			testname: "provided RegistryName but not SchemaName",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "42",
					"method": %q,
					"params": {
						"RegistryName": "something",
						"Region": %q
					}
				}`, test_method, s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  `requires both "RegistryName" and "SchemaName"`,
		},
		{
			testname: "missing both RegistryName and SchemaName",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "42",
					"method": %q,
					"params": {
						"Region": %q
					}
				}`, test_method, s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  `requires both "RegistryName" and "SchemaName"`,
		},
	}

	for _, tt := range cases {
		s.T().Run(tt.testname, func(t *testing.T) {
			req := tt.request()
			res := s.invoke(req)
			require.NotNil(t, res.Error)
			assert.Equal(t, tt.expectErrCode, res.Error.Code)
			assert.Contains(t, res.Error.Message, tt.expectErrMsg)
		})
	}
}

func (s *GenerateMockEventSuite) invoke(req []byte) jsonrpc.Response {
	var stdout strings.Builder
	var stderr strings.Builder
	test := s.T()
	test.Logf("request: %v", string(req))
	zion.Invoke(test, req, &stdout, &stderr, nil)

	test.Logf("response: %v", stdout.String())
	var res jsonrpc.Response
	err := json.Unmarshal([]byte(stdout.String()), &res)
	s.Require().NoError(err, "cannot unmarshal response")
	return res
}
