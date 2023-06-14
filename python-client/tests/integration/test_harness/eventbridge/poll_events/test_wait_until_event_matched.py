"""
Integration tests for zion.wait_until_event_matched
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

from zion import (
    Zion,
    AddEbListenerParams,
    RemoveListenersParams,
    PollEventsParams,
    WaitUntilEventMatchedParams,
    ZionException,
)
from zion.add_eb_listener import AddEbListener_InputTrasnformer
from zion.poll_events import InvalidParamException


LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="zion", level=logging.DEBUG)


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


class TestZion_wait_until_event_matched(TestCase):
    zion = Zion()
    eb_client = boto3.client("events")
    sqs_client = boto3.client("sqs")
    event_bus_name: str = f"eb-{uuid4()}"
    listener_ids: List[str] = []
    queue_urls: List[str] = []
    add_listener_params: List[AddEbListenerParams] = []

    @classmethod
    def setUpClass(cls) -> None:
        cls.eb_client.create_event_bus(
            Name=cls.event_bus_name,
        )

        cls.add_listener_params = [
            AddEbListenerParams(
                event_bus_name=cls.event_bus_name,
                event_pattern='{"source":[{"prefix":"com.test.0"}]}',
            ),
            AddEbListenerParams(
                event_bus_name=cls.event_bus_name,
                event_pattern='{"source":[{"prefix":"com.test.1"}]}',
                input='"hello, world!"',
            ),
            AddEbListenerParams(
                event_bus_name=cls.event_bus_name,
                event_pattern='{"source":[{"prefix":"com.test.2"}]}',
                input_path="$.detail-type",
            ),
            AddEbListenerParams(
                event_bus_name=cls.event_bus_name,
                event_pattern='{"source":[{"prefix":"com.test.3"}]}',
                input_transformer=AddEbListener_InputTrasnformer(
                    input_template='{"source": "<source>", "foo": "<foo>"}',
                    input_paths_map={"foo": "$.detail.foo", "source": "$.source"},
                ),
            ),
        ]

        for params in cls.add_listener_params:
            cls.add_listener(params)

        LOG.debug("created listeners: %s", cls.listener_ids)
        LOG.debug("queue urls: %s", cls.queue_urls)

    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove listeners")
        if cls.listener_ids:
            out = cls.remove_listeners(cls.listener_ids)
            LOG.debug(out)

        LOG.debug("delete test event bus")
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
        found = self.zion.wait_until_event_matched(
            params=WaitUntilEventMatchedParams(
                listener_id=self.listener_ids[listener_idx],
                condition=condition_func,
                timeout_seconds=timeout_seconds,
            )
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
        found = self.zion.wait_until_event_matched(
            params=WaitUntilEventMatchedParams(
                listener_id=self.listener_ids[listener_idx],
                condition=condition_func,
                timeout_seconds=timeout_seconds,
            )
        )
        self.assertFalse(found)

    def test_invalid_input(self):
        listener_idx = 0
        with self.assertRaises(InvalidParamException):
            self.zion.wait_until_event_matched(
                params=WaitUntilEventMatchedParams(
                    listener_id=self.listener_ids[listener_idx],
                    condition=condition_func_0,
                    timeout_seconds=10000,
                )
            )

    def purge_listener(self, listener_idx):
        # TODO (hawflau): revisit if this should be baked into Zion
        # A listener might have leftover events from previous test cases and might affect current test run
        consecutive_empty_count = 0
        params = PollEventsParams(
            listener_id=self.listener_ids[listener_idx],
            wait_time_seconds=0,
            max_number_of_messages=10,
        )
        while consecutive_empty_count < 5:
            output = self.zion.poll_events(params=params)
            if not output.events:
                consecutive_empty_count += 1
            else:
                consecutive_empty_count = 0

    @classmethod
    def add_listener(cls, params: AddEbListenerParams) -> str:
        params.event_bus_name = cls.event_bus_name
        output = cls.zion.add_listener(params=params)
        cls.listener_ids.append(output.id)
        cls.queue_urls.append(output.components[0].physical_id)
        return output.id

    @classmethod
    def remove_listeners(cls, ids: List[str]):
        params = RemoveListenersParams(ids=ids)
        output = cls.zion.remove_listeners(params=params)
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
