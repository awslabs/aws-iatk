# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Integration tests for aws_ctk.wait_until_event_matched
"""
import json
import logging
from uuid import uuid4
from unittest import TestCase
from typing import List, Dict, Callable
from dataclasses import dataclass

import boto3
from botocore.exceptions import ClientError
from parameterized import parameterized

from aws_ctk import AWSCtk
from aws_ctk.poll_events import InvalidParamException


LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="aws_ctk", level=logging.DEBUG)


@dataclass
class EbEvent:
    source: str
    detail_type: str
    detail: Dict[str, str]

    def to_entry(self, event_bus_name: str):
        return {
            "EventBusName": event_bus_name,
            "Source": self.source,
            "DetailType": self.detail_type,
            "Detail": json.dumps(self.detail),
        }


def condition_func_0(received: str) -> bool:
    LOG.debug("received: %s", received)
    payload = json.loads(received)
    try:
        detail = payload["detail"]
        detail_type = payload["detail-type"]
        source = payload["source"]
        matched = (
            detail["abc"] == "def"
            and detail["id"] == "2"
            and source == "com.test.0"
            and detail_type == "foo"
        )
        LOG.debug("matched: %s", matched)
        return matched
    except Exception as e:
        LOG.debug("error: %s", e)
        return False


def condition_func_1(received: str) -> bool:
    LOG.debug("received: %s", received)
    return received == '"hello, world!"'


def condition_func_2(received: str) -> bool:
    LOG.debug("received: %s", received)
    return received == '"xyz"'


def condition_func_3(received: str) -> bool:
    LOG.debug("received: %s", received)
    return received == '{"source": "com.test.3", "foo": "bar"}'

@dataclass
class EbConfiguration:
    rule_name: str
    target_id: str
    event_bus_name: str
    event_pattern: str
    input_template: str = None
    input_paths_map: Dict = None
    input: str = None
    input_path: str = None


class TestCTK_wait_until_event_matched(TestCase):
    ctk = AWSCtk()
    eb_client = boto3.client("events")
    sqs_client = boto3.client("sqs")
    sns_client = boto3.client("sns")
    event_bus_name: str = f"eb-{uuid4()}"
    sns_topic_name: str = f"zsns-{uuid4()}"
    sns_topic_arn: str = None
    listener_ids: List[str] = []
    queue_urls: List[str] = []
    test_eb_configs: List[EbConfiguration] = []

    @classmethod
    def setUpClass(cls) -> None:
        cls.eb_client.create_event_bus(
            Name=cls.event_bus_name,
        )

        cls.sns_topic_arn = cls.sns_client.create_topic(Name=cls.sns_topic_name).get("TopicArn", None)

        cls.test_eb_configs = [
            EbConfiguration(
                rule_name = f"ebrn-{uuid4()}",
                target_id = f"ebti-{uuid4()}",
                event_bus_name=cls.event_bus_name,
                event_pattern='{"source":[{"prefix":"com.test.0"}]}',
            ),
            EbConfiguration(
                rule_name = f"ebrn-{uuid4()}",
                target_id = f"ebti-{uuid4()}",
                event_bus_name=cls.event_bus_name,
                event_pattern='{"source":[{"prefix":"com.test.1"}]}',
                input='"hello, world!"',
            ),
            EbConfiguration(
                rule_name = f"ebrn-{uuid4()}",
                target_id = f"ebti-{uuid4()}",
                event_bus_name=cls.event_bus_name,
                event_pattern='{"source":[{"prefix":"com.test.2"}]}',
                input_path="$.detail-type",
            ),
            EbConfiguration(
                rule_name = f"ebrn-{uuid4()}",
                target_id = f"ebti-{uuid4()}",
                event_bus_name=cls.event_bus_name,
                event_pattern='{"source":[{"prefix":"com.test.3"}]}',
                input_template='{"source": "<source>", "foo": "<foo>"}',
                input_paths_map={"foo": "$.detail.foo", "source": "$.source"},
            ),
        ]

        for params in cls.test_eb_configs:
            cls.eb_client.put_rule(
                Name=params.rule_name,
                EventPattern=params.event_pattern,
                EventBusName=cls.event_bus_name
            )

            target = {"Id": params.target_id, "Arn": cls.sns_topic_arn}

            if params.input:
                target["Input"] = params.input 

            if params.input_path:
                target["InputPath"] = params.input_path
            
            if params.input_paths_map and params.input_template:
                target["InputTransformer"] = {"InputPathsMap": params.input_paths_map, "InputTemplate": params.input_template}

            cls.eb_client.put_targets(
                Rule=params.rule_name,
                EventBusName=cls.event_bus_name,
                Targets=[target]
            )
            
            cls.add_listener(
                event_bus_name=cls.event_bus_name,
                rule_name=params.rule_name,
                target_id=params.target_id
            )

        LOG.debug("created listeners: %s", cls.listener_ids)
        LOG.debug("queue urls: %s", cls.queue_urls)

    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove listeners")
        if cls.listener_ids:
            out = cls.remove_listeners(cls.listener_ids)
            LOG.debug(out)

        LOG.debug("delete test event bus")
        for eb_config in cls.test_eb_configs:
            cls.eb_client.remove_targets(
                Rule=eb_config.rule_name,
                EventBusName=cls.event_bus_name,
                Ids=[eb_config.target_id],
                Force=True
            )
            cls.eb_client.delete_rule(
                Name=eb_config.rule_name,
                EventBusName=cls.event_bus_name,
                Force=True
            )
        cls.sns_client.delete_topic(TopicArn=cls.sns_topic_arn)
        cls.eb_client.delete_event_bus(Name=cls.event_bus_name)
        

    @parameterized.expand(
        [
            (
                0,
                [
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "1", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "2", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.null",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                ],
                condition_func_0,
                10,
            ),
            (
                1,
                [
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "1", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "2", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.null",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.1",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                ],
                condition_func_1,
                5,
            ),
            (
                2,
                [
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "1", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.2",
                        detail_type="foo",
                        detail={"id": "2", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.null",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.2",
                        detail_type="xyz",
                        detail={"id": "2", "abc": "def"},
                    ),
                ],
                condition_func_2,
                5,
            ),
            (
                3,
                [
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "1", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "2", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.null",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.3",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def", "foo": "bar"},
                    ),
                ],
                condition_func_3,
                5,
            ),
        ]
    )
    def test_method_must_return_true(
        self,
        listener_idx: int,
        events: List[EbEvent],
        condition_func: Callable[[str], bool],
        timeout_seconds: int,
    ):
        LOG.debug("purging listener to delete events from previous tests")
        self.purge_listener(listener_idx)

        LOG.debug("sending events")
        self.send_events(events)

        LOG.debug("waiting for event")
        found = self.ctk.wait_until_event_matched(
            listener_id=self.listener_ids[listener_idx],
            condition=condition_func,
            timeout_seconds=timeout_seconds
        )
        self.assertTrue(found)

    @parameterized.expand(
        [
            (  # no event is sent
                0,
                [],
                condition_func_0,
                10,
            ),
            (  # all events don't match
                2,
                [
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.0",
                        detail_type="foo",
                        detail={"id": "1", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.test.2",
                        detail_type="foo",
                        detail={"id": "2", "abc": "def"},
                    ),
                    EbEvent(
                        source="com.null",
                        detail_type="foo",
                        detail={"id": "0", "abc": "def"},
                    ),
                ],
                condition_func_2,
                5,
            ),
        ]
    )
    def test_must_timeout(
        self,
        listener_idx: int,
        events: List[EbEvent],
        condition_func: Callable[[str], bool],
        timeout_seconds: int,
    ):
        LOG.debug("purging listener to delete events from previous tests")
        self.purge_listener(listener_idx)

        LOG.debug("sending events")
        self.send_events(events)

        LOG.debug("waiting for event")
        found = self.ctk.wait_until_event_matched(
            listener_id=self.listener_ids[listener_idx],
            condition=condition_func,
            timeout_seconds=timeout_seconds
        )
        self.assertFalse(found)

    def test_invalid_input(self):
        listener_idx = 0
        with self.assertRaises(InvalidParamException):
            self.ctk.wait_until_event_matched(
                listener_id=self.listener_ids[listener_idx],
                condition=condition_func_0,
                timeout_seconds=10000
            )

    def purge_listener(self, listener_idx):
        # TODO (hawflau): revisit if this should be baked into CTK
        # A listener might have leftover events from previous test cases and might affect current test run
        consecutive_empty_count = 0
        while consecutive_empty_count < 5:
            output = self.ctk.poll_events(
                listener_id=self.listener_ids[listener_idx],
                wait_time_seconds=0,
                max_number_of_messages=10,
            )
            if not output.events:
                consecutive_empty_count += 1
            else:
                consecutive_empty_count = 0

    @classmethod
    def add_listener(cls, event_bus_name, rule_name, target_id) -> str:
        output = cls.ctk.add_listener(event_bus_name, rule_name, target_id)
        cls.listener_ids.append(output.id)
        cls.queue_urls.append(output.components[0].physical_id)
        return output.id

    @classmethod
    def remove_listeners(cls, ids: List[str]):
        output = cls.ctk.remove_listeners(ids=ids)
        LOG.debug(output)

    def send_events(self, events: List[EbEvent]):
        if not events:
            return
        entries = [e.to_entry(self.event_bus_name) for e in events]
        try:
            self.eb_client.put_events(
                Entries=entries,
            )
        except ClientError:
            self.fail("failed to send events")
