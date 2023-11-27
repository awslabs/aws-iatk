# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from dataclasses import dataclass
from .xray import Tree

from .jsonrpc import Payload


LOG = logging.getLogger(__name__)


@dataclass
class GetTraceTreeOutput:
    """
    AwsIatk.get_trace_tree output

    Parameters
    ----------
    trace_tree : Tree
        Trace tree structure of the provided trace id
    """
    trace_tree: Tree

    def __init__(self, data_dict) -> None:
        trace_tree_output = data_dict.get("result", {}).get("output", {})
        self.trace_tree = Tree(trace_tree_output)


@dataclass
class GetTraceTreeParams:
    """
    AwsIatk.get_trace_tree parameters

    Parameters
    ----------
    tracing_header : str
        Trace header to get the trace tree
    fetch_child_traces: bool
        Flag to determine if linked traces will be included in the tree
    """
    tracing_header: str
    fetch_child_traces: bool
    _rpc_method: str = "get_trace_tree"

    def to_dict(self) -> dict:
        return {
            "TracingHeader": self.tracing_header,
            "FetchChildTraces": self.fetch_child_traces
        }

    def to_payload(self, region, profile) -> Payload:
        return Payload(self._rpc_method, self.to_dict(), region, profile)
