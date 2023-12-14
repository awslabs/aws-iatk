# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from dataclasses import dataclass
from typing import List, Callable

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class PollEventsOutput:
    """
    AwsIatk.poll_events Output

    Parameters
    ----------
    events : List[str]
        List of event found
    """
    events: List[str]

    def __init__(self, data_dict) -> None:
        output = data_dict.get("result", {}).get("output", [])
        self.events = output


@dataclass
class PollEventsParams:
    """
    AwsIatk.poll_events params

    Parameters
    ----------
    listener_id : str
        Id of the Listener that was created
    wait_time_seconds : int
        Time in seconds to wait for polling
    max_number_of_messages : int
        Max number of messages to poll
    """
    listener_id: str
    wait_time_seconds: int
    max_number_of_messages: int

    _rpc_method: str = "test_harness.eventbridge.poll_events"

    def to_dict(self):
        params = {}
        params["ListenerId"] = self.listener_id
        if self.wait_time_seconds is not None:
            params["WaitTimeSeconds"] = self.wait_time_seconds
        if self.max_number_of_messages is not None:
            params["MaxNumberOfMessages"] = self.max_number_of_messages
        return params

    def to_payload(self, region, profile):
        return Payload(self._rpc_method, self.to_dict(), region, profile)


@dataclass
class WaitUntilEventMatchedParams:
    """
    AwsIatk.wait_until_event_matched params

    Parameters
    ----------
    listener_id : str
        Id of the Listener that was created
    assertion_fn : Callable[[str], None]
        Callable fuction that makes an assertion and raises an AssertionError if it fails
    timeout_seconds : int
        Timeout (in seconds) to stop the polling
    """
    assertion_fn: Callable[[str], None]    
    timeout_seconds: int

    _poll_event_params: PollEventsParams

    def __init__(
        self,
        listener_id: str,
        assertion_fn: Callable[[str], None],
        timeout_seconds: int = 30,
    ):
        if timeout_seconds <= 0 or timeout_seconds > 999:
            raise InvalidParamException("timeout_seconds must be between 1 and 999")

        self.assertion_fn = assertion_fn
        self.timeout_seconds = timeout_seconds

        self._poll_event_params = PollEventsParams(
            listener_id=listener_id,
            wait_time_seconds=3,
            max_number_of_messages=10,
        )


class InvalidParamException(Exception):
    pass
