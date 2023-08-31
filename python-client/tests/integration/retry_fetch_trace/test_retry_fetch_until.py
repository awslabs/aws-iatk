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
from zion import RetryFetchXRayTraceUntilParams
from aws_xray_sdk.core.models.traceid import TraceId

LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="zion", level=logging.DEBUG)

class TestZion_retry_fetch_until(TestCase):
    counter = 0
    xray_trace_id = ""
    dummy_trace_id = TraceId().to_id()
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

    # @pytest.mark.flaky(reruns=3)
    # def test_get_traces_pass(self):
    #     time.sleep(5)
    #     def num_is_10(trace):
    #         self.counter = random.randrange(0,11)
    #         return self.counter == 10
    #     params = RetryFetchXRayTraceUntilParams(
    #         trace_id=self.xray_trace_id,
    #         condition=num_is_10,
    #         timeout_seconds=10)
    #     start = time.time()
    #     response = self.zion.retry_fetch_trace_until(params=params)
    #     end = time.time()
    #     self.assertTrue(response)
    #     self.assertEqual(self.counter, 10)
    #     self.assertLess(end - start, 10)
    
    # @pytest.mark.flaky(reruns=3)
    # def test_get_traces_fail(self):
    #     time.sleep(5)
    #     def num_is_10(trace):
    #         self.counter = random.randrange(0,10)
    #         return self.counter == 10
    #     params = RetryFetchXRayTraceUntilParams(
    #         trace_id=self.xray_trace_id,
    #         condition=num_is_10,
    #         timeout_seconds=10)
    #     start = time.time()
    #     response = self.zion.retry_fetch_trace_until(params=params)
    #     end = time.time()
    #     self.assertFalse(response)
    #     self.assertNotEqual(self.counter, 10)
    #     self.assertGreater(end - start, 10)

    # def test_unsampled_traceid_fail(self):
    #     def num_is_10(trace):
    #         self.counter = random.randrange(0,10)
    #         return self.counter == 10
    #     params = RetryFetchXRayTraceUntilParams(
    #         trace_id=self.dummy_trace_id,
    #         condition=num_is_10,
    #         timeout_seconds=10)
    #     start = time.time()
    #     with pytest.raises(IndexError) as e:  
    #         self.zion.retry_fetch_trace_until(params=params)
    #         end = time.time()
    #         self.assertNotEqual(self.counter, 10)
    #         self.assertGreater(end - start, 10)
    #         self.assertEqual(e, "trace id must be sampled and exist on aws")

    # @pytest.mark.timeout(timeout=2500, method="thread")
    # def test_get_traces_infinite_timeout_pass(self):
    #     time.sleep(5)
    #     def num_is_10(trace):
    #         time.sleep(1.5)
    #         self.counter += 1
    #         return self.counter == 10
    #     params = RetryFetchXRayTraceUntilParams(
    #         trace_id=self.xray_trace_id,
    #         condition=num_is_10,
    #         timeout_seconds=0)
    #     start = time.time()
    #     response = self.zion.retry_fetch_trace_until(params=params)
    #     end = time.time()
    #     self.assertTrue(response)
    #     self.assertEqual(self.counter, 10)
    #     self.assertGreater(end - start, 15)

    def test_invalid_traceid_fail(self):
        def num_is_10(trace):
            self.counter = random.randrange(0,10)
            return self.counter == 10
        params = RetryFetchXRayTraceUntilParams(
            trace_id="test",
            condition=num_is_10,
            timeout_seconds=10)
        start = time.time()
        with pytest.raises(IndexError) as e:  
            self.zion.retry_fetch_trace_until(params=params)
            end = time.time()
            self.assertNotEqual(self.counter, 10)
            self.assertGreater(end - start, 10)
            self.assertEqual(e, "trace id must be sampled and exist on aws")