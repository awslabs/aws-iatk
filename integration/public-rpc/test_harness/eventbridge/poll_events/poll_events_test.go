// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pollevents_test

import (
	"context"
	"encoding/json"
	"fmt"
	"iatk/integration/iatk"
	"iatk/internal/pkg/jsonrpc"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	ebtypes "github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/rs/xid"
	"github.com/stretchr/testify/suite"
)

const poll_events_method string = "test_harness.eventbridge.poll_events"
const add_listener_method string = "test_harness.eventbridge.add_listener"
const remove_listeners_method string = "test_harness.eventbridge.remove_listeners"

func TestPollEvents(t *testing.T) {
	region := "us-west-2"
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		t.Fatalf("failed to get aws config: %v", err)
	}
	ebClient := eventbridge.NewFromConfig(cfg)
	snsClient := sns.NewFromConfig(cfg)

	s := new(PollEventsSuite)
	s.ebClient = ebClient
	s.snsClient = snsClient
	s.eventBusName = "eb-" + xid.New().String()
	s.snsTopicName = "zsns" + xid.New().String()
	s.region = region
	s.listenerParams = []EbConfiguration{
		{TargetId: "ebtn-" + xid.New().String(), RuleName: "ebrn-" + xid.New().String(), EventPattern: `{"source":[{"prefix":"com.test.0"}]}`},
		{TargetId: "ebtn-" + xid.New().String(), RuleName: "ebrn-" + xid.New().String(), EventPattern: `{"source":[{"prefix":"com.test.1"}]}`, TargetInput: aws.String(`"hello, world!"`)},
		{TargetId: "ebtn-" + xid.New().String(), RuleName: "ebrn-" + xid.New().String(), EventPattern: `{"source":[{"prefix":"com.test.2"}]}`, TargetInputPath: aws.String(`$.detail-type`)},
		{TargetId: "ebtn-" + xid.New().String(), RuleName: "ebrn-" + xid.New().String(), EventPattern: `{"source":[{"prefix":"com.test.3"}]}`, TargetInputTemplate: aws.String(`{"source": "<source>", "foo": "<foo>"}`), TargetInputPathsMaps: map[string]string{
			"foo":    "$.detail.foo",
			"source": "$.source",
		}},
	}

	suite.Run(t, s)

}

type EbConfiguration struct {
	TargetInput          *string
	TargetInputPath      *string
	TargetInputTemplate  *string
	TargetInputPathsMaps map[string]string
	EventPattern         string
	TargetId             string
	RuleName             string
}

type PollEventsSuite struct {
	suite.Suite

	eventBusName   string
	snsTopicName   string
	snsTopicArn    *string
	listenerParams []EbConfiguration
	listenerIDs    []string
	region         string

	ebClient  *eventbridge.Client
	snsClient *sns.Client
}

func (s *PollEventsSuite) SetupSuite() {
	s.T().Log("setup suite start")

	topic, err := s.snsClient.CreateTopic(context.TODO(), &sns.CreateTopicInput{
		Name: aws.String(s.snsTopicName),
	})
	s.Require().NoErrorf(err, "failed to create sns topic: %v", err)
	s.snsTopicArn = topic.TopicArn

	_, err = s.ebClient.CreateEventBus(context.TODO(), &eventbridge.CreateEventBusInput{
		Name: aws.String(s.eventBusName),
	})
	s.Require().NoErrorf(err, "failed to create event bus: %v", err)

	for _, p := range s.listenerParams {
		_, err = s.ebClient.PutRule(context.TODO(), &eventbridge.PutRuleInput{
			Name:         aws.String(p.RuleName),
			EventBusName: aws.String(s.eventBusName),
			EventPattern: aws.String(p.EventPattern),
		})
		s.Require().NoErrorf(err, "failed to create eventbridge rule: %v", err)

		target := ebtypes.Target{
			Id:  aws.String(p.TargetId),
			Arn: topic.TopicArn,
		}

		if p.TargetInput != nil {
			target.Input = p.TargetInput
		}

		if p.TargetInputPath != nil {
			target.InputPath = p.TargetInputPath
		}

		if p.TargetInputPathsMaps != nil && p.TargetInputTemplate != nil {
			target.InputTransformer = &ebtypes.InputTransformer{
				InputPathsMap: p.TargetInputPathsMaps,
				InputTemplate: p.TargetInputTemplate,
			}
		}

		_, err = s.ebClient.PutTargets(context.TODO(), &eventbridge.PutTargetsInput{
			EventBusName: aws.String(s.eventBusName),
			Rule:         aws.String(p.RuleName),
			Targets:      []ebtypes.Target{target},
		})

		s.Require().NoErrorf(err, "failed to create eventbridge target: %v", err)

		id := s.addListener(map[string]any{"TargetId": p.TargetId, "RuleName": p.RuleName})
		s.listenerIDs = append(s.listenerIDs, id)
	}

	s.T().Log("setup suite complete")
}

func (s *PollEventsSuite) TearDownSuite() {
	s.T().Log("tear down suite start")
	if len(s.listenerIDs) > 0 {
		s.removeListeners()
	}

	for _, p := range s.listenerParams {
		_, err := s.ebClient.RemoveTargets(context.TODO(), &eventbridge.RemoveTargetsInput{
			Ids:          []string{p.TargetId},
			Rule:         aws.String(p.RuleName),
			EventBusName: aws.String(s.eventBusName),
		})

		s.Require().NoErrorf(err, "failed to delete target %v", p.TargetId)

		_, err = s.ebClient.DeleteRule(context.TODO(), &eventbridge.DeleteRuleInput{
			Name:         aws.String(p.RuleName),
			EventBusName: aws.String(s.eventBusName),
		})

		s.Require().NoErrorf(err, "failed to delete rule %v", p.RuleName)
	}

	err := deleteEventBus(s.ebClient, s.eventBusName)
	s.Require().NoErrorf(err, "failed to delete bus %v", s.eventBusName)

	_, err = s.snsClient.DeleteTopic(context.TODO(), &sns.DeleteTopicInput{
		TopicArn: s.snsTopicArn,
	})
	s.Require().NoErrorf(err, "failed to delete sns %v", s.snsTopicArn)
	s.T().Log("tear down suite complete")
}

func (s *PollEventsSuite) TestReceiveMatchingEvents() {
	cases := []struct {
		testname            string
		listenerIdx         int
		pollWaitTimeSeconds *int32
		maxNumMessages      *int32
		events              func() []ebEvent
		matchFunc           func(r string) bool
		expectCount         int
	}{
		{
			testname:    "no input transformation",
			listenerIdx: 0,
			events: func() []ebEvent {
				return []ebEvent{
					{Source: "com.test.0", DetailType: "foo", Detail: map[string]string{"id": "0", "abc": "def"}},
					{Source: "com.test.0", DetailType: "foo", Detail: map[string]string{"id": "1", "abc": "def"}},
					{Source: "com.test.0", DetailType: "foo", Detail: map[string]string{"id": "2", "abc": "def"}},
					{Source: "com.null", DetailType: "foo", Detail: map[string]string{"id": "3", "abc": "def"}}, // not captured by listener
				}

			},
			matchFunc: func(r string) bool {
				var payload interface{}
				err := json.Unmarshal([]byte(r), &payload)
				s.Require().NoErrorf(err, "unexpected recevied payload: %v", r)

				s.T().Log(payload)
				detail := payload.(map[string]interface{})["detail"].(map[string]interface{})
				detailType := payload.(map[string]interface{})["detail-type"].(string)
				source := payload.(map[string]interface{})["source"].(string)
				return detail["abc"].(string) == "def" && detailType == "foo" && source == "com.test.0"
			},
			pollWaitTimeSeconds: aws.Int32(3),
			maxNumMessages:      aws.Int32(1),
			expectCount:         3,
		},
		{
			testname:    "input transformation via Input",
			listenerIdx: 1,
			events: func() []ebEvent {
				events := make([]ebEvent, 0, 10)
				for i := 0; i < 5; i++ {
					events = append(events, ebEvent{Source: "com.test.1", DetailType: "foo", Detail: map[string]string{"abc": "def"}})
				}
				for i := 0; i < 5; i++ {
					// not captured by listener
					events = append(events, ebEvent{Source: "com.null", DetailType: "xxx", Detail: map[string]string{"x": "y"}})
				}
				return events
			},
			matchFunc: func(r string) bool {
				return r == `"hello, world!"`
			},
			pollWaitTimeSeconds: aws.Int32(0),
			maxNumMessages:      aws.Int32(1),
			expectCount:         5,
		},
		{
			testname:    "input transformation via InputPath",
			listenerIdx: 2,
			events: func() []ebEvent {
				events := make([]ebEvent, 0, 8)
				for i := 0; i < 3; i++ {
					events = append(events, ebEvent{Source: "com.test.2", DetailType: "xyz", Detail: map[string]string{"abc": "def"}})
				}
				for i := 0; i < 5; i++ {
					// not captured by listener
					events = append(events, ebEvent{Source: "com.null", DetailType: "xxx", Detail: map[string]string{"x": "y"}})
				}
				return events
			},
			matchFunc: func(r string) bool {
				return r == `"xyz"`
			},
			pollWaitTimeSeconds: aws.Int32(0),
			maxNumMessages:      aws.Int32(10),
			expectCount:         3,
		},
		{
			testname:    "input transformation via InputTransformer",
			listenerIdx: 3,
			events: func() []ebEvent {
				events := make([]ebEvent, 0, 6)
				for i := 0; i < 4; i++ {
					events = append(events, ebEvent{Source: "com.test.3", DetailType: "mydetailtype", Detail: map[string]string{"foo": "bar"}})
				}
				for i := 0; i < 2; i++ {
					// not captured by listener
					events = append(events, ebEvent{Source: "com.null", DetailType: "xxx", Detail: map[string]string{"x": "y"}})
				}
				return events
			},
			matchFunc: func(r string) bool {
				return r == `{"source": "com.test.3", "foo": "bar"}`
			},
			pollWaitTimeSeconds: aws.Int32(2),
			maxNumMessages:      aws.Int32(10),
			expectCount:         4,
		},
		{
			testname:    "default wait time and max number of messages",
			listenerIdx: 0,
			events: func() []ebEvent {
				return []ebEvent{
					{Source: "com.test.0", DetailType: "foo", Detail: map[string]string{"id": "0", "abc": "def"}},
					{Source: "com.test.0", DetailType: "foo", Detail: map[string]string{"id": "1", "abc": "def"}},
					{Source: "com.test.0", DetailType: "foo", Detail: map[string]string{"id": "2", "abc": "def"}},
					{Source: "com.test.0", DetailType: "foo", Detail: map[string]string{"id": "3", "abc": "def"}},
					{Source: "com.null", DetailType: "foo", Detail: map[string]string{"id": "3", "abc": "def"}}, // not captured by listener
				}

			},
			matchFunc: func(r string) bool {
				var payload interface{}
				err := json.Unmarshal([]byte(r), &payload)
				s.Require().NoErrorf(err, "unexpected recevied payload: %v", r)

				s.T().Log(payload)
				detail := payload.(map[string]interface{})["detail"].(map[string]interface{})
				detailType := payload.(map[string]interface{})["detail-type"].(string)
				source := payload.(map[string]interface{})["source"].(string)
				return detail["abc"].(string) == "def" && detailType == "foo" && source == "com.test.0"
			},
			expectCount: 4,
		},
	}

	for _, tt := range cases {
		s.Run(tt.testname, func() {
			s.T().Log("sending events")
			err := s.sendEvents(tt.events())
			s.Require().NoError(err, "failed to send events")
			s.T().Log("sleep for 3s")
			time.Sleep(3 * time.Second)

			s.T().Log("polling events")
			listenerID := s.listenerIDs[tt.listenerIdx]
			received := make([]string, 0, len(tt.events()))
			for i := 0; i < len(tt.events()); i++ {
				// NOTE: make repeated calls since number of events is small (<1,000)
				// see https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_ReceiveMessage.html
				events := s.invokeAndAssertPollEventsRPC(listenerID, tt.pollWaitTimeSeconds, tt.maxNumMessages)
				received = append(received, events...)
			}
			s.Len(received, tt.expectCount)

			for _, r := range received {
				s.True(tt.matchFunc(r))
			}
		})
	}
}

func (s *PollEventsSuite) TestErrors() {
	cases := []struct {
		testname      string
		request       func() []byte
		expectErrCode int
		expectErrMsg  string
	}{
		{
			testname: "missing listener id",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "42",
					"method": %q,
					"params": {
						"Region": %q
					}
				}`, poll_events_method, s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  `missing required param "ListenerId"`,
		},
		{
			testname: "invalid wait time",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "42",
					"method": %q,
					"params": {
						"ListenerId": %q,
						"Region": %q,
						"WaitTimeSeconds": 200
					}
				}`, poll_events_method, s.listenerIDs[0], s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  `"WaitTimeSeconds" must be an integer between 0 and 20`,
		},
		{
			testname: "invalid max number of messages",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "42",
					"method": %q,
					"params": {
						"ListenerId": %q,
						"Region": %q,
						"MaxNumberOfMessages": -200
					}
				}`, poll_events_method, s.listenerIDs[0], s.region))
			},
			expectErrCode: 10,
			expectErrMsg:  `"MaxNumberOfMessages" must be an integer between 1 and 10`,
		},
		{
			testname: "max num messages overflow",
			request: func() []byte {
				return []byte(fmt.Sprintf(`
				{
					"jsonrpc": "2.0",
					"id": "42",
					"method": %q,
					"params": {
						"ListenerId": %q,
						"Region": %q,
						"MaxNumberOfMessages": 90000000000
					}
				}`, poll_events_method, s.listenerIDs[0], s.region))
			},
			expectErrCode: -32602,
			expectErrMsg:  "Invalid params",
		},
	}

	for _, tt := range cases {
		s.Run(tt.testname, func() {
			req := tt.request()
			res := s.invoke(req)
			s.Require().NotNil(res.Error)
			s.Equal(tt.expectErrCode, res.Error.Code)
			s.Contains(tt.expectErrMsg, res.Error.Message)
		})
	}
}

func (s *PollEventsSuite) invokeAndAssertPollEventsRPC(listenerID string, waitTimeSeconds, maxNumberOfMessages *int32) []string {
	params := map[string]any{
		"ListenerId": listenerID,
		"Region":     s.region,
	}
	if waitTimeSeconds != nil {
		params["WaitTimeSeconds"] = *waitTimeSeconds
	}
	if maxNumberOfMessages != nil {
		params["MaxNumberOfMessages"] = *maxNumberOfMessages
	}

	reqMap := map[string]any{
		"jsonrpc": "2.0",
		"id":      "42",
		"method":  poll_events_method,
		"params":  params,
	}
	req, _ := json.Marshal(reqMap)
	s.T().Logf("poll request: %v", reqMap)
	start := time.Now()
	res := s.invoke([]byte(req))
	s.T().Logf("elasped: %vms", time.Since(start).Milliseconds())
	s.Require().Nilf(res.Error, "failed to poll events: %v", res.Error)
	s.Require().NotNil(res.Result)
	output, ok := res.Result.(map[string]any)["output"].([]any)
	s.Require().True(ok, "output of poll_events must be a slice")
	events := make([]string, 0, len(output))
	for _, o := range output {
		event, ok := o.(string)
		s.Require().True(ok, "item of output must be a string")
		events = append(events, event)
	}
	return events
}

func (s *PollEventsSuite) invoke(req []byte) jsonrpc.Response {
	var stdout strings.Builder
	var stderr strings.Builder
	test := s.T()
	test.Logf("request: %v", string(req))
	iatk.Invoke(test, req, &stdout, &stderr, nil)

	test.Logf("response: %v", stdout.String())
	var res jsonrpc.Response
	err := json.Unmarshal([]byte(stdout.String()), &res)
	s.Require().NoError(err, "cannot unmarshal response")
	return res
}

func (s *PollEventsSuite) addListener(params map[string]any) string {
	params["EventBusName"] = s.eventBusName
	params["Region"] = s.region
	reqMap := map[string]any{
		"jsonrpc": "2.0",
		"id":      "42",
		"method":  add_listener_method,
		"params":  params,
	}
	req, _ := json.Marshal(reqMap)
	res := s.invoke([]byte(req))
	s.Require().Nilf(res.Error, "failed to add listener: %v", res.Error)
	output := res.Result.(map[string]any)["output"].(map[string]any)
	id := output["Id"].(string)
	return id
}

func (s *PollEventsSuite) removeListeners() {
	reqMap := map[string]any{
		"jsonrpc": "2.0",
		"id":      "42",
		"method":  remove_listeners_method,
		"params": map[string]any{
			"Ids":    s.listenerIDs,
			"Region": s.region,
		},
	}
	req, _ := json.Marshal(reqMap)
	res := s.invoke([]byte(req))
	s.Require().Nilf(res.Error, "failed to remove listeners: %v", res.Error)
}

func (s *PollEventsSuite) sendEvents(events []ebEvent) error {
	entries := make([]ebtypes.PutEventsRequestEntry, 0, len(events))
	for _, e := range events {
		entries = append(entries, s.ebEventToEntry(e))
	}

	_, err := s.ebClient.PutEvents(context.TODO(), &eventbridge.PutEventsInput{
		Entries: entries,
	})
	s.NoError(err, "failed to send event")
	return err
}

func (s *PollEventsSuite) ebEventToEntry(event ebEvent) ebtypes.PutEventsRequestEntry {
	detailJSON, _ := json.Marshal(event.Detail)
	return ebtypes.PutEventsRequestEntry{
		EventBusName: aws.String(s.eventBusName),
		Source:       aws.String(event.Source),
		DetailType:   aws.String(event.DetailType),
		Detail:       aws.String(string(detailJSON)),
	}
}

func deleteEventBus(client *eventbridge.Client, eventBusName string) error {
	_, err := client.DeleteEventBus(context.TODO(), &eventbridge.DeleteEventBusInput{
		Name: aws.String(eventBusName),
	})
	return err
}

type ebEvent struct {
	Source     string            `json:"source"`
	DetailType string            `json:"detail-type"`
	Detail     map[string]string `json:"detail"`
}
