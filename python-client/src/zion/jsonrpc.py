# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import json
import logging
from uuid import uuid4
from typing import Optional

from zion.version import _version

LOG = logging.getLogger(__name__)
MODULE_NAME = "zion"


class Payload:
    jsonrpc: str = "2.0"
    id: str
    method: str
    params: dict
    _client: str = "python"
    _version: str = _version

    def __init__(self, method: str, params: dict, region: str = None, profile: str = None):
        self.id = str(uuid4())
        self.method = method
        self.params = params
        if region:
            self.params["Region"] = region
        if profile:
            self.params["Profile"] = profile

    def to_dict(self, caller: Optional[str]=None):
        _dict = {
            "jsonrpc": self.jsonrpc,
            "id": self.id,
            "method": self.method,
            "params": self.params,
            "metadata": {
                "client": self._client,
                "version": self._version,
                "caller": caller if caller else self.method,
            }
        }
        return _dict

    def dump_bytes(self, caller: Optional[str]=None):
        return bytes(json.dumps(self.to_dict(caller)), "utf-8")