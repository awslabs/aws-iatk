import logging
import json
import pathlib
import uuid

import boto3
import zion
from zion.context_generation import eventbridge_event_context

LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
lambda_client = boto3.client("lambda")

def read_cdk_outputs() -> dict:
    with open(pathlib.Path(__file__).parent.parent.parent / "outputs.json") as f:
        outputs = json.load(f)
    return outputs

def test_generate_barebone_event():
    stack_name = "cdk-example-schemaStack"
    stack_outputs = read_cdk_outputs().get(stack_name, {})
    registry_name = stack_outputs["RegistryName"]
    schema_name = stack_outputs["SchemaName"]
    z = zion.Zion()
    barebone_event = z.generate_mock_event(
        registry_name=registry_name,
        schema_name=schema_name,
        event_ref="MyEvent",
    ).event
    LOG.debug(json.dumps(barebone_event, indent=2))
    for key in ["address", "customerId", "datetime", "membershipType", "orderItems"]:
        assert key in barebone_event
    assert barebone_event["address"] == ""
    assert barebone_event["customerId"] == ""
    assert barebone_event["orderItems"] == []
    
def test_generate_contextful_event():
    stack_name = "cdk-example-schemaStack"
    stack_outputs = read_cdk_outputs().get(stack_name, {})
    registry_name = stack_outputs["RegistryName"]
    schema_name = stack_outputs["SchemaName"]
    function_name = stack_outputs["CalculatorFunction"]
    z = zion.Zion()
    
    def apply_context(event: dict) -> dict:
        event["customerId"] = str(uuid.uuid4())
        event["address"] = "99 Some Street"
        for i in range(5):
            item = {
                "unitPrice": (i + 1) * 2,
                "count": i + 1, 
            }
            event["orderItems"].append(item)
        return event
        
    mock_event = z.generate_mock_event(
        registry_name=registry_name,
        schema_name=schema_name,
        event_ref="MyEvent",
        contexts=[apply_context],
    ).event
    LOG.debug(json.dumps(mock_event, indent=2))
    for key in ["address", "customerId", "datetime", "membershipType", "orderItems"]:
        assert key in mock_event
    assert mock_event["customerId"] != ""
    assert mock_event["address"] == "99 Some Street"
    assert len(mock_event["orderItems"]) > 0

    response = lambda_client.invoke(
        FunctionName=function_name,
        Payload=bytes(json.dumps(mock_event), encoding="utf-8"),
    )
    result = int(response['Payload'].read())
    assert result == 110

def test_generate_eventbridge_event():
    z = zion.Zion()

    mock_eb_event = z.generate_mock_event(
        registry_name="aws.events",
        schema_name="aws.autoscaling@EC2InstanceLaunchSuccessful",
        schema_version="2",
        event_ref="AWSEvent",
        contexts=[eventbridge_event_context],
    ).event
    LOG.debug(json.dumps(mock_eb_event, indent=2))
    for key in ["detail-type", "resources", "id", "source", "time", "detail", "region", "version", "account"]:
        assert key in mock_eb_event
    assert mock_eb_event["id"] != ""
    assert mock_eb_event["account"] != ""
    assert mock_eb_event["time"] != ""
