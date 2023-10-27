# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from dataclasses import dataclass
from typing import List, Optional

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class RemoveListenersOutput:
    """
    AWSCtk.remove_listeners Output

    Parameters
    ----------
    message : str
        Message indicates whether or not the remove succeeded.
    """
    message: str

    def __init__(self, data_dict: dict) -> None:
        self.message = data_dict.get("result", {}).get("output", "")


@dataclass
class RemoveListeners_TagFilter:
    """
    Tag filters

    Parameters
    ----------
    key : str
        One part of a key-value pair that makes up a tag. A key is a general label that acts like a category for more specific tag values.
    values : List[str], optional
        One part of a key-value pair that make up a tag. A value acts as a descriptor within a tag category (key). The value can be empty or null.
    """
    key: str
    values: Optional[List[str]] = None

    def to_dict(self) -> dict:
        d = {
            "Key": self.key,
        }
        if self.values:
            d["Values"] = self.values
        return d


@dataclass
class RemoveListenersParams:
    """
    AWSCtk.remove_listeners parameters

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

    def to_dict(self) -> dict:
        params = {}
        if self.ids:
            params["Ids"] = self.ids
        if self.tag_filters:
            params["TagFilters"] = [
                tag_filter.to_dict() for tag_filter in self.tag_filters
            ]
        return params
    
    def to_payload(self, region, profile):
        return Payload(self._rpc_method, self.to_dict(), region, profile)
