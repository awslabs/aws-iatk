# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from dataclasses import dataclass
from typing import Callable
from .get_trace_tree import (
    GetTraceTreeOutput
)

LOG = logging.getLogger(__name__)


@dataclass
class RetryGetTraceTreeUntilParams:
    """
    AwsIatk.wait_until_event_matched params

    Parameters
    ----------
    trace_header:
        x-ray trace header
    assertion_fn : Callable[[GetTraceTreeOutput], None]
        Callable fuction that makes an assertion and raises an AssertionError if it fails
    timeout_seconds : int
        Timeout (in seconds) to stop the fetching
    fetch_child_traces: bool
        Flag to determine if linked traces will be included in the tree
    """
    assertion_fn: Callable[[GetTraceTreeOutput], None]    
    timeout_seconds: int
    tracing_header: str
    fetch_child_traces: bool

    def __init__(
        self,
        tracing_header: str,
        assertion_fn: Callable[[GetTraceTreeOutput], None],
        fetch_child_traces: bool,
        timeout_seconds: int = 30,
    ):
        self.assertion_fn = assertion_fn
        self.timeout_seconds = timeout_seconds
        self.tracing_header = tracing_header
        self.fetch_child_traces = fetch_child_traces


class InvalidParamException(Exception):
    pass
