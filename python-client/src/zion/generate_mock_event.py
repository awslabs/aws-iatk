# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass
from typing import Optional, List, Callable

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class GenerateBareboneEventOutput:
    """
    zion.generate_barebone_event output
    
    Parameters
    ----------
    event : dict
        mock event
    """
    event: dict

    def __init__(self, data_dict: dict) -> None:
        event = data_dict.get("result", {}).get("output")
        self.event = json.loads(event) if event else None


@dataclass
class GenerateBareboneEventParams:
    """
    zion.generate_barebone_event parameters
    
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
    skip_optional : bool
        if set to true, do not generate optional fields
    """
    registry_name: Optional[str] = None
    schema_name: Optional[str] = None
    schema_version: Optional[str] = None
    event_ref: Optional[str] = None
    skip_optional: Optional[bool] = None

    _rpc_method: str = "mock.generate_barebone_event"

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
        if self.skip_optional:
            params["SkipOptional"] = self.skip_optional
        return params

    def to_payload(self, region, profile) -> Payload:
        return Payload(self._rpc_method, self.to_dict(), region, profile)


# NOTE (hawflau): client method output
@dataclass
class GenerateMockEventOutput:
    """
    zion.generate_mock_event output
    
    Parameters
    ----------
    event : dict
        mock event
    """
    event: dict


# NOTE (hawflau): client method param
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
    skip_optional : bool
        if set to true, do not generate optional fields
    contexts : List[Callable[[dict], dict]]
        a list of callables to apply context on the generated mock event
    """
    registry_name: Optional[str] = None
    schema_name: Optional[str] = None
    schema_version: Optional[str] = None
    event_ref: Optional[str] = None
    skip_optional: Optional[bool] = None
    contexts: Optional[List[Callable[[dict], dict]]] = None
