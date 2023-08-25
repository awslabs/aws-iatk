# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

from subprocess import Popen, PIPE
from datetime import datetime
import pathlib
import json
import boto3
import logging
from dataclasses import dataclass
from functools import wraps

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
    "WaitUntilEventMatchedParams"]

LOG = logging.getLogger(__name__)
zion_service_logger = logging.getLogger("zion.service")


def log_duration(func):
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
    region: str = None
    profile: str = None

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
        input_data = params.jsonrpc_dumps(self.region, self.profile)
        stdout_data = self._popen_zion(input_data, {})
        self._raise_error_if_returned(stdout_data)

        output = PhysicalIdFromStackOutput(stdout_data)

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
        input_data = params.jsonrpc_dumps(self.region, self.profile)
        stdout_data = self._popen_zion(input_data, {})
        self._raise_error_if_returned(stdout_data)

        output = GetStackOutputsOutput(stdout_data)

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
        input_data = params.jsonrpc_dumps(self.region, self.profile)
        stdout_data = self._popen_zion(input_data, {})
        self._raise_error_if_returned(stdout_data)

        output = AddEbListenerOutput(stdout_data)

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
        input_data = params.jsonrpc_dumps(self.region, self.profile)
        stdout_data = self._popen_zion(input_data, {})
        self._raise_error_if_returned(stdout_data)

        output = RemoveListenersOutput(stdout_data)

        LOG.debug(f"Output: {output}")

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
        input_data = params.jsonrpc_dumps(self.region, self.profile)
        stdout_data = self._popen_zion(input_data, {})
        self._raise_error_if_returned(stdout_data)

        output = PollEventsOutput(stdout_data)

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
            out = self.poll_events(params=params._poll_event_params)
            events = out.events
            if events:
                for event in events:
                    if params.condition(event):
                        LOG.debug("event matched")
                        return True

        LOG.debug(f"timeout after {params.timeout_seconds} seconds")
        LOG.debug("no matching event found")
        return False
        
    def patch_aws_client(self, client: boto3.client, sampled = 1) -> boto3.client:
        """
        Patches boto3 client to register event to include generated x-ray trace id and sampling rule as part of request header before invoke/execution

         Parameters
        ----------
        params : client
            boto3.client for specified aws service
               : sampled
            int, value 0 or 1 to select if trace has been sampled or not
            
        
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

    @log_duration
    def _popen_zion(self, input, env_vars):
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


# Set up logging to ``/dev/null`` like a library is supposed to.
# https://docs.python.org/3.3/howto/logging.html#configuring-logging-for-a-library
class NullHandler(logging.Handler):
    def emit(self, record):
        pass


logging.getLogger("zion").addHandler(NullHandler())
