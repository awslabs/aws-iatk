# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Integration tests for zion.retry_until
"""
import logging
from unittest import TestCase
import time
import boto3
import random
import zion
import os
import pytest
import json
from ....src.zion.retry_xray_trace import (
    RetryFetchXRayTraceUntilParams,
)

LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="zion", level=logging.DEBUG)

class TestZion_retry_fetch_until(TestCase):
    xray_trace_id = ""
    zion = zion.Zion()
    lambda_client = boto3.client("lambda")
    iam_client = boto3.client("iam")
    xray_client = boto3.client("xray")
    lambda_function_name = "test_lambda" + str(random.randrange(0,100000))

    @classmethod
    def setUpClass(cls) -> None:
        LOG.debug("creating resources")
        try:
            cls.lambda_client = cls.zion.patch_aws_client(cls.lambda_client, 1)
            current_path = os.path.realpath(__file__)
            current_dir = os.path.dirname(current_path)
            test_lambda_path = os.path.join(current_dir, "testdata","helloworld.zip")
            with open(test_lambda_path, "rb") as f:
                zipped_code = f.read()
            time.sleep(10)
            role = cls.iam_client.get_role(
                RoleName="xray-integration-role"
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
            time.sleep(5)
            response = cls.lambda_client.invoke(
            FunctionName=cls.lambda_function_name,
            Payload='{ "key": "value" }'
            )
            cls.xray_trace_id = response["ResponseMetadata"]["HTTPHeaders"]["x-amzn-trace-id"].split(";")[0].split("root=")[1]
        except Exception as e:
            LOG.debug(e)

    
    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove lambda function")
        cls.lambda_client.delete_function(
            FunctionName=cls.lambda_function_name
        )

    def test_get_traces(self):
        def trace_is
        params = RetryFetchXRayTraceUntilParams(trace_id=self.xray_trace_id)
        time.sleep(3)
        print(self.xray_trace_id)
        response = self.xray_client.batch_get_traces(TraceIds=[self.xray_trace_id])
        print(json.dumps(response, indent=2))
        self.assertEqual(0,1)