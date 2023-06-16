import json
import logging
from dataclasses import dataclass
from typing import List, Callable


LOG = logging.getLogger(__name__)


@dataclass
class PollEventsOutput:
    """
    zion.poll_events Output

    Parameters
    ----------
    events : List[str]
        List of event found
    """
    events: List[str]

    def __init__(self, jsonrpc_data_bytes) -> None:
        jsonrpc_data = jsonrpc_data_bytes.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())
        output = data_dict.get("result", {}).get("output", [])
        self.events = output


@dataclass
class PollEventsParams:
    """
    zion.poll_events params

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

    def jsonrpc_dumps(self, region, profile):
        jsonrpc_data = {
            "jsonrpc": "2.0",
            "id": "42",
            "method": self._rpc_method,
            "params": {},
        }
        jsonrpc_data["params"]["ListenerId"] = self.listener_id
        if self.wait_time_seconds is not None:
            jsonrpc_data["params"]["WaitTimeSeconds"] = self.wait_time_seconds
        if self.max_number_of_messages is not None:
            jsonrpc_data["params"]["MaxNumberOfMessages"] = self.max_number_of_messages
        if region:
            jsonrpc_data["params"]["Region"] = region
        if profile:
            jsonrpc_data["params"]["Profile"] = profile

        return bytes(json.dumps(jsonrpc_data), "utf-8")


@dataclass
class WaitUntilEventMatchedParams:
    """
    zion.wait_until_event_matched params

    Parameters
    ----------
    listener_id : str
        Id of the Listener that was created
    condition : Callable[[str], bool]
        Callable fuction that takes a str and returns a bool
    timeout_seconds : int
        Timeout (in seconds) to stop the polling
    """
    condition: Callable[[str], bool]    
    timeout_seconds: int

    _poll_event_params: PollEventsParams

    def __init__(
        self,
        listener_id: str,
        condition: Callable[[str], bool],
        timeout_seconds: int = 30,
    ):
        if timeout_seconds <= 0 or timeout_seconds > 999:
            raise InvalidParamException("timeout_seconds must be between 1 and 999")

        self.condition = condition
        self.timeout_seconds = timeout_seconds

        self._poll_event_params = PollEventsParams(
            listener_id=listener_id,
            wait_time_seconds=3,
            max_number_of_messages=10,
        )


class InvalidParamException(Exception):
    pass
