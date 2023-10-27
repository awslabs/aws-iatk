# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Integration tests for aws_ctk.patch_aws_client
"""
import json
import logging
from uuid import uuid4
from unittest import TestCase
from typing import List, Dict, Callable
from dataclasses import dataclass
from aws_ctk import AWSCtk
from pathlib import Path
import time
import os
import boto3
from botocore.exceptions import ClientError
from parameterized import parameterized
import pytest
import random



LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="aws_ctk", level=logging.DEBUG)

class TestCTK_xray_tracing_lambda(TestCase):
    ctk = AWSCtk(region="us-east-1")
    lambda_client = boto3.client("lambda")
    iam_client = boto3.client("iam")
    lambda_function_name = "test_lambda" + str(random.randrange(0,100000))
    @classmethod
    def setUpClass(cls) -> None:
        LOG.debug("creating resources")
        try:
            current_path = os.path.realpath(__file__)
            current_dir = os.path.dirname(current_path)
            test_lambda_path = os.path.join(current_dir, "testdata","helloworld.zip")
            with open(test_lambda_path, "rb") as f:
                zipped_code = f.read()
            time.sleep(10)
            role = cls.iam_client.get_role(
                RoleName="xray-integration-role-lambda"
                )
            LOG.debug("creating lambda function")
            cls.lambda_client.create_function(
                FunctionName=cls.lambda_function_name,
                Role=role["Role"]["Arn"],
                Runtime='python3.9',
                Handler='helloworld.handler',
                Code=dict(ZipFile=zipped_code),
                TracingConfig={'Mode': 'Active'}
            )
        except Exception as e:
            LOG.debug(e)


    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove lambda function")
        cls.lambda_client.delete_function(
            FunctionName=cls.lambda_function_name
        )
    @pytest.mark.timeout(timeout=1000, method="thread")
    def test_sampled_xray_trace_lambda(self):
        self.ctk.patch_aws_client(self.lambda_client, 1)
        status = self.lambda_client.get_function(
            FunctionName=self.lambda_function_name
        )["Configuration"]["State"]
        LOG.debug(status)
        while status != "Active":
            time.sleep(1)
            status = self.lambda_client.get_function(
            FunctionName=self.lambda_function_name
            )["Configuration"]["State"]
            LOG.debug(status)
        response = self.lambda_client.invoke(
            FunctionName=self.lambda_function_name,
            Payload='{ "key": "value" }'
        )
        LOG.debug(response)
        self.assertIn("x-amzn-trace-id",response["ResponseMetadata"]["HTTPHeaders"])
        trace_id = response["ResponseMetadata"]["HTTPHeaders"]["x-amzn-trace-id"]
        sampled_string = trace_id.split(";")[1]
        self.assertEqual(sampled_string[len(sampled_string) - 1],"1")

    @pytest.mark.timeout(timeout=1000, method="thread")
    def test_unsampled_xray_trace_lambda(self):
        self.ctk.patch_aws_client(self.lambda_client, 0)
        status = self.lambda_client.get_function(
            FunctionName=self.lambda_function_name
        )["Configuration"]["State"]
        LOG.debug(status)
        while status != "Active":
            time.sleep(1)
            status = self.lambda_client.get_function(
            FunctionName=self.lambda_function_name
            )["Configuration"]["State"]
            LOG.debug(status)
        response = self.lambda_client.invoke(
            FunctionName=self.lambda_function_name,
            Payload='{ "key": "value" }'
        )
        LOG.debug(response)
        self.assertIn("x-amzn-trace-id",response["ResponseMetadata"]["HTTPHeaders"])
        trace_id = response["ResponseMetadata"]["HTTPHeaders"]["x-amzn-trace-id"]
        sampled_string = trace_id.split(";")[1]
        self.assertEqual(sampled_string[len(sampled_string) - 1],"0")


class TestCTK_xray_tracing_sfn(TestCase):
    ctk = AWSCtk(region="us-east-1")
    sfn_client = boto3.client("stepfunctions")
    iam_client = boto3.client("iam")
    sfn_machine_name = "test_xray_tracing_state_machine" + str(random.randrange(0,100000))
    definition = {
        "Comment": "A Hello World example of the Amazon States Language using Pass states",
        "StartAt": "Hello",
        "States": {
            "Hello": {
            "Type": "Pass",
            "Result": "Hello",
            "Next": "World"
            },
            "World": {
            "Type": "Pass",
            "Result": "World",
            "End": True
            }
        }
        }
    def get_state_machine_arn(self, sfn_client):
        response = sfn_client.list_state_machines()["stateMachines"]
        for sfn in response:
            if sfn['name'] == self.sfn_machine_name:
                return sfn["stateMachineArn"]
    @classmethod
    def setUpClass(cls) -> None:
        LOG.debug("creating state machine")
        try:
            role = cls.iam_client.get_role(
                RoleName="xray-integration-role-sfn"
            )
            cls.sfn_client.create_state_machine(
                name=cls.sfn_machine_name,
                roleArn=role["Role"]["Arn"],
                definition=json.dumps(cls.definition),
                tracingConfiguration={
                    "enabled":True
                }
            )
        except Exception as e:
            LOG.debug(e)
    
    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove step functions state machine")
        cls.sfn_client.delete_state_machine(
            stateMachineArn=cls.get_state_machine_arn(cls, cls.sfn_client)
        )
    
    def test_sampled_xray_trace_sfn(self):
        time.sleep(3)
        self.ctk.patch_aws_client(self.sfn_client, 1)
        
        start_response = self.sfn_client.start_execution(
            stateMachineArn=self.get_state_machine_arn(self.sfn_client),
            input='{"IsHelloWorldExample": true}'
        )
        describe_response = self.sfn_client.describe_execution(
            executionArn=start_response["executionArn"]
            )
        LOG.debug(describe_response)
        self.assertIn("traceHeader", describe_response)
        sampled_string = describe_response["traceHeader"].split(";")[1]
        self.assertEqual(sampled_string[len(sampled_string) - 1],"1")

    def test_unsampled_xray_trace_sfn(self):
        time.sleep(3)
        self.ctk.patch_aws_client(self.sfn_client, 0)
        start_response = self.sfn_client.start_execution(
            stateMachineArn=self.get_state_machine_arn(self.sfn_client),
            input='{"IsHelloWorldExample": true}'
        )
        time.sleep(3)
        describe_response = self.sfn_client.describe_execution(
            executionArn=start_response["executionArn"]
            )
        LOG.debug(describe_response)
        self.assertIn("traceHeader", describe_response)
        sampled_string = describe_response["traceHeader"].split(";")[1]
        self.assertEqual(sampled_string[len(sampled_string) - 1],"0")



