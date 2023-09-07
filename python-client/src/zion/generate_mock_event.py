# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass
from os import PathLike
from typing import Optional, Union

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

    def __init__(self, jsonrpc_data_bytes: bytes) -> None:
        jsonrpc_data = jsonrpc_data_bytes.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())
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
    schema_file : str or path-like 
        path to a local schema file
    event_ref : str
        location to the event in the schema in json schema ref syntax, only applicable for openapi schema
    overrides : dict
        dictionary of overrides to apply on the generated mock event
    skip_optional : bool
        if set to true, do not generate optional fields
    """
    registry_name: Optional[str] = None
    schema_name: Optional[str] = None
    schema_version: Optional[str] = None
    schema_file: Optional[Union[str, PathLike]] = None
    event_ref: Optional[str] = None
    overrides: Optional[dict] = None
    skip_optional: Optional[bool] = None

    _rpc_method: str = "generate_mock_event"

    def jsonrpc_dumps(self, region, profile) -> bytes:
        params = {}
        if self.registry_name:
            params["RegistryName"] = self.registry_name
        if self.schema_name:
            params["SchemaName"] = self.schema_name
        if self.schema_version:
            params["SchemaVersion"] = self.schema_version
        if self.schema_file:
            params["SchemaFile"] = self.schema_file
        if self.event_ref:
            params["EventRef"] = self.event_ref
        if self.overrides:
            params["Overrides"] = self.overrides
        if self.skip_optional:
            params["SkipOptional"] = self.skip_optional
        return Payload(self._rpc_method, params, region, profile).dump_bytes()