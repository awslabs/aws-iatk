# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from dataclasses import dataclass
from typing import Optional, List

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class GenerateMockEventOutput:
    """
    zion.generate_mock_event output
    
    Parameters
    ----------
    event : str
        mock event
    """
    event: str

    def __init__(self, data_dict: dict) -> None:
        self.event = data_dict.get("result", {}).get("output", "")


@dataclass
class GenerateMockEventParams:
    """
    zion.generate_mock_event parameters
    
    Parameters
    ----------
    registry_name : str
        name of the registry of the schema stored in EventBridge Schema Registry
    schema_name : str
        name of the schema stored in EventBridge Schema Registry
    schema_version : str
        version of the schema stored in EventBridge Schema Registry
    event_ref : str
        location to the event in the schema in json schema ref syntax, only applicable for openapi schema
    context: List[str]
        a list of context to apply on the generated event. Currently only support "eventbridge.v0", which applies context for an EventBridge Event.
    overrides : dict
        dictionary of overrides to apply on the generated mock event
    skip_optional : bool
        if set to true, do not generate optional fields
    """
    registry_name: Optional[str] = None
    schema_name: Optional[str] = None
    schema_version: Optional[str] = None
    event_ref: Optional[str] = None
    context: Optional[List[str]] = None
    overrides: Optional[dict] = None
    skip_optional: Optional[bool] = None

    _rpc_method: str = "generate_mock_event"

    def to_dict(self) -> dict:
        params = {}
        if self.registry_name:
            params["RegistryName"] = self.registry_name
        if self.schema_name:
            params["SchemaName"] = self.schema_name
        if self.schema_version:
            params["SchemaVersion"] = self.schema_version
        if self.event_ref:
            params["EventRef"] = self.event_ref
        if self.context:
            params["Context"] = self.context
        if self.overrides:
            params["Overrides"] = self.overrides
        if self.skip_optional:
            params["SkipOptional"] = self.skip_optional
        return params

    def to_payload(self, region, profile) -> Payload:
        return Payload(self._rpc_method, self.to_dict(), region, profile)