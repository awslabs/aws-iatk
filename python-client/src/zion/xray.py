# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

import logging
from dataclasses import dataclass
from enum import Enum
from typing import List, Dict, Optional, Union


LOG = logging.getLogger(__name__)


class ReferenceType(Enum):
    parent = "parent"
    child = "child"


@dataclass
class LinkAttributes:
    """
    Link Attribute
    
    Parameters
    ----------
    reference_type : ReferenceType
        Type of the link reference
    """
    reference_type: ReferenceType

    def __init__(self, input_dict: dict):
        self.reference_type = ReferenceType(
            input_dict["aws.xray.reserved.reference_type"]
        )


@dataclass
class Link:
    """
    Link

    Parameters
    ----------
    trace_id : str
        ID of the linked trace
    id : str
        ID of the link
    attributes : 
    """
    trace_id: str
    id: str
    attributes: LinkAttributes

    def __init__(self, input_dict: dict):
        self.trace_id = input_dict["trace_id"]
        self.id = input_dict["id"]
        self.attributes = LinkAttributes(input_dict["attributes"])


@dataclass
class Sql:
    """
    Sql
    
    Parameters
    ----------
    connection_string : str, optional
        For SQL Server or other database connections that don't use URL connection strings, record the connection string, excluding passwords
    url : str, optional
        For a database connection that uses a URL connection string, record the URL, excluding passwords
    sanitized_query : str, optional
        The database query, with any user provided values removed or replaced by a placeholder
    database_type : str, optional
        The name of the database engine
    database_version : str, optional
        The version number of the database engine
    driver_version : str, optional
        The name and version number of the database engine driver that your application uses
    user : str, optional
        The database username
    preparation : str, optional
        call if the query used a PreparedCall; statement if the query used a PreparedStatement
    """
    connection_string: Optional[str]
    url: Optional[str]
    sanitized_query: Optional[str]
    database_type: Optional[str]
    database_version: Optional[str]
    driver_version: Optional[str]
    user: Optional[str]
    preparation: Optional[str]

    def __init__(self, input_dict: dict):
        self.connection_string = input_dict.get("connection_string")
        self.url = input_dict.get("url")
        self.sanitized_query = input_dict.get("sanitized_query")
        self.database_type = input_dict.get("database_type")
        self.database_version = input_dict.get("database_version")
        self.driver_version = input_dict.get("driver_version")
        self.user = input_dict.get("user")
        self.preparation = input_dict.get("preparation")


@dataclass
class Response:
    """
    Response
    
    Parameters
    ----------
    status : int, optional
        Integer indicating the HTTP status of the response
    content_lenght: int, optional
        Integer indicating the length of the response body in bytes
    """
    status: Optional[int]
    content_length: Optional[int]

    def __init__(self, input_dict: Dict):
        self.status = input_dict.get("status")
        self.content_length = input_dict.get("content_length")


@dataclass
class Request:
    """
    Request
    
    Parameters
    ----------
    method : str, optional
        The request method. For example, GET
    client_ip : str, optional
        The IP address of the requester
    url : str, optional
        The full URL of the request, compiled from the protocol, hostname, and path of the request
    user_agent : str, optional
        The user agent string from the requester's client
    x_forwarded_for : bool, optional
        (segments only) boolean indicating that the client_ip was read from an X-Forwarded-For header and is not reliable as it could have been forged
    traced : bool, optional
        (subsegments only) boolean indicating that the downstream call is to another traced service
    """
    method: Optional[str]
    client_ip: Optional[str]
    url: Optional[str]
    user_agent: Optional[str]
    x_forwarded_for: Optional[bool]
    traced: Optional[bool]
    def __init__(self, input_dict: dict):
        self.method = input_dict.get("method")
        self.client_ip = input_dict.get("client_ip")
        self.url = input_dict.get("url")
        self.user_agent = input_dict.get("user_agent")
        self.x_forwarded_for = input_dict.get("x_forwarded_for")
        self.traced = input_dict.get("traced")


@dataclass
class Http:
    """
    Http
    
    Parameters
    ----------
    request : Request, optional
        Information about a request
    response : Response, optional
        Information about a response
    """
    request: Optional[Request]
    response: Optional[Response]
    def __init__(self, input_dict: dict):
        self.request = Request(input_dict["request"]) if "request" in input_dict else None
        self.response = Response(input_dict["response"]) if "response" in input_dict else None


@dataclass
class StackFrame:
    """
    Stack Frame
    
    Parameters
    ----------
    path : str, optional
        The relative path to the file
    line : int, optional
        The line in the file
    label : str, optional
        The function or method name
    """
    path: Optional[str]
    line: Optional[int]
    label: Optional[str]
    def __init__(self, input_dict: dict):
        self.path = input_dict.get("path")
        self.line = input_dict.get("line")
        self.label = input_dict.get("label")

@dataclass
class Exception:
    """
    Exception
    
    Parameters
    ----------
    id : str, optional
        A 64-bit identifier for the exception, unique among segments in the same trace, in 16 hexadecimal digits
    message : str, optional
        The exception message
    type : str, optional
        The exception type
    remote : bool, optional
        boolean indicating that the exception was caused by an error returned by a downstream service
    truncated : int, optional
        integer indicating the number of stack frames that are omitted from the stack
    skipped : int, optional
        integer indicating the number of exceptions that were skipped between this exception and its child, that is, the exception that it caused
    cause : str, optional
        Exception ID of the exception's parent, that is, the exception that caused this exception
    stack : List[StackFrame], optional
        List of StackFrame
    """
    id: Optional[str]
    message: Optional[str]
    type: Optional[str]
    remote: Optional[bool]
    truncated: Optional[int]
    skipped: Optional[int]
    cause: Optional[str]
    stack: Optional[List[StackFrame]]
    def __init__(self, input_dict: dict):
        self.id = input_dict.get("id")
        self.message = input_dict.get("message")
        self.type = input_dict.get("type")
        self.remote = input_dict.get("remote")
        self.truncated = input_dict.get("truncated")
        self.skipped = input_dict.get("skipped")
        self.cause = input_dict.get("cause")
        self.stack = [StackFrame(sf) for sf in input_dict["stack"]] if "stack" in input_dict else None


@dataclass
class Cause:
    """
    Cause
    
    Parameters
    ----------
    working_directory : str
        The full path of the working directory when the exception occurred
    paths: List[str]
        List of paths to libraries or modules in use when the exception occurred
    exceptions: List[Exception]
        List of Exception object
    """
    workding_directory: str
    paths: List[str]
    exceptions: List[Exception]
    def __init__(self, input_dict: dict):
        self.workding_directory = input_dict.get("working_directory", "")
        self.paths = [p for p in input_dict.get("paths", [])]
        self.exceptions = [Exception(e) for e in input_dict.get("exceptions", [])]


@dataclass
class Service:
    """
    Service
    
    Parameters
    ----------
    type: str, optional
        Type of the service
    """
    type: Optional[str]
    def __init__(self, input_dict: dict):
        self.type = input_dict.get("type")


@dataclass
class Subsegment:
    """
    Subsegment
    
    Parameters
    ----------
    id : str
        A 64-bit identifier for the subsegment, unique among segments in the same trace
    name : str
        The logical name of the subsegment
    trace_id : str
        Trace ID of the subsegment's parent segment
    start_time : float
        The time this subsegment was created
    end_time : float
        The time this subsegment was closed
    in_progress : bool
        Set to true instead of specifying an end_time to record that a subsegment is started, but is not complete
    parent_id : str
        Segment ID of the subsegment's parent segment
    type : str, optional
        Must be "subsegment"
    error : bool, optional
        Boolean indicating that a client error occurred (response status code was 4XX Client Error)
    throttle : bool, optional
        Boolean indicating that a request was throttled (response status code was 429 Too Many Requests)
    fault : bool, optional
        Boolean indicating that a server error occurred (response status code was 5XX Server Error)
    http : Http, optional
        Information about an outgoing HTTP call
    sql : Sql, optional
        Information about a SQL query
    namespace : str, optional
        "aws" for AWS SDK calls; "remote" for other downstream calls
    traced : bool, optional
        boolean indicating that the downstream call is to another traced service.
    precursor_ids : List[str], optional
        List of subsegment IDs that identifies subsegments with the same parent that completed prior to this subsegment
    cause : Union[str, Cause], optional
        Either a 16 character exception ID or an object with information about the cause
    aws : dict, optional
        Information about the downstream AWS resource that your application called
    annotations : dict, optional
        Key-value pairs that you want X-Ray to index for search
    metadata : dict, optional
        Any additional data that you want to store in the subsegment
    subsegments : List[Subsegment], optional
        List of subsegments
    links : List[Link], optional
        List of linked traces
    """
    id: str
    name: str
    trace_id: str
    start_time: float
    end_time: Optional[float]
    in_progress: Optional[bool]
    parent_id: str
    type: Optional[str]
    error: Optional[bool]
    throttle: Optional[bool]
    fault: Optional[bool]
    http: Optional[Http]
    sql: Optional[Sql]
    namespace: Optional[str]
    traced: Optional[bool]
    precursor_ids: Optional[List[str]]
    cause: Optional[Union[str, Cause]]
    aws: Optional[dict]
    annotations: Optional[dict]
    metadata: Optional[dict]
    subsegments: Optional[List["Subsegment"]]
    links: Optional[List[Link]]
    def __init__(self, input_dict:dict):
        self.id = input_dict.get("id", "")
        self.name = input_dict.get("name", "")
        self.trace_id = input_dict.get("trace_id", "")
        self.start_time = input_dict.get("start_time", 0)
        self.end_time = input_dict.get("end_time")
        self.in_progress = input_dict.get("in_progress")
        self.parent_id = input_dict.get("parent_id")
        self.type = input_dict.get("type")
        self.error = input_dict.get("error")
        self.throttle = input_dict.get("throttle")
        self.fault = input_dict.get("fault")
        self.http = Http(input_dict["http"]) if "http" in input_dict else None
        self.sql = Sql(input_dict["sql"]) if "sql" in input_dict else None
        self.namespace = input_dict.get("namespace")
        self.traced = input_dict.get("traced")
        self.precursor_ids = input_dict.get("precursor_ids")
        self.cause = None
        cause = input_dict.get("cause")
        if cause:
            self.cause = cause if isinstance(cause, str) else Cause(cause)
        self.aws = input_dict.get("aws")
        self.annotations = input_dict.get("annotations")
        self.metadata = input_dict.get("metadata")
        self.subsegments = [self.__class__(s) for s in input_dict["subsegments"]] if "subsegments" in input_dict else None
        self.links = [Link(l) for l in input_dict["links"]] if "links" in input_dict else None

    def __eq__(self, __value: "Subsegment") -> bool:
        return self.id == __value.id and self.trace_id == __value.trace_id


@dataclass
class Segment:
    """
    Segment
    
    Parameter
    ---------
    id : str
        A 64-bit identifier for the segment, unique among segments in the same trace
    name : str
        The logical name of the service that handled the request
    trace_id : str
        A unique identifier that connects all segments and subsegments originating from a single client request
    start_time : float
        The time this segment was created
    end_time : float, optional
        The time this segment was closed
    in_progress : bool, optional
        Set to True if a segment is started, but is not complete
    service : Service, optional
        Information about your service
    user : str, optional
        A string that identifies the user who sent the request
    origin : str, optional
        The type of AWS resource running your application
    parent_id : str, optional
        A subsegment ID you specify if the request originated from an instrumented application
    http : Http, optional
        Information about the original HTTP request
    error : bool, optional
        Boolean indicating that a client error occurred (response status code was 4XX Client Error)
    throttle : bool, optional
        Boolean indicating that a request was throttled (response status code was 429 Too Many Requests)
    fault : bool, optional
        Boolean indicating that a server error occurred (response status code was 5XX Server Error)
    cause : Union[str, Cause], optional
        Either a 16 character exception ID or an object with information about the cause
    aws : dict, optional
        Information about the AWS resource on which your application served the request
    annotations : dict, optional
        Object with key-value pairs that you want X-Ray to index for search
    metadata : dict, optional
        Object with any additional data that you want to store in the segment
    subsegments : List[Subsegment]
        List of subsegments
    links : List[Link]
        List of linked traces
    """
    id: str
    name: str
    trace_id: str
    start_time: float
    end_time: Optional[float]
    in_progress: Optional[bool]
    service: Optional[Service]
    user: Optional[str]
    origin: Optional[str]
    parent_id: Optional[str]
    http: Optional[Http]
    error: Optional[bool]
    throttle: Optional[bool]
    fault: Optional[bool]
    cause: Optional[Union[str, Cause]]
    aws: Optional[dict]
    annotations: Optional[dict]
    metadata: Optional[dict]
    subsegments: Optional[List["Subsegment"]]
    links: Optional[List[Link]]
    def __init__(self, input_dict:dict):
        self.id = input_dict.get("id", "")
        self.name = input_dict.get("name", "")
        self.trace_id = input_dict.get("trace_id", "")
        self.start_time = input_dict.get("start_time", 0)
        self.end_time = input_dict.get("end_time")
        self.in_progress = input_dict.get("in_progress")
        self.service = input_dict.get("service")
        self.user = input_dict.get("user")
        self.origin = input_dict.get("origin")
        self.parent_id = input_dict.get("parent_id")
        self.http = Http(input_dict["http"]) if "http" in input_dict else None
        self.error = input_dict.get("error")
        self.throttle = input_dict.get("throttle")
        self.fault = input_dict.get("fault")
        self.cause = None
        cause = input_dict.get("cause")
        if cause:
            self.cause = cause if isinstance(cause, str) else Cause(cause)
        self.aws = input_dict.get("aws")
        self.annotations = input_dict.get("annotations")
        self.metadata = input_dict.get("metadata")
        self.subsegments = [self.__class__(s) for s in input_dict["subsegments"]] if "subsegments" in input_dict else None
        self.links = [Link(l) for l in input_dict["links"]] if "links" in input_dict else None

    def __eq__(self, __value: "Segment") -> bool:
        return self.id == __value.id and self.trace_id == __value.trace_id


@dataclass
class Trace:
    """
    X-Ray Trace
    
    Parameters
    ----------
    id : str
        Trace ID
    duration : float
        Duration of the trace
    limit_exceeded : bool
        Boolean value indicating whether the trace has exceede its size limit
    segments: List[Segment]
        List of segments
    """
    id: str
    duration: float
    limit_exceeded: bool
    segments: List[Segment]
    def __init__(self, input_dict: dict):
        self.id = input_dict.get("id", "")
        self.duration = input_dict.get("duration", 0)
        self.limit_exceeded = input_dict.get("limit_exceeded", False)
        self.segments = [Segment(s) for s in input_dict.get("segments", [])]

    def __eq__(self, __value: "Trace") -> bool:
        return self.id == __value.id


@dataclass
class Tree:
    """
    X-Ray Trace Tree

    Parameters
    ----------
    root : Segment
        The root Segment of the Trace and child Traces
    paths : List[List[Segment]]
        List of List of Segments for all paths from Root Segment to each Leaf Segment
    source_trace : Trace
        The Trace containing the Root Segment
    child_traces: Dict[str, Trace]
        Map of child Trace ID to child Trace
    """
    root: Segment
    paths: List[List[Segment]]
    source_trace: Trace
    child_traces: Dict[str, Trace]
    def __init__(self, input_dict: dict):
        self.root = Segment(input_dict["root"])
        self.paths = [[Segment(s) for s in path] for path in input_dict["paths"]]
        self.source_trace = Trace(input_dict["source_trace"])
        if input_dict.get("child_traces", {}):
            self.child_traces = {
                cid: Trace(child) for cid, child in input_dict.get("child_traces", {}).items()
            }
        else:
            self.child_traces = {}