package generatemockevent_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"zion/integration/zion"
	"zion/internal/pkg/jsonrpc"

	"github.com/aws/aws-sdk-go-v2/service/schemas"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const test_method = "generate_mock_event"

func TestGenerateMockEvent(t *testing.T) {
	region := "ap-southeast-1"
	// cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	// if err != nil {
	// 	t.Fatalf("failed to get aws config: %v", err)
	// }
	s := new(GenerateMockEventSuite)
	s.region = region

	suite.Run(t, s)
}

type GenerateMockEventSuite struct {
	suite.Suite

	region string

	schemasClient *schemas.Client
}

func (s *GenerateMockEventSuite) SetupSuite() {
	s.T().Log("setup suite start")
	s.T().Log("")
}

func (s *GenerateMockEventSuite) TestErrors() {
	cases := []struct {
		testname      string
		request       func() []byte
		expectErrCode int
		expectErrMsg  string
	}{
		{
			testname: "specified both RegistryName and SchemaFile",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "42",
					"method": %q,
					"params": {
						"RegistryName": "X",
						"SchemaFile": "path/to/my/schema",
						"Region": %q
					}
				}`, test_method, s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  `provide either "RegistryName and SchemaName" or "SchemaFile", not both`,
		},
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
			testname: "missing (RegistryName & SchemaName) and (SchemaFile)",
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
			expectErrMsg:  `missing either "RegistryName and SchemaName" or "SchemaFile"`,
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
