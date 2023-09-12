# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass
from .xray import Tree

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class GetTraceTreeOutput:
    """
    zion.get_trace_tree output

    Parameters
    ----------
    trace_tree : Tree
        Trace tree structure of the provided trace id
    """
    trace_tree: Tree

    def __init__(self, jsonrpc_data_bytes) -> None:
        jsonrpc_data = jsonrpc_data_bytes.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())
        trace_tree_output = data_dict.get("result", {}).get("output", {})
        self.trace_tree = Tree(trace_tree_output)


@dataclass
class GetTraceTreeParams:
    """
    zion.get_trace_tree parameters

    Parameters
    ----------
    tracing_header : str
        Trace header to get the trace tree
    """
    tracing_header: str
    _rpc_method: str = "get_trace_tree"

    def jsonrpc_dumps(self, region, profile) -> bytes:
        params = {}
        params["TracingHeader"] = self.tracing_header
        return Payload(self._rpc_method, params, region, profile).dump_bytes()
