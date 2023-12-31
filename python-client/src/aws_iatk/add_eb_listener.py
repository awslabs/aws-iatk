# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass
from typing import List, Dict, Optional

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class AddEbListener_Resource:
    """
    Data class that represents the a Resource created during
    AwsIatk.add_listener

    Parameters
    ----------
    type : str
        Type of resource created (CloudFormation Types e.g AWS::SQS::Queue)
    physical_id : str
        Physical Id of the resource created
    arn : str
        Arn of the resource created
    """
    type: str
    physical_id: str
    arn: str

    def __init__(self, jsonrpc_data_dict) -> None:
        self.type = jsonrpc_data_dict.get("Type", "")
        self.physical_id = jsonrpc_data_dict.get("PhysicalID", "")
        self.arn = jsonrpc_data_dict.get("ARN", "")


@dataclass
class AddEbListenerOutput:
    """
    AwsIatk.add_listener output

    Parameters
    ----------
    id : str
        Id that corresponds to the listener created
    target_under_test : AddEbListener_Resource
        Target Resource that test resources were added
    components : List[AddEbListener_Resource]
        List of all Resources created to support the listener
        on the `target_under_test`
    """
    id: str
    target_under_test: AddEbListener_Resource
    components: List[AddEbListener_Resource]

    def __init__(self, data_dict: dict) -> None:
        output = data_dict.get("result", {}).get("output", {})
        self.id = output.get("Id", "")
        self.target_under_test = AddEbListener_Resource(
            output.get("TargetUnderTest", {})
        )
        self.components = []
        for component in output.get("Components", []):
            self.components.append(AddEbListener_Resource(component))


@dataclass
class AddEbListenerParams:
    """
    AwsIatk.add_listener parameters

    Parameters
    ----------
    event_bus_name : str
        Name of the AWS Event Bus
    rule_name : str
        Name of a Rule on the EventBus to replicate
    target_id : str, optional
        Target Id on the given rule to replicate
    tags : Dict[str, str], optional
        A key-value pair associated EventBridge rule.
    """
    event_bus_name: str
    rule_name: str
    target_id: Optional[str] = None
    tags: Optional[Dict[str, str]] = None

    _rpc_method: str = "test_harness.eventbridge.add_listener"

    def to_dict(self) -> dict:
        params = {}
        params["EventBusName"] = self.event_bus_name
        params["RuleName"] = self.rule_name
        if self.target_id:
            params["TargetId"] = self.target_id
        if self.tags:
            params["Tags"] = self.tags
        return params

    def to_payload(self, region, profile) -> Payload:
        return Payload(self._rpc_method, self.to_dict(), region, profile)
