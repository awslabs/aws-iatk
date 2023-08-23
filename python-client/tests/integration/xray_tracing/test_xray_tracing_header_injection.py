# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Integration tests for zion.patch_aws_client
"""
import json
import logging
from uuid import uuid4
from unittest import TestCase
from typing import List, Dict, Callable
from dataclasses import dataclass
from zion import Zion
from pathlib import Path
import time
import os
import boto3
from botocore.exceptions import ClientError
from parameterized import parameterized
import pytest



LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="zion", level=logging.DEBUG)

class TestZion_xray_tracing_lambda(TestCase):
    zion = Zion("region=us-east-1")
    lambda_client = boto3.client("lambda")
    iam_client = boto3.client("iam")
    role_policy = {
    "Version": "2012-10-17",
    "Statement": [
        {
        "Sid": "",
        "Effect": "Allow",
        "Principal": {
            "Service": "lambda.amazonaws.com"
        },
        "Action": "sts:AssumeRole"
        }
    ]
    }
    def check_lambda_function_exists(function):
        return function["FunctionName"] == "test_lambda"

    @classmethod
    def setUpClass(cls) -> None:
        LOG.debug("creating resources")
        try:
            cls.iam_client.create_role(
                RoleName="test_lambda_basic_execution",
                AssumeRolePolicyDocument=json.dumps(cls.role_policy),
            )
        except Exception as e:
            LOG.debug(e)
        try:
            current_path = os.path.realpath(__file__)
            current_dir = os.path.dirname(current_path)
            test_lambda_path = os.path.join(current_dir, "testdata","helloworld.zip")
            with open(test_lambda_path, "rb") as f:
                zipped_code = f.read()
            time.sleep(10)
            role = cls.iam_client.get_role(
                RoleName="test_lambda_basic_execution"
                )
            response = cls.lambda_client.list_functions()
            filtered = filter(cls.check_lambda_function_exists,response["Functions"])
            if len(list(filtered)) == 0:
                LOG.debug("creating lambda function")
                cls.lambda_client.create_function(
                    FunctionName="test_lambda",
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
            FunctionName="test_lambda"
        )
        LOG.debug("delete role")
        cls.iam_client.delete_role(
            RoleName="test_lambda_basic_execution"
        )
    
    def test_sampled_xray_trace_lambda(self):
        time.sleep(3)
        self.zion.patch_aws_client(self.lambda_client, 1)
        status = self.lambda_client.get_function(
            FunctionName="test_lambda"
        )["Configuration"]["State"]
        LOG.debug(status)
        while status != "Active":
            time.sleep(1)
            status = self.lambda_client.get_function(
            FunctionName="test_lambda"
            )
            LOG.debug(status)
        response = self.lambda_client.invoke(
            FunctionName='test_lambda',
            Payload='{ "key": "value" }'
        )
        LOG.debug(response)
        self.assertIn("x-amzn-trace-id",response["ResponseMetadata"]["HTTPHeaders"])
        trace_id = response["ResponseMetadata"]["HTTPHeaders"]["x-amzn-trace-id"]
        sampled_string = trace_id.split(";")[1]
        self.assertEqual(sampled_string[len(sampled_string) - 1],"1")

    def test_unsampled_xray_trace_lambda(self):
        time.sleep(3)
        self.zion.patch_aws_client(self.lambda_client, 0)
        status = self.lambda_client.get_function(
            FunctionName="test_lambda"
        )["Configuration"]["State"]
        LOG.debug(status)
        while status != "Active":
            time.sleep(1)
            status = self.lambda_client.get_function(
            FunctionName="test_lambda"
            )
            LOG.debug(status)
        response = self.lambda_client.invoke(
            FunctionName='test_lambda',
            Payload='{ "key": "value" }'
        )
        LOG.debug(response)
        self.assertIn("x-amzn-trace-id",response["ResponseMetadata"]["HTTPHeaders"])
        trace_id = response["ResponseMetadata"]["HTTPHeaders"]["x-amzn-trace-id"]
        sampled_string = trace_id.split(";")[1]
        self.assertEqual(sampled_string[len(sampled_string) - 1],"0")


class TestZion_xray_tracing_sfn(TestCase):
    zion = Zion("region=us-east-1")
    sfn_client = boto3.client("stepfunctions")
    iam_client = boto3.client("iam")
    state_machin_arn = None
    role_policy = {
        "Version": "2012-10-17",
        "Statement": [
            {
            "Sid": "",
            "Effect": "Allow",
            "Principal": {
                "Service": "states.amazonaws.com"
            },
            "Action": "sts:AssumeRole"
            }
        ]
        }
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
    @classmethod
    def setUpClass(cls) -> None:
        LOG.debug("creating resources")
        try:
            cls.iam_client.create_role(
                RoleName="test_sfn_execution",
                AssumeRolePolicyDocument=json.dumps(cls.role_policy),
            )
        except Exception as e:
            LOG.debug(e)
        try:
            role = cls.iam_client.get_role(
                RoleName="test_sfn_execution"
            )
            create_state_machine_response = cls.sfn_client.create_state_machine(
                name="test_xray_tracing_state_machine",
                roleArn=role["Role"]["Arn"],
                definition=json.dumps(cls.definition),
                tracingConfiguration={
                    "enabled":True
                }
            )
            cls.state_machin_arn = create_state_machine_response["stateMachineArn"]
        except Exception as e:
            LOG.debug(e)
    
    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove step functions state machine")
        LOG.debug(cls.state_machin_arn)
        cls.sfn_client.delete_state_machine(
            stateMachineArn=cls.state_machin_arn
        )
        LOG.debug("delete role")
        cls.iam_client.delete_role(
            RoleName="test_sfn_execution"
        )
    
    def test_sampled_xray_trace_sfn(self):
        time.sleep(3)
        self.zion.patch_aws_client(self.sfn_client, 1)
        
        start_response = self.sfn_client.start_execution(
            stateMachineArn=self.__class__.state_machin_arn,
            input='{"IsHelloWorldExample": true}'
        )
        describe_response = self.sfn_client.describe_execution(
            executionArn=start_response["executionArn"]
            )
        self.assertIn("traceHeader", describe_response)
        sampled_string = describe_response["traceHeader"].split(";")[1]
        self.assertEqual(sampled_string[len(sampled_string) - 1],"1")

    def test_unsampled_xray_trace_sfn(self):
        time.sleep(3)
        self.zion.patch_aws_client(self.sfn_client, 0)
        start_response = self.sfn_client.start_execution(
            stateMachineArn=self.__class__.state_machin_arn,
            input='{"IsHelloWorldExample": true}'
        )
        time.sleep(3)
        describe_response = self.sfn_client.describe_execution(
            executionArn=start_response["executionArn"]
            )
        self.assertIn("traceHeader", describe_response)
        sampled_string = describe_response["traceHeader"].split(";")[1]
        self.assertEqual(sampled_string[len(sampled_string) - 1],"0")



