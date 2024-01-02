# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Integration tests for aws_iatk.retry_until
"""
import logging
from unittest import TestCase
from unittest.mock import patch
import time
import boto3
import random
from aws_iatk import AwsIatk, IatkException
import os
import pytest
import json

LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="aws_iatk", level=logging.DEBUG)

class TestIatk_retry_fetch_until(TestCase):
    counter = 0
    xray_trace_header = ""
    dummy_trace_header = "Root=test,Sampled=1"
    region = "us-east-1"
    iatk = AwsIatk(region=region)
    lambda_client = boto3.client("lambda", region_name=region)
    iam_client = boto3.client("iam", region_name=region)
    lambda_function_name = "test_lambda" + str(random.randrange(0,100000))

    @classmethod
    def setUpClass(cls) -> None:
        LOG.debug("creating resources")
        try:
            cls.lambda_client = cls.iatk.patch_aws_client(cls.lambda_client, 1)
            current_path = os.path.realpath(__file__)
            current_dir = os.path.dirname(current_path)
            test_lambda_path = os.path.join(current_dir, "testdata","helloworld.zip")
            with open(test_lambda_path, "rb") as f:
                zipped_code = f.read()
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
            time.sleep(5)
            response = cls.lambda_client.invoke(
            FunctionName=cls.lambda_function_name,
            Payload='{ "key": "value" }'
            )
            cls.xray_trace_header = response["ResponseMetadata"]["HTTPHeaders"]["x-amzn-trace-id"]
        except Exception as e:
            LOG.debug(e)

    
    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove lambda function")
        cls.lambda_client.delete_function(
            FunctionName=cls.lambda_function_name
        )

    @pytest.mark.flaky(reruns=3)
    def test_get_traces_pass(self):
        time.sleep(5)
        def trace_header_is_root(tree):
            xray_trace_id = self.xray_trace_header.split(";")[0].split("=")[1]
            assert tree.trace_tree.root.name == self.lambda_function_name and tree.trace_tree.root.trace_id == xray_trace_id
        start = time.time()
        response = self.iatk.retry_get_trace_tree_until(
            tracing_header=self.xray_trace_header,
            assertion_fn=trace_header_is_root,
            timeout_seconds=10,
        )
        end = time.time()
        self.assertTrue(response)
        self.assertLess(end - start, 10)
    
    def test_get_traces_fail(self):
        time.sleep(5)
        def num_is_10(trace):
            self.counter = random.randrange(0,10)
            assert self.counter == 10
        start = time.time()
        response = self.iatk.retry_get_trace_tree_until(
            tracing_header=self.xray_trace_header,
            assertion_fn=num_is_10,
            timeout_seconds=10,
        )
        end = time.time()
        self.assertFalse(response)
        self.assertNotEqual(self.counter, 10)
        self.assertGreater(end - start, 10)

    @pytest.mark.timeout(timeout=2500, method="thread")
    def test_get_traces_infinite_timeout_pass(self):
        time.sleep(5)
        def num_is_10(trace):
            time.sleep(1.5)
            self.counter += 1
            assert self.counter == 10
        start = time.time()
        response = self.iatk.retry_get_trace_tree_until(
            tracing_header=self.xray_trace_header,
            assertion_fn=num_is_10,
            timeout_seconds=0,
        )
        end = time.time()
        self.assertTrue(response)
        self.assertEqual(self.counter, 10)
        self.assertGreater(end - start, 15)

    def test_invalid_traceid_fail(self):
        def num_is_10(trace):
            self.counter = random.randrange(0,10)
            assert self.counter == 10
        with pytest.raises(IatkException) as e:
            self.iatk.retry_get_trace_tree_until(
                tracing_header="test",
                assertion_fn=num_is_10,
                timeout_seconds=10,
            )
        self.assertNotEqual(self.counter, 10)
        self.assertIn("error while getting trace_id from", str(e.value))

    def test_condition_not_function_error(self):
        with pytest.raises(TypeError) as e:
            self.iatk.retry_get_trace_tree_until(
                tracing_header=self.xray_trace_header,
                assertion_fn=0,
                timeout_seconds=10,
            )
            self.assertNotEqual(self.counter, 10)
        self.assertEqual(str(e.value), "condition is not a callable function")

    def test_retry_trace_not_found(self):
        def num_is_5(trace):
            assert random.randrange(0,5) == 5
        start = time.time()
        response = self.iatk.retry_get_trace_tree_until(
            tracing_header="Root=1-652850da-255d5ae071f55e4aef339837;Sampled=1",
            assertion_fn=num_is_5,
            timeout_seconds=10,
        )
        end = time.time()
        self.assertGreaterEqual(end - start, 10)
        self.assertFalse(response)        
        
    def test_retry_trace_segment_not_found(self):
        patched_get_trace_tree = IatkException("found a segment 123456789 with no parent", 500)
        with patch("aws_iatk.AwsIatk._get_trace_tree", return_value=patched_get_trace_tree):
            def num_is_5(trace):
                assert random.randrange(0,5) == 5
            start = time.time()
            response = self.iatk.retry_get_trace_tree_until(
                tracing_header="Root=1-652850da-255d5ae071f55e4aef339837;Sampled=1",
                assertion_fn=num_is_5,
                timeout_seconds=10,
            )
            end = time.time()
            self.assertGreaterEqual(end - start, 10)
