# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass
from typing import List, Callable
from .get_trace_tree import (
    GetTraceTreeOutput
)

LOG = logging.getLogger(__name__)


@dataclass
class RetryFetchXRayTraceUntilParams:
    """
    zion.wait_until_event_matched params

    Parameters
    ----------
    trace_header:
        x-ray trace header
    condition : Callable[[str], bool]
        Callable fuction that takes a str and returns a bool
    timeout_seconds : int
        Timeout (in seconds) to stop the fetching
    """
    condition: Callable[[GetTraceTreeOutput], bool]    
    timeout_seconds: int
    trace_header: str

    def __init__(
        self,
        trace_header: str,
        condition: Callable[[], bool],
        timeout_seconds: int = 30,
    ):
        self.condition = condition
        self.timeout_seconds = timeout_seconds
        self.trace_header = trace_header


class InvalidParamException(Exception):
    pass
