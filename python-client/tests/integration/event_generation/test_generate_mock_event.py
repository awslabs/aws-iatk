# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Integration tests for zion.generate_mock_event
"""
import json
import logging
from uuid import uuid4
from unittest import TestCase
from dataclasses import dataclass
from zion import Zion, context_generation, ZionException
import time
import os
import boto3
import pytest
import uuid


LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="zion", level=logging.DEBUG)

class TestZion_generate_mock_event(TestCase):
    zion = Zion(region="us-east-1")
    cfn_client = boto3.client("cloudformation", region_name="us-east-1")
    test_stack_name = "testMockEventStack" + str(uuid4())
    schema_details = {}
    @classmethod
    def setUpClass(cls) -> None:
        LOG.debug("creating stack")
        try:
            current_path = os.path.realpath(__file__)
            waiter = cls.cfn_client.get_waiter('stack_create_complete')
            current_dir = os.path.dirname(current_path)
            test_stack_path = os.path.join(current_dir, "testdata","test_stack.yaml")
            with open(test_stack_path, 'r') as content_file:
                content = content_file.read()
            cls.cfn_client.create_stack(
                StackName=cls.test_stack_name,
                TemplateBody=content
            )
            waiter.wait(
                StackName=cls.test_stack_name,
                WaiterConfig={
                    "Delay": 3,
                    "MaxAttempts": 10
                })
            output = cls.cfn_client.describe_stacks(StackName=cls.test_stack_name)
            for detail in output["Stacks"][0]["Outputs"]:
                cls.schema_details[detail["OutputKey"]] = detail["OutputValue"]
        except Exception as e:
            LOG.debug(e)


    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove stack")
        try:
            waiter = cls.cfn_client.get_waiter('stack_delete_complete')
            cls.cfn_client.delete_stack(StackName=cls.test_stack_name)
            waiter.wait(
                StackName=cls.test_stack_name,
                WaiterConfig={
                    "Delay": 3,
                    "MaxAttempts": 10
                })
            LOG.debug("stack succesfully deleted")
        except Exception as e:
            LOG.debug(e)
        
    def test_json_schema_success_with_context(self):
        required_attributes = ["detail-type", "resources", "id", "source", "time", "detail", "region", "version", "account"]
        detail_attributes = ["actionOnFailure", "clusterId", "message", "name", "severity", "state", "stepId"]
        output = self.zion.generate_mock_event(
            registry_name=self.schema_details["TestSchemaRegistryName"],
            schema_name=self.schema_details["TestEBEventSchemaJSONSchemaName"],
            schema_version=self.schema_details["TestEBEventSchemaJSONSchemaVersion"],
            skip_optional=False,
            contexts=[context_generation.eventbridge_event_context]
        ).event
        for attribute in required_attributes:
            self.assertIn(attribute, output)
        self.assertIn("detail", output)
        for detail in detail_attributes:
            self.assertIn(detail, output["detail"])
        self.assertEqual(len(output["account"]), 12)
        self.assertEqual("us-east-1", output["region"])

    def test_json_schema_success_default(self):
        required_attributes = ["detail-type", "resources", "id", "source", "time", "detail", "region", "version", "account"]
        detail_attributes = ["actionOnFailure", "clusterId", "message", "name", "severity", "state", "stepId"]
        output = self.zion.generate_mock_event(
            registry_name=self.schema_details["TestSchemaRegistryName"],
            schema_name=self.schema_details["TestEBEventSchemaJSONSchemaName"],
            schema_version=self.schema_details["TestEBEventSchemaJSONSchemaVersion"],
            skip_optional=False,
            contexts=[]
        ).event
        for attribute in required_attributes:
            self.assertIn(attribute, output)
        self.assertIn("detail", output)
        for detail in detail_attributes:
            self.assertIn(detail, output["detail"])
        self.assertEqual(output["account"], "")
        self.assertEqual(output["region"], "")
        
    def test_json_schema_success_required(self):
        required_attributes = ["detail-type", "detail", "region"]
        not_required_attributes = ["source", "id", "version", "account", "time"]
        detail_attributes = ["actionOnFailure", "clusterId", "message", "name", "severity", "state", "stepId"]
        output = self.zion.generate_mock_event(
            registry_name=self.schema_details["TestSchemaRegistryName"],
            schema_name=self.schema_details["TestEBEventSchemaJSONSchemaName"],
            schema_version=self.schema_details["TestEBEventSchemaJSONSchemaVersion"],
            skip_optional=True,
            contexts=[]
        ).event
        for attribute in required_attributes:
            self.assertIn(attribute, output)
        for attribute in not_required_attributes:
            self.assertNotIn(attribute, output)
        self.assertIn("detail", output)
        for detail in detail_attributes:
            self.assertIn(detail, output["detail"])
        self.assertEqual(output["region"], "")

    def test_json_schema_success_with_overrides(self):
        def override(event):
            event["account"] = "testid"
            event["newKey"] = "test"
            event["region"] = "testRegion"
            return event
        required_attributes = ["detail-type", "resources", "id", "source", "time", "detail", "region", "version", "account"]
        detail_attributes = ["actionOnFailure", "clusterId", "message", "name", "severity", "state", "stepId"]
        output = self.zion.generate_mock_event(
            registry_name=self.schema_details["TestSchemaRegistryName"],
            schema_name=self.schema_details["TestEBEventSchemaJSONSchemaName"],
            schema_version=self.schema_details["TestEBEventSchemaJSONSchemaVersion"],
            skip_optional=False,
            contexts=[context_generation.eventbridge_event_context, override]
        ).event
        for attribute in required_attributes:
            self.assertIn(attribute, output)
        self.assertIn("detail", output)
        for detail in detail_attributes:
            self.assertIn(detail, output["detail"])
        self.assertEqual(output["account"], "testid")
        self.assertEqual(output["newKey"], "test")
        self.assertEqual(output["region"], "testRegion")
    
    def test_openapi_schema_success_with_context(self):
        required_attributes = ["detail-type", "resources", "id", "source", "time", "detail", "region", "version", "account"]
        detail_attributes = ["creator", "department", "ticketId"]
        output = self.zion.generate_mock_event(
            registry_name=self.schema_details["TestSchemaRegistryName"],
            schema_name=self.schema_details["TestEBEventSchemaOpenAPIName"],
            schema_version=self.schema_details["TestEBEventSchemaOpenAPIVersion"],
            skip_optional=False,
            contexts=[context_generation.eventbridge_event_context],
            event_ref="MyEvent"
        ).event
        for attribute in required_attributes:
            self.assertIn(attribute, output)
        self.assertIn("detail", output)
        for detail in detail_attributes:
            self.assertIn(detail, output["detail"])
        self.assertEqual(len(output["account"]), 12)
        self.assertEqual("ap-south-1", output["region"])

    def test_openapi_schema_success_default(self):
        required_attributes = ["detail-type", "resources", "id", "source", "time", "detail", "region", "version", "account"]
        detail_attributes = ["creator", "department", "ticketId"]
        output = self.zion.generate_mock_event(
            registry_name=self.schema_details["TestSchemaRegistryName"],
            schema_name=self.schema_details["TestEBEventSchemaOpenAPIName"],
            schema_version=self.schema_details["TestEBEventSchemaOpenAPIVersion"],
            skip_optional=False,
            contexts=[],
            event_ref="MyEvent"
        ).event
        for attribute in required_attributes:
            self.assertIn(attribute, output)
        self.assertIn("detail", output)
        for detail in detail_attributes:
            self.assertIn(detail, output["detail"])
        self.assertEqual(output["account"], "")
        self.assertEqual(output["region"], "ap-south-1")
        
    def test_openapi_schema_success_required(self):
        required_attributes = ["detail-type", "detail", "region"]
        not_required_attributes = ["source", "id", "version", "account", "time"]
        detail_attributes = ["creator", "department", "ticketId"]
        output = self.zion.generate_mock_event(
            registry_name=self.schema_details["TestSchemaRegistryName"],
            schema_name=self.schema_details["TestEBEventSchemaOpenAPIName"],
            schema_version=self.schema_details["TestEBEventSchemaOpenAPIVersion"],
            skip_optional=True,
            contexts=[],
            event_ref="MyEvent"
        ).event
        for attribute in required_attributes:
            self.assertIn(attribute, output)
        for attribute in not_required_attributes:
            self.assertNotIn(attribute, output)
        self.assertIn("detail", output)
        for detail in detail_attributes:
            self.assertIn(detail, output["detail"])

    def test_openapi_schema_success_with_overrides(self):
        def override(event):
            event["account"] = "testid"
            event["newKey"] = "test"
            event["region"] = "testRegion"
            return event
        required_attributes = ["detail-type", "resources", "id", "source", "time", "detail", "region", "version", "account"]
        detail_attributes = ["creator", "department", "ticketId"]
        output = self.zion.generate_mock_event(
            registry_name=self.schema_details["TestSchemaRegistryName"],
            schema_name=self.schema_details["TestEBEventSchemaOpenAPIName"],
            schema_version=self.schema_details["TestEBEventSchemaOpenAPIVersion"],
            skip_optional=False,
            contexts=[context_generation.eventbridge_event_context, override],
            event_ref="MyEvent"
        ).event
        for attribute in required_attributes:
            self.assertIn(attribute, output)
        self.assertIn("detail", output)
        for detail in detail_attributes:
            self.assertIn(detail, output["detail"])
        self.assertEqual(output["account"], "testid")
        self.assertEqual(output["newKey"], "test")
        self.assertEqual(output["region"], "testRegion")

    def test_openapi_ref_failure(self):
        with pytest.raises(ZionException) as e:
            self.zion.generate_mock_event(
                registry_name=self.schema_details["TestSchemaRegistryName"],
                schema_name=self.schema_details["TestEBEventSchemaOpenAPIName"],
                schema_version=self.schema_details["TestEBEventSchemaOpenAPIVersion"],
                skip_optional=False,
                contexts=[],
                event_ref=""
            ).event
        self.assertEqual(str(e.value), "error generating mock event: error generating mock event: no eventRef specified to generate a mock event")
