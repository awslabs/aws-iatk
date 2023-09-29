import logging
import json
import pathlib
import time
from unittest import TestCase

import boto3
import zion

def read_cdk_outputs() -> dict:
    with open(pathlib.Path(__file__).parent.parent.parent / "outputs.json") as f:
        outputs = json.load(f)
    return outputs

def test_event_generation():
    stack_name = "cdk-example-schemaStack"
    stack_outputs = read_cdk_outputs().get(stack_name, {})
    registry_name = stack_outputs["RegistryName"]
    schema_name = stack_outputs["SchemaName"]
    function_name = stack_outputs["CalculatorFunction"]
    z = zion.Zion()
    barebone_event = z.generate_mock_event(
        zion.GenerateMockEventParams(
            registry_name=registry_name,
            schema_name=schema_name,
            event_ref="MyEvent",
        )
    ).event
    for key in ["address", "customerId", "datetime", "membershipType", "orderItems"]:
        assert key in barebone_event
    assert barebone_event["address"] == ""
    assert barebone_event["customerId"] == ""
    assert barebone_event["orderItems"] == []
