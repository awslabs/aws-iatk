# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

from subprocess import Popen, PIPE
from datetime import datetime
import pathlib
import json
import logging
from dataclasses import dataclass
from functools import wraps
from typing import TYPE_CHECKING, Optional
import time
import math

from .get_physical_id_from_stack import (
    PhysicalIdFromStackOutput,
    PhysicalIdFromStackParams,
)
from .get_stack_outputs import (
    GetStackOutputsOutput,
    GetStackOutputsParams,
)
from .add_eb_listener import (
    AddEbListenerOutput,
    AddEbListenerParams,
)
from .remove_listeners import (
    RemoveListenersOutput,
    RemoveListenersParams,
    RemoveListeners_TagFilter,
)
from .poll_events import (
    PollEventsOutput,
    PollEventsParams,
    WaitUntilEventMatchedParams,
)
from .get_trace_tree import (
    GetTraceTreeOutput,
    GetTraceTreeParams
)
from .retry_xray_trace import (
    RetryGetTraceTreeUntilParams,
)
from .generate_mock_event import (
    GenerateBareboneEventOutput,
    GenerateBareboneEventParams,
    GenerateMockEventOutput,
    GenerateMockEventParams,
)
from .jsonrpc import Payload

if TYPE_CHECKING:
    import boto3

__all__ = [
    "Zion", 
    "ZionException", 
    "PhysicalIdFromStackOutput", 
    "PhysicalIdFromStackParams", 
    "GetStackOutputsOutput", 
    "GetStackOutputsParams", 
    "AddEbListenerOutput", 
    "AddEbListenerParams",
    "RemoveListenersOutput",
    "RemoveListenersParams",
    "RemoveListeners_TagFilter",
    "PollEventsOutput",
    "PollEventsParams",
    "WaitUntilEventMatchedParams",
    "GetTraceTreeParams",
    "GetTraceTreeOutput",
    "RetryGetTraceTreeUntilParams",
    "GenerateBareboneEventOutput",
    "GenerateBareboneEventParams",
    "GenerateMockEventOutput",
    "GenerateMockEventParams",
]

LOG = logging.getLogger(__name__)
zion_service_logger = logging.getLogger("zion.service")


def _log_duration(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        start = datetime.now()
        ret = func(*args, **kwargs)
        LOG.debug("elapsed: %s seconds", (datetime.now() - start).total_seconds())
        return ret

    return wrapper


class ZionException(Exception):
    def __init__(self, message, error_code) -> None:
        super().__init__(message)

        self.error_code = error_code


@dataclass
class Zion:
    """
    Creates and setups Zion

    Parameters
    ----------
    region : str, optional
        AWS Region used to interact with AWS
    profile: str, optional
        AWS Profile used to communicate with AWS resources
    """
    region: Optional[str] = None
    profile: Optional[str] = None

    _zion_binary_path = (
        pathlib.Path(__file__).parent.parent.joinpath("zion_service", "zion").absolute()
    )

    def get_physical_id_from_stack(
        self, params: PhysicalIdFromStackParams
    ) -> PhysicalIdFromStackOutput:
        """
        Fetch a Phsyical Id from a Logical Id within an AWS CloudFormation stack

        IAM Permissions Needed
        ----------------------
        cloudformation:DescribeStackResources

        Parameters
        ----------
        params : PhysicalIdFromStackParams
            Data Class that holds required parameters
        
        Returns
        -------
        PhysicalIdFromStackOutput
            Data Class that holds the Phsyical Id of the resource

        Raises
        ------
        ZionException
            When failed to fetch Phsyical Id
        """
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_zion(payload)
        output = PhysicalIdFromStackOutput(response)
        LOG.debug(f"Physical id: {output.physical_id}, Logical id: {params.logical_resource_id}")
        return output

    def get_stack_outputs(self, params: GetStackOutputsParams) -> GetStackOutputsOutput:
        """
        Fetch Stack Outputs from an AWS CloudFormation stack

        IAM Permissions Needed
        ----------------------
        cloudformation:DescribeStacks

        Parameters
        ----------
        params : GetStackOutputsParams
            Data Class that holds required parameters
        
        Returns
        -------
        GetStackOutputsOutput
            Data Class that holds the Stack Outputs of the resource

        Raises
        ------
        ZionException
            When failed to fetch Stack Outputs
        """
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_zion(payload)
        output = GetStackOutputsOutput(response)
        LOG.debug(f"Output: {output}")
        return output

    def add_listener(self, params: AddEbListenerParams) -> AddEbListenerOutput:
        """
        Add Listener Resource to an AWS Event Bridge Bus to enable testing

        IAM Permissions Needed
        ----------------------
        events:DescribeEventBus
        events:DescribeRule
        events:PutRule
        events:PutTargets
        events:DeleteRule
        events:RemoveTargets
        events:TagResource

        sqs:CreateQueue
        sqs:GetQueueAttributes
        sqs:GetQueueUrl
        sqs:DeleteQueue
        sqs:TagQueue

        Parameters
        ----------
        params : AddEbListenerParams
            Data Class that holds required parameters
        
        Returns
        -------
        AddEbListenerOutput
            Data Class that holds the Listener created

        Raises
        ------
        ZionException
            When failed to add listener
        """
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_zion(payload)
        output = AddEbListenerOutput(response)
        LOG.debug(f"Output: {output}")
        return output

    def remove_listeners(self, params: RemoveListenersParams) -> RemoveListenersOutput:
        """
        Remove Listener Resource(s) from an AWS Event Bridge Bus

        IAM Permissions Needed
        ----------------------
        tag:GetResources
        
        sqs:DeleteQueue
        sqs:GetQueueUrl
        sqs:GetQueueAttributes
        sqs:ListQueueTags

        events:DeleteRule
        events:RemoveTargets
        events:ListTargetsByRule
        
        Parameters
        ----------
        params : RemoveListenersParams
            Data Class that holds required parameters
        
        Returns
        -------
        RemoveListenersOutput
            Data Class that holds the Listener(s) that were removed

        Raises
        ------
        ZionException
            When failed to remove listener(s)
        """
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_zion(payload)
        output = RemoveListenersOutput(response)
        LOG.debug(f"Output: {output}")
        return output

    def _poll_events(self, params: PollEventsParams, caller: str=None) -> PollEventsOutput:
        """
        underlying implementation for poll_events and wait_until_event_matched
        """
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_zion(payload, caller)
        output = PollEventsOutput(response)
        return output

    def poll_events(self, params: PollEventsParams) -> PollEventsOutput:
        """
        Poll Events from a specific Listener

        IAM Permissions Needed
        ----------------------
        sqs:GetQueueUrl
        sqs:ListQueueTags
        sqs:ReceiveMessage
        sqs:DeleteMessage
        sqs:GetQueueAttributes

        events:DescribeRule

        Parameters
        ----------
        params : PollEventsParams
            Data Class that holds required parameters
        
        Returns
        -------
        PollEventsOutput
            Data Class that holds the Events captured by the Listener

        Raises
        ------
        ZionException
            When failed to Poll Events
        """
        output = self._poll_events(params)
        LOG.debug(f"Output: {output}")
        return output

    def wait_until_event_matched(self, params: WaitUntilEventMatchedParams) -> bool:
        """
        Poll Events on a given Listener until a match is found or timeout met.

        IAM Permissions Needed
        ----------------------
        sqs:GetQueueUrl
        sqs:ListQueueTags
        sqs:ReceiveMessage
        sqs:DeleteMessage
        sqs:GetQueueAttributes

        events:DescribeRule
        
        Parameters
        ----------
        params : WaitUntilEventMatchedParams
            Data Class that holds required parameters
        
        Returns
        -------
        bool
            True if the event was matched before timeout otherwise False

        Raises
        ------
        ZionException
            When failed to Poll Events
        """
        start = datetime.now()
        elapsed = lambda _: (datetime.now() - start).total_seconds()
        while elapsed(None) < params.timeout_seconds:
            out = self._poll_events(params=params._poll_event_params, caller="wait_until_event_matched")
            events = out.events
            if events:
                for event in events:
                    if params.condition(event):
                        LOG.debug("event matched")
                        return True

        LOG.debug(f"timeout after {params.timeout_seconds} seconds")
        LOG.debug("no matching event found")
        return False

    def _get_trace_tree(self, params: GetTraceTreeParams, caller: str=None) -> GetTraceTreeOutput:
        """
        underlying implementation for get_trace_tree and retry_get_trace_tree_until
        """
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_zion(payload, caller)
        output = GetTraceTreeOutput(response)
        return output

    def get_trace_tree(
        self, params: GetTraceTreeParams
    ) -> GetTraceTreeOutput:
        """
        Fetch the trace tree structure using the provided tracing_header

        IAM Permissions Needed
        ----------------------
        xray:BatchGetTraces
        
        Parameters
        ----------
        params : GetTraceTreeParams
            Data Class that holds required parameters
        
        Returns
        -------
        GetTraceTreeOutput
            Data Class that holds the trace tree structure

        Raises
        ------
        ZionException
            When failed to fetch a trace tree
        """
        output = self._get_trace_tree(params)
        LOG.debug(f"Output: {output}")
        return output

    # TODO (lauwing): make it a "private" method since there's no strong use case for using it alone
    def generate_barebone_event(
        self, params: GenerateBareboneEventParams,
    ) -> GenerateBareboneEventOutput:
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_zion(payload)
        output = GenerateBareboneEventOutput(response)
        LOG.debug(f"Output: {output}")
        return output

    def generate_mock_event(
        self, params: GenerateMockEventParams
    ) -> GenerateMockEventOutput:
        """
        Generate a mock event based on a schema from EventBridge Schema Registry

        IAM Permissions Needed
        ----------------------
        schemas:DescribeSchema
        
        Parameters
        ----------
        params : GenerateMockEventParams
            Data Class that holds required parameters
        
        Returns
        -------
        GenerateMockEventOutput
            Data Class that holds the trace tree structure

        Raises
        ------
        ZionException
            When failed to fetch a trace tree
        """
        out = self.generate_barebone_event(
            GenerateBareboneEventParams(
                registry_name=params.registry_name,
                schema_name=params.schema_name,
                schema_version=params.schema_version,
                event_ref=params.event_ref,
                skip_optional=params.skip_optional,
            )
        )
        event = out.event

        # TODO: apply context

        return GenerateMockEventOutput(event)
    
    def retry_until(self, condition, timeout = 10):
        """
        Decorator function to retry until condition or timeout is met

        IAM Permissions Needed
        ----------------------
        
        Parameters
        ----------
        condition: Callable[[any], bool]
        Callable function that takes any type and returns a bool

        timeout: int or float
        value that specifies how long the function will retry for until it times out
        
        Returns
        -------
        bool
            True if the condition was met or false if the timeout is met

        Raises
        ------
        ValueException
            When timeout is a negative number
        TypeException
            When timeout or condition is not a suitable type
        """
        if(not(isinstance(timeout, int) or  isinstance(timeout, float))):
            raise TypeError("timeout must be an int or float")
        elif(timeout < 0):
            raise ValueError("timeout must not be a negative value")
        if(not callable(condition)):
            raise TypeError("condition is not a callable function")
        def retry_until_decorator(func):
            @wraps(func)
            def _wrapper(*args, **kwargs):
                start = datetime.now()
                attempt = 1
                delay = 0.05
                elapsed = lambda _: (datetime.now() - start).total_seconds()
                if timeout == 0:
                    elapsed = lambda _: -1
                while elapsed(None) < timeout:
                    output = func(*args, **kwargs)
                    if condition(output):
                        return True
                    time.sleep(math.pow(2, attempt) * delay)
                    attempt += 1
                LOG.debug(f"timeout after {timeout} seconds")
                LOG.debug("condition not satisfied")
                return False
            return _wrapper
        return retry_until_decorator
    
        
    def patch_aws_client(self, client: "boto3.client", sampled = 1) -> "boto3.client":
        """
        Patches boto3 client to register event to include generated x-ray trace id and sampling rule as part of request header before invoke/execution

        Parameters
        ----------
        params : client
            boto3.client for specified aws service
        sampled : int
            value 0 to not sample the request or value 1 to sample the request
        
        Returns
        -------
        boto3.client
            same client passed in the params with the event registered
        """
        def _add_header(request, **kwargs):
            trace_id_string= 'Root=;Sampled={}'.format(sampled)
            
            request.headers.add_header('X-Amzn-Trace-Id', trace_id_string)
            LOG.debug(f"Trace ID format: {trace_id_string}")

        service_name = client.meta.service_model.service_name
        event_string = 'before-sign.{}.*'
        LOG.debug(f"service id: {client.meta.service_model.service_id}, service name: {service_name}")
        
        client.meta.events.register(event_string.format(service_name), _add_header)
        
        return client

    @_log_duration
    def _invoke_zion(self, payload: Payload, caller: str=None) -> dict:
        input_data = payload.dump_bytes(caller)
        LOG.debug("payload: %s", input_data)
        stdout_data = self._popen_zion(input_data)
        jsonrpc_data = stdout_data.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())

        # Error is only returned in the response if one happened. Otherwise it is omitted
        if data_dict.get("error", None):
            message = data_dict.get("error", {}).get("message", "")
            error_code = data_dict.get("error", {}).get("Code", 0)
            raise ZionException(message=message, error_code=error_code)
        
        return data_dict

    def _popen_zion(self, input: bytes, env_vars: Optional[dict]=None) -> bytes:
        LOG.debug("calling zion rpc with input %s", input)
        p = Popen([self._zion_binary_path], stdout=PIPE, stdin=PIPE, stderr=PIPE)

        out, err = p.communicate(input=input)
        for line in err.splitlines():
            zion_service_logger.debug(line.decode())
        return out

    def _raise_error_if_returned(self, output):
        jsonrpc_data = output.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())

        # Error is only returned in the response if one happened. Otherwise it is omitted
        if data_dict.get("error", None):
            message = data_dict.get("error", {}).get("message", "")
            error_code = data_dict.get("error", {}).get("Code", 0)
            raise ZionException(message=message, error_code=error_code)

        
    def retry_get_trace_tree_until(self, params: RetryGetTraceTreeUntilParams):
        """
        function to retry get_trace_tree condition or timeout is met

        IAM Permissions Needed
        ----------------------
        xray:BatchGetTraces
        
        Parameters
        ----------
        condition: Callable[[GetTraceTreeOutput], bool]
        Callable function that takes any type and returns a bool

        timeout: int or float
        value that specifies how long the function will retry for until it times out
        
        Returns
        -------
        bool
            True if the condition was met or false if the timeout is met

        Raises
        ------
        Zionexception
            When an exception occurs during get_trace_tree
        """
        @self.retry_until(condition=params.condition, timeout=params.timeout_seconds)
        def fetch_trace_tree():
            response = self._get_trace_tree(
                params=GetTraceTreeParams(tracing_header=params.tracing_header),
                caller="retry_get_trace_tree_until",
            )
            return response
        try:
            response = fetch_trace_tree()
            return response
        except ZionException as e:
            raise ZionException(e, 500)


# Set up logging to ``/dev/null`` like a library is supposed to.
# https://docs.python.org/3.3/howto/logging.html#configuring-logging-for-a-library
class NullHandler(logging.Handler):
    def emit(self, record):
        pass


logging.getLogger("zion").addHandler(NullHandler())
