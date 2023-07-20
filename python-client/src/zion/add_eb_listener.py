# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass
from typing import List, Dict, Optional


LOG = logging.getLogger(__name__)


@dataclass
class AddEbListener_Resource:
    """
    Data class that represents the a Resource created during
    zion.add_listener

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
class AddEbListener_InputTrasnformer:
    """
    zion.add_listener parameters

    Parameters
    ----------
    input_template : str
        Input template where you specify placeholders that will be filled with the values of the keys from InputPathsMap to customize the data sent to the target. Enclose each InputPathsMaps value in brackets: <value> 
    input_paths_map : Dict[str, str]
        Map of JSON paths to be extracted from the event. You can then insert these in the template in InputTemplate to produce the output you want to be sent to the target.

        InputPathsMap is an array key-value pairs, where each value is a valid JSON path. You can have as many as 100 key-value pairs. You must use JSON dot notation, not bracket notation.

        The keys cannot start with "AWS." 
    """
    input_template: str
    input_paths_map: Dict[str, str]


@dataclass
class AddEbListenerOutput:
    """
    zion.add_listener output

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

    def __init__(self, jsonrpc_data_bytes) -> None:
        jsonrpc_data = jsonrpc_data_bytes.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())
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
    zion.add_listener parameters

    Parameters
    ----------
    event_bus_name : str
        Name of the AWS Event Bus
    event_pattern : str
        Event Pattern to filter events that arrive on the
        AWS Event Bus
    rule_name : str, optional
        Name of a Rule on the EventBus to replicate
    target_id : str, optional
        Target Id on the given rule to replicate
    tags : Dict[str, str], optional
        A key-value pair associated EventBridge rule.
    """
    event_bus_name: str
    event_pattern: str
    rule_name: Optional[str] = None
    target_id: Optional[str] = None
    tags: Optional[Dict[str, str]] = None

    _rpc_method: str = "test_harness.eventbridge.add_listener"

    def jsonrpc_dumps(self, region, profile):
        jsonrpc_data = {
            "jsonrpc": "2.0",
            "id": "42",
            "method": self._rpc_method,
            "params": {},
        }
        jsonrpc_data["params"]["EventBusName"] = self.event_bus_name
        jsonrpc_data["params"]["EventPattern"] = self.event_pattern
        if self.rule_name:
            jsonrpc_data["params"]["RuleName"] = self.rule_name
        if self.target_id:
            jsonrpc_data["params"]["TargetId"] = self.target_id
        if self.tags:
            jsonrpc_data["params"]["Tags"] = self.tags
        if region:
            jsonrpc_data["params"]["Region"] = region
        if profile:
            jsonrpc_data["params"]["Profile"] = profile

        print(jsonrpc_data)

        return bytes(json.dumps(jsonrpc_data), "utf-8")
