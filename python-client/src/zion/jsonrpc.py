# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import json
import logging
from uuid import uuid4


LOG = logging.getLogger(__name__)


class Payload:
    jsonrpc: str = "2.0"
    id: str
    method: str
    params: dict

    def __init__(self, method: str, params: dict, region: str = None, profile: str = None):
        self.id = str(uuid4())
        self.method = method
        self.params = params
        if region:
            self.params["Region"] = region
        if profile:
            self.params["Profile"] = profile

    def to_dict(self):
        return {
            "jsonrpc": self.jsonrpc,
            "id": self.id,
            "method": self.method,
            "params": self.params,
        }

    def dump_bytes(self):
        return bytes(json.dumps(self.to_dict()), "utf-8")