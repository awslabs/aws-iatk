# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from dataclasses import dataclass

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class PhysicalIdFromStackOutput:
    """
    AWSCtk.get_physical_id_from_stack output

    Parameters
    ----------
    physical_id : str
        Physical Id of the Resource requested
    """
    physical_id: str

    def __init__(self, data_dict) -> None:
        self.physical_id = data_dict.get("result", {}).get("output", "")


@dataclass
class PhysicalIdFromStackParams:
    """
    AWSCtk.get_physical_id_from_stack parameters

    Parameters
    ----------
    logical_resource_id : str
        Name of the Logical Id within the Stack to fetch
    stack_name : str
        Name of the CloudFormation Stack
    """
    logical_resource_id: str
    stack_name: str
    _rpc_method: str = "get_physical_id"

    def to_dict(self) -> dict:
        return {
            "LogicalResourceId": self.logical_resource_id,
            "StackName": self.stack_name,
        }

    def to_payload(self, region, profile) -> Payload:
        return Payload(self._rpc_method, self.to_dict(), region, profile)
        
