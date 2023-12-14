# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from dataclasses import dataclass
from typing import List, Dict

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class GetStackOutputsOutput:
    """
    AwsIatk.get_stack_outputs output

    Parameters
    ----------
    outputs : Dict[str, str]
        Dictionary of keys being the StackOutput Key and value
        being the StackOutput Value       
    """
    outputs: Dict[str, str]

    def __init__(self, data_dict: dict) -> None:
        self.outputs = data_dict.get("result", {}).get("output", {})


@dataclass
class GetStackOutputsParams:
    """
    AwsIatk.get_stack_outputs parameters

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

    def to_dict(self) -> dict:
        return {
            "StackName": self.stack_name,
            "OutputNames": self.output_names,
        }

    def to_payload(self, region, profile) -> Payload:
        return Payload(self._rpc_method, self.to_dict(), region, profile)
