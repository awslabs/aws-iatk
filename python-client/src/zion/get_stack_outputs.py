# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass
from typing import List, Dict

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class GetStackOutputsOutput:
    """
    zion.get_stack_outputs output

    Parameters
    ----------
    outputs : Dict[str, str]
        Dictionary of keys being the StackOutput Key and value
        being the StackOutput Value       
    """
    outputs: Dict[str, str]

    def __init__(self, jsonrpc_data_bytes: bytes) -> None:
        jsonrpc_data = jsonrpc_data_bytes.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())
        self.outputs = data_dict.get("result", {}).get("output", {})


@dataclass
class GetStackOutputsParams:
    """
    zion.get_stack_outputs parameters

    Parameters
    ----------
    stack_name : str
        Name of the Stack
    output_names : List[str] 
        List of strings that represent the StackOutput Keys        
    """
    stack_name: str
    output_names: List[str]
    _rpc_method: str = "get_stack_outputs"

    def jsonrpc_dumps(self, region, profile) -> bytes:
        params = {}
        params["StackName"] = self.stack_name
        params["OutputNames"] = self.output_names
        return Payload(self._rpc_method, params, region, profile).dump_bytes()
