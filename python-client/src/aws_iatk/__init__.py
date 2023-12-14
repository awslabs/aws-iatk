# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

from subprocess import Popen, PIPE
from datetime import datetime
import pathlib
import json
import logging
from dataclasses import dataclass
from functools import wraps
from typing import TYPE_CHECKING, Optional, Dict, List, Callable
import time
import math
import uuid

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
    "AwsIatk", 
    "IatkException", 
    "PhysicalIdFromStackOutput", 
    "GetStackOutputsOutput", 
    "AddEbListenerOutput", 
    "RemoveListenersOutput",
    "RemoveListeners_TagFilter",
    "PollEventsOutput",
    "GetTraceTreeOutput",
    "GenerateBareboneEventOutput",
    "GenerateMockEventOutput",
]

LOG = logging.getLogger(__name__)
iatk_service_logger = logging.getLogger("iatk.service")


def _log_duration(func):
    @wraps(func)
    def wrapper(*args, **kwargs):
        start = datetime.now()
        ret = func(*args, **kwargs)
        LOG.debug("elapsed: %s seconds", (datetime.now() - start).total_seconds())
        return ret

    return wrapper


class IatkException(Exception):
    def __init__(self, message, error_code) -> None:
        super().__init__(message)

        self.error_code = error_code

class RetryableException(Exception):
    def __init__(self, message) -> None:
        super().__init__(message)

@dataclass
class AwsIatk:
    """
    Creates and setups AWS Integrated Application Test Kit

    Parameters
    ----------
    region : str, optional
        AWS Region used to interact with AWS
    profile: str, optional
        AWS Profile used to communicate with AWS resources
    """
    region: Optional[str] = None
    profile: Optional[str] = None

    _iatk_binary_path = (
        pathlib.Path(__file__).parent.parent.joinpath("iatk_service", "iatk").absolute()
    )

    def get_physical_id_from_stack(
        self, logical_resource_id: str, stack_name: str
    ) -> PhysicalIdFromStackOutput:
        """
        Fetch a Physical Id from a Logical Id within an AWS CloudFormation stack

        IAM Permissions Needed
        ----------------------
        cloudformation:DescribeStackResources

        Parameters
        ----------
        logical_resource_id : str
            Name of the Logical Id within the Stack to fetch
        stack_name : str
            Name of the CloudFormation Stack
        
        Returns
        -------
        PhysicalIdFromStackOutput
            Data Class that holds the Physical Id of the resource

        Raises
        ------
        IatkException
            When failed to fetch Physical Id
        """
        params = PhysicalIdFromStackParams(logical_resource_id, stack_name)
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_iatk(payload)
        output = PhysicalIdFromStackOutput(response)
        LOG.debug(f"Physical id: {output.physical_id}, Logical id: {params.logical_resource_id}")
        return output

    def get_stack_outputs(self, stack_name: str, output_names: List[str]) -> GetStackOutputsOutput:
        """
        Fetch Stack Outputs from an AWS CloudFormation stack

        IAM Permissions Needed
        ----------------------
        cloudformation:DescribeStacks

        Parameters
        ----------
        stack_name : str
            Name of the Stack
        output_names : List[str] 
            List of strings that represent the StackOutput Keys   
        
        Returns
        -------
        GetStackOutputsOutput
            Data Class that holds the Stack Outputs of the resource

        Raises
        ------
        IatkException
            When failed to fetch Stack Outputs
        """
        params = GetStackOutputsParams(stack_name, output_names)
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_iatk(payload)
        output = GetStackOutputsOutput(response)
        LOG.debug(f"Output: {output}")
        return output

    def add_listener(self, event_bus_name: str, rule_name: str, target_id: Optional[str] = None, tags: Optional[Dict[str, str]] = None) -> AddEbListenerOutput:
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
        event_bus_name : str
            Name of the AWS Event Bus
        rule_name : str
            Name of a Rule on the EventBus to replicate
        target_id : str, optional
            Target Id on the given rule to replicate
        tags : Dict[str, str], optional
            A key-value pair associated EventBridge rule.
        
        Returns
        -------
        AddEbListenerOutput
            Data Class that holds the Listener created

        Raises
        ------
        IatkException
            When failed to add listener
        """
        params = AddEbListenerParams(event_bus_name, rule_name, target_id, tags)
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_iatk(payload)
        output = AddEbListenerOutput(response)
        LOG.debug(f"Output: {output}")
        return output

    def remove_listeners(self, ids: Optional[List[str]] = None, tag_filters: Optional[List[RemoveListeners_TagFilter]] = None) -> RemoveListenersOutput:
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
        ids : List[str], optional
            List of Listener Ids to remove, one of ids and tag_filters must be supplied
        tag_filters : List[RemoveListeners_TagFilter], optional
            List of RemoveListeners_TagFilter, one of ids and tag_filters must be supplied
        
        Returns
        -------
        RemoveListenersOutput
            Data Class that holds the Listener(s) that were removed

        Raises
        ------
        IatkException
            When failed to remove listener(s)
        """
        params = RemoveListenersParams(ids, tag_filters)
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_iatk(payload)
        output = RemoveListenersOutput(response)
        LOG.debug(f"Output: {output}")
        return output

    def _poll_events(self, params: PollEventsParams, caller: dict = {}) -> PollEventsOutput:
        """
        underlying implementation for poll_events and wait_until_event_matched
        """
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_iatk(payload, caller)
        output = PollEventsOutput(response)
        return output

    def poll_events(self, listener_id: str, wait_time_seconds: int, max_number_of_messages: int) -> PollEventsOutput:
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
        listener_id : str
            Id of the Listener that was created
        wait_time_seconds : int
            Time in seconds to wait for polling
        max_number_of_messages : int
            Max number of messages to poll
        
        Returns
        -------
        PollEventsOutput
            Data Class that holds the Events captured by the Listener

        Raises
        ------
        IatkException
            When failed to Poll Events
        """
        params = PollEventsParams(listener_id, wait_time_seconds, max_number_of_messages)
        output = self._poll_events(params)
        LOG.debug(f"Output: {output}")
        return output

    def wait_until_event_matched(self, listener_id: str, assertion_fn: Callable[[str], None], timeout_seconds: int = 30) -> bool:
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
        listener_id : str
            Id of the Listener that was created
        assertion_fn : Callable[[str], bool]
            Callable function that has an assertion and raises an AssertionError if it fails
        timeout_seconds : int
            Timeout (in seconds) to stop the polling
        
        Returns
        -------
        bool
            True if the event was matched before timeout otherwise False

        Raises
        ------
        IatkException
            When failed to Poll Events
        """
        params = WaitUntilEventMatchedParams(listener_id, assertion_fn, timeout_seconds)
        start = datetime.now()
        elapsed = lambda _: (datetime.now() - start).total_seconds()
        while elapsed(None) < params.timeout_seconds:
            out = self._poll_events(params=params._poll_event_params, caller={"caller": "wait_until_event_matched", "request_id": str(uuid.uuid4())})
            events = out.events
            if events:
                for event in events:
                    try: 
                        params.assertion_fn(event)
                        LOG.debug("event matched")
                        return True
                    except AssertionError as e:
                        LOG.debug(f"Assertion failed: {e}")
                        
        LOG.debug(f"timeout after {params.timeout_seconds} seconds")
        LOG.debug("no matching event found")
        return False

    def _get_trace_tree(self, params: GetTraceTreeParams, caller: dict = {}) -> GetTraceTreeOutput:
        """
        underlying implementation for get_trace_tree and retry_get_trace_tree_until
        """
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_iatk(payload, caller)
        output = GetTraceTreeOutput(response)
        return output

    def get_trace_tree(
        self, tracing_header: str, fetch_child_traces: Optional[bool] = False
    ) -> GetTraceTreeOutput:
        """
        Fetch the trace tree structure using the provided tracing_header

        IAM Permissions Needed
        ----------------------
        xray:BatchGetTraces
        
        Parameters
        ----------
        tracing_header : str
            Trace header to get the trace tree
        fetch_child_traces: bool
            Flag to determine if linked traces will be included in the tree
        
        Returns
        -------
        GetTraceTreeOutput
            Data Class that holds the trace tree structure

        Raises
        ------
        IatkException
            When failed to fetch a trace tree
        """
        params = GetTraceTreeParams(tracing_header, fetch_child_traces)
        output = self._get_trace_tree(params)
        return output

    def _generate_barebone_event(
        self, params: GenerateBareboneEventParams,
    ) -> GenerateBareboneEventOutput:
        payload = params.to_payload(self.region, self.profile)
        response = self._invoke_iatk(payload)
        output = GenerateBareboneEventOutput(response)
        return output

    def generate_mock_event(
        self, 
        registry_name: Optional[str] = None, 
        schema_name: Optional[str] = None,
        schema_version: Optional[str] = None,
        event_ref: Optional[str] = None,
        skip_optional: Optional[bool] = None,
        contexts: Optional[List[Callable[[dict], dict]]] = None
    ) -> GenerateMockEventOutput:
        """
        Generate a mock event based on a schema from EventBridge Schema Registry

        IAM Permissions Needed
        ----------------------
        schemas:DescribeSchema
        
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
        
        Returns
        -------
        GenerateMockEventOutput
            Data Class that holds the trace tree structure

        Raises
        ------
        IatkException
            When failed to fetch a trace tree
        """
        params = GenerateMockEventParams(registry_name, schema_name, schema_version, event_ref, skip_optional, contexts)
        out = self._generate_barebone_event(
            GenerateBareboneEventParams(
                registry_name=params.registry_name,
                schema_name=params.schema_name,
                schema_version=params.schema_version,
                event_ref=params.event_ref,
                skip_optional=params.skip_optional,
            )
        )
        event = out.event

        if params.contexts is not None and type(params.contexts) == list:
            event = self._apply_contexts(event, params.contexts)

        return GenerateMockEventOutput(event)
    
    def retry_until(self, assertion_fn, timeout = 10, retryable_exceptions = (RetryableException,)):
        """
        Decorator function to retry until condition or timeout is met

        IAM Permissions Needed
        ----------------------
        
        Parameters
        ----------
        assertion_fn: Callable[[any], None]
            Callable function that has an assertion and raises an AssertionError if it fails
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
        if(not callable(assertion_fn)):
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
                    try:
                        output = func(*args, **kwargs)
                    except retryable_exceptions:
                        continue
                    try:
                        assertion_fn(output)
                        return True
                    except AssertionError as e:
                        LOG.debug(f"Assertion failed: {e}")
                    time.sleep(math.pow(2, attempt) * delay)
                    attempt += 1
                LOG.debug(f"timeout after {timeout} seconds")
                LOG.debug("condition not satisfied")
                return False
            return _wrapper
        return retry_until_decorator

    def _apply_contexts(self, generated_event: dict, callable_contexts: List[Callable]) -> dict:
        """
        function for looping through provided functions, modifying the event as the client specifies
        """
        for func in callable_contexts:
            generated_event = func(generated_event)
            try:
                json.dumps(generated_event)
            except TypeError:
                raise IatkException(f"context applier {func.__name__} returns a non-JSON-serializable result", 400)
        if generated_event is None:
            raise IatkException("event is empty, make sure function returns a valid event", 404)
        return generated_event
        
    def patch_aws_client(self, client: "boto3.client", sampled = 1) -> "boto3.client":
        """
        Patches boto3 client to register event to include generated x-ray trace id and sampling rule as part of request header before invoke/execution

        Parameters
        ----------
        params : boto3.client
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
    def _invoke_iatk(self, payload: Payload, caller: dict = {}) -> dict:
        input_data = payload.dump_bytes(caller)
        LOG.debug("payload: %s", input_data)
        stdout_data = self._popen_iatk(input_data)
        jsonrpc_data = stdout_data.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())

        # Error is only returned in the response if one happened. Otherwise it is omitted
        if data_dict.get("error", None):
            message = data_dict.get("error", {}).get("message", "")
            error_code = data_dict.get("error", {}).get("Code", 0)
            raise IatkException(message=message, error_code=error_code)
        
        return data_dict

    def _popen_iatk(self, input: bytes, env_vars: Optional[dict]=None) -> bytes:
        LOG.debug("calling iatk rpc with input %s", input)
        p = Popen([self._iatk_binary_path], stdout=PIPE, stdin=PIPE, stderr=PIPE)

        out, err = p.communicate(input=input)
        for line in err.splitlines():
            iatk_service_logger.debug(line.decode())
        return out

    def _raise_error_if_returned(self, output):
        jsonrpc_data = output.decode("utf-8")
        data_dict = json.loads(jsonrpc_data.strip())

        # Error is only returned in the response if one happened. Otherwise it is omitted
        if data_dict.get("error", None):
            message = data_dict.get("error", {}).get("message", "")
            error_code = data_dict.get("error", {}).get("Code", 0)
            raise IatkException(message=message, error_code=error_code)

        
    def retry_get_trace_tree_until(self, tracing_header: str, assertion_fn: Callable[[GetTraceTreeOutput], None], fetch_child_traces: Optional[bool] = False, timeout_seconds: int = 30):
        """
        function to retry get_trace_tree condition or timeout is met

        IAM Permissions Needed
        ----------------------
        xray:BatchGetTraces
        
        Parameters
        ----------
        trace_header:
            x-ray trace header
        assertion_fn : Callable[[GetTraceTreeOutput], bool]
            Callable fuction that makes an assertion and raises an AssertionError if it fails
        timeout_seconds : int
            Timeout (in seconds) to stop the fetching
        fetch_child_traces: bool
            Flag to determine if linked traces will be included in the tree
        Returns
        -------
        bool
            True if the condition was met or false if the timeout is met

        Raises
        ------
        IatkException
            When an exception occurs during get_trace_tree
        """
        params = RetryGetTraceTreeUntilParams(tracing_header, assertion_fn, fetch_child_traces, timeout_seconds)
        @self.retry_until(assertion_fn=params.assertion_fn, timeout=params.timeout_seconds)
        def fetch_trace_tree():
            try:
                response = self._get_trace_tree(
                    params=GetTraceTreeParams(tracing_header=params.tracing_header, fetch_child_traces=fetch_child_traces),
                    caller={"caller" : "retry_get_trace_tree_until", "request_id": str(uuid.uuid4())}
                )
                return response
            except IatkException as e:
                 if "trace not found" in str(e):
                    pass
                    raise RetryableException(e)
                 else:
                    raise IatkException(e, 500)
        try:
            response = fetch_trace_tree()
            return response
        except IatkException as e:
            raise IatkException(e, 500)


# Set up logging to ``/dev/null`` like a library is supposed to.
# https://docs.python.org/3.3/howto/logging.html#configuring-logging-for-a-library
class NullHandler(logging.Handler):
    def emit(self, record):
        pass


logging.getLogger("aws_iatk").addHandler(NullHandler())
