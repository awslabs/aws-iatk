import json
import logging
from dataclasses import dataclass
from typing import List, Optional

LOG = logging.getLogger(__name__)


@dataclass
class RemoveListenersOutput:
    """
    zion.remove_listeners Output

    Parameters
    ----------
    message : str
        Message indicates whether or not the remove succeeded.
    """
    message: str

    def __init__(self, jsonrpc_data_bytes: bytes) -> None:
        jsonrpc_data = jsonrpc_data_bytes.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())
        self.message = data_dict.get("result", {}).get("output", "")


@dataclass
class RemoveListeners_TagFilter:
    """
    Tag filters

    Parameters
    ----------
    key : str
        One part of a key-value pair that makes up a tag. A key is a general label that acts like a category for more specific tag values.
    values : List[str]
        One part of a key-value pair that make up a tag. A value acts as a descriptor within a tag category (key). The value can be empty or null.
    """
    key: str
    values: List[str]


@dataclass
class RemoveListenersParams:
    """
    zion.remove_listeners parameters

    Parameters
    ----------
    ids : List[str], optional
        List of Listener Ids to remove
    tag_filters : List[RemoveListeners_TagFilter], optional
        List of RemoveListeners_TagFilter
    """

    ids: Optional[List[str]] = None
    tag_filters: Optional[List[RemoveListeners_TagFilter]] = None
    _rpc_method: str = "test_harness.eventbridge.remove_listeners"

    def jsonrpc_dumps(self, region, profile):
        jsonrpc_data = {
            "jsonrpc": "2.0",
            "id": "42",
            "method": self._rpc_method,
            "params": {},
        }
        if self.ids:
            jsonrpc_data["params"]["Ids"] = self.ids
        if self.tag_filters:
            jsonrpc_data["params"]["TagFilters"] = self.tag_filters
        if region:
            jsonrpc_data["params"]["Region"] = region
        if profile:
            jsonrpc_data["params"]["Profile"] = profile

        return bytes(json.dumps(jsonrpc_data), "utf-8")
