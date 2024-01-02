# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Integration tests for aws_iatk.retry_until
"""
import logging
from unittest import TestCase
import time
import aws_iatk
import pytest
import random
from unittest.mock import patch
LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)

class TestIatk_retry_until_timeout(TestCase):
    iatk = aws_iatk.AwsIatk()
    num = 0


    def test_retry_timeout_is_default_fail(self):
        def num_is_ten(val):
            assert val == 10
        @self.iatk.retry_until(assertion_fn=num_is_ten)
        def num_add_one_slow():
            self.num = self.num + 1
            return self.num
        start = time.time()
        response = num_add_one_slow()
        end = time.time()
        self.assertGreater(end - start, 10)
        self.assertNotEqual(self.num, 10)
        self.assertFalse(response)

    def test_retry_timeout_is_default_pass(self):
        def num_is_five(val):
            assert val == 5
        @self.iatk.retry_until(assertion_fn=num_is_five)
        def num_add_one():
            self.num = self.num + 1
            return self.num
        start = time.time()
        response = num_add_one()
        end = time.time()
        self.assertLess(end - start, 5)
        self.assertEqual(self.num, 5)
        self.assertTrue(response)

    @pytest.mark.timeout(timeout=2500, method="thread")
    def test_retry_timeout_is_infinite(self):
        def num_is_ten(val):
            assert val == 10
        @self.iatk.retry_until(assertion_fn=num_is_ten, timeout=0)
        def num_add_one_slow():
            time.sleep(2)
            self.num = self.num + 1
            return self.num
        start = time.time()
        response = num_add_one_slow()
        end = time.time()
        self.assertGreater(end - start, 10)
        self.assertEqual(self.num, 10)
        self.assertTrue(response)

    def test_retry_condition_not_function_error(self):
        with pytest.raises(TypeError) as e:  
            @self.iatk.retry_until(assertion_fn=0, timeout=5)
            def num_add_one():
                self.num = self.num + 1
                return self.num 
            num_add_one()
            self.assertEqual(e, "condition is not a callable function")
    
    def test_retry_timeout_wrong_type_error(self):
        def num_is_ten(val):
            assert val == 10
        with pytest.raises(TypeError) as e:  
            @self.iatk.retry_until(assertion_fn=num_is_ten, timeout="test")
            def num_add_one():
                self.num = self.num + 1
                return self.num 
            num_add_one()
            self.assertEqual(e, "timeout must be an int or float")

    def test_retry_timeout_less_than_zero_error(self):
        def num_is_ten(val):
            assert val == 10
        with pytest.raises(ValueError) as e:  
            @self.iatk.retry_until(assertion_fn=num_is_ten, timeout=-1)
            def num_add_one():
                self.num = self.num + 1
                return self.num 
            num_add_one()
            self.assertEqual(e, "timeout must not be a negative value")

    def test_retry_should_pass(self):
        def num_is_five(val):
            assert val == 5
        @self.iatk.retry_until(assertion_fn=num_is_five, timeout=5)
        def num_add_one():
            self.num = self.num + 1
            return self.num
        response = num_add_one()
        self.assertEqual(self.num, 5)
        self.assertTrue(response)

    def test_retry_should_timeout(self):
        start = time.time()
        def num_is_negative(val):
            assert val < 0
        @self.iatk.retry_until(assertion_fn=num_is_negative, timeout=5)
        def num_add_one():
            self.num = self.num + 1
            return self.num
        response = num_add_one()
        end = time.time()
        self.assertGreater(self.num, 0)
        self.assertFalse(response)
        self.assertGreaterEqual(end - start, 5)

    def test_retry_should_pass_multiple_arguments(self):
        def num_is_five(val):
            assert val == 5
        @self.iatk.retry_until(assertion_fn=num_is_five, timeout=5)
        def num_add_one(dummy, dummy1, dummy2="dummy2"):
            self.num = self.num + 1
            return self.num
        response = num_add_one(0, 0, dummy2="test")
        self.assertEqual(self.num, 5)
        self.assertTrue(response)

    def test_retry_should_timeout_multiple_arguments(self):
        start = time.time()
        def num_is_negative(val):
            assert val < 0
        @self.iatk.retry_until(assertion_fn=num_is_negative, timeout=5)
        def num_add_one(dummy, dummy1, dummy2="dummy2"):
            self.num = self.num + 1
            return self.num
        response = num_add_one(0, 0, dummy2="test")
        end = time.time()
        self.assertGreater(self.num, 0)
        self.assertFalse(response)
        self.assertGreaterEqual(end - start, 5)

    def test_retry_trace_segment_not_found(self):
        patched_get_trace_tree = aws_iatk.IatkException("found a segment 123456789 with no parent", 500)
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
            self.assertFalse(response)