# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import json
import logging
from dataclasses import dataclass
from typing import List, Callable


LOG = logging.getLogger(__name__)


@dataclass
class RetryFetchXRayTraceUntilParams:
    """
    zion.wait_until_event_matched params

    Parameters
    ----------
    trace_id:
        Id of the x-ray trace
    condition : Callable[[str], bool]
        Callable fuction that takes a str and returns a bool
    timeout_seconds : int
        Timeout (in seconds) to stop the fetching
    """
    condition: Callable[[str], bool]    
    timeout_seconds: int
    trace_id: str

    def __init__(
        self,
        trace_id: str,
        condition: Callable[[str], bool],
        timeout_seconds: int = 30,
    ):
        if timeout_seconds <= 0 or timeout_seconds > 999:
            raise InvalidParamException("timeout_seconds must be between 1 and 999")

        self.condition = condition
        self.timeout_seconds = timeout_seconds
        self.trace_id = trace_id


class InvalidParamException(Exception):
    pass
