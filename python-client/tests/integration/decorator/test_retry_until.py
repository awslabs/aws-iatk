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

LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)
boto3.set_stream_logger(name="zion", level=logging.DEBUG)

class TestZion_retry_until_timeout(TestCase):
    zion = zion.Zion
    num = 0
    status = ""
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
        except Exception as e:
            LOG.debug(e)

    
    @classmethod
    def tearDownClass(cls) -> None:
        LOG.debug("remove lambda function")
        cls.lambda_client.delete_function(
            FunctionName=cls.lambda_function_name
        )

    def test_retry_should_pass(self):
        def num_is_ten(val):
            return val == 10
        @self.zion.retry_until(condition=num_is_ten, timeout=5)
        def num_add_one():
            self.num = self.num + 1
            return self.num
        response = num_add_one()
        self.assertEqual(self.num, 10)
        self.assertTrue(response)

    def test_retry_should_timeout(self):
        start = time.time()
        def num_is_negative(val):
            return val < 0
        @self.zion.retry_until(condition=num_is_negative, timeout=5)
        def num_add_one():
            self.num = self.num + 1
            return self.num
        response = num_add_one()
        end = time.time()
        self.assertGreater(self.num, 0)
        self.assertFalse(response)
        self.assertGreaterEqual(end - start, 5)

    def test_retry_should_pass_multiple_arguments(self):
        def num_is_ten(val):
            return val == 10
        @self.zion.retry_until(condition=num_is_ten, timeout=5)
        def num_add_one(dummy, dummy1, dummy2="dummy2"):
            self.num = self.num + 1
            return self.num
        response = num_add_one(0, 0, dummy2="test")
        self.assertEqual(self.num, 10)
        self.assertTrue(response)

    def test_retry_should_timeout_multiple_arguments(self):
        start = time.time()
        def num_is_negative(val):
            return val < 0
        @self.zion.retry_until(condition=num_is_negative, timeout=5)
        def num_add_one(dummy, dummy1, dummy2="dummy2"):
            self.num = self.num + 1
            return self.num
        response = num_add_one(0, 0, dummy2="test")
        end = time.time()
        self.assertGreater(self.num, 0)
        self.assertFalse(response)
        self.assertGreaterEqual(end - start, 5)
    
    def test_lambda_status_active(self):
        def is_active(response: any) -> bool:
            self.status = response["Configuration"]["State"]
            return response["Configuration"]["State"] == "Active"
        @self.zion.retry_until(condition=is_active, timeout=5)
        def get_lambda():
            return self.lambda_client.get_function(
            FunctionName=self.lambda_function_name
            )
        response = get_lambda()
        self.assertTrue(response)
        self.assertEqual(self.status, "Active")
    
    def test_lambda_status_failed_timeout(self):
        start = time.time()
        def is_active(response: any) -> bool:
            self.status = response["Configuration"]["State"]
            return response["Configuration"]["State"] == "Failed"
        @self.zion.retry_until(condition=is_active, timeout=5)
        def get_lambda():
            return self.lambda_client.get_function(
            FunctionName=self.lambda_function_name
            )
        response = get_lambda()
        end = time.time()
        self.assertFalse(response)
        self.assertNotEqual(self.status, "Failed")
        self.assertGreaterEqual(end - start, 5)