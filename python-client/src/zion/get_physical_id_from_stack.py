# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class PhysicalIdFromStackOutput:
    """
    zion.get_physical_id_from_stack output

    Parameters
    ----------
    physical_id : str
        Physical Id of the Resource requested
    """
    physical_id: str

    def __init__(self, jsonrpc_data_btyes) -> None:
        jsonrpc_data = jsonrpc_data_btyes.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())
        self.physical_id = data_dict.get("result", {}).get("output", "")


@dataclass
class PhysicalIdFromStackParams:
    """
    zion.get_physical_id_from_stack parameters

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

    def jsonrpc_dumps(self, region, profile) -> bytes:
        params = {}
        params["LogicalResourceId"] = self.logical_resource_id
        params["StackName"] = self.stack_name
        return Payload(self._rpc_method, params, region, profile).dump_bytes()
