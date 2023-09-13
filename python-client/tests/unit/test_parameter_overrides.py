# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

"""
Integration tests for zion.retry_until
"""
import logging
from unittest import TestCase
import time
import zion
import pytest
LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)

class TestZion_parameter_overrides(TestCase):
    zion = zion.Zion()
    num = 0


    def test_override_string(self):
        event_input = {"test":"before"}
        override = {"test":"after"}
        event = self.zion._construct_mock_event_parameter_overrides(event=event_input, override=override)
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_object(self):
        event_input = {"test":{"test1": "before"}}
        override = {"test":"after"}
        event = self.zion._construct_mock_event_parameter_overrides(event=event_input, override=override)
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_default(self):
        event_input = {}
        override = {"test":"after"}
        event = self.zion._construct_mock_event_parameter_overrides(event=event_input, override=override)
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_nested_object(self):
        event_input = {"test1":{"test": "before"}}
        override = {"test":"after"}
        event = self.zion._construct_mock_event_parameter_overrides(event=event_input, override=override)
        self.assertIn("test1", event)
        self.assertEqual("after", event["test1"]["test"])
        self.assertEqual(1, len(event))

    def test_override_nested_array(self):
        event_input = {"test1": [{"test" : "before"}]}
        override = {"test":"after"}
        event = self.zion._construct_mock_event_parameter_overrides(event=event_input, override=override)
        self.assertIn("test1", event)
        self.assertEqual("after", event["test1"][0]["test"])
        self.assertEqual(1, len(event))

    def test_override_invalid_type_function(self):
        def test_func():
            return 5
        with pytest.raises(zion.ZionException) as e: 
            event_input = {"test":"before"}
            override = {"test": test_func}
            event = self.zion._construct_mock_event_parameter_overrides(event=event_input, override=override)
        self.assertEqual(str(e.value), "json does not support non-serializable value provided such as class instances or functions")

    def test_override_invalid_type_class(self):
        class test_class(object):
            pass
        with pytest.raises(zion.ZionException) as e: 
            event_input = {"test":"before"}
            override = {"test": test_class}
            event = self.zion._construct_mock_event_parameter_overrides(event=event_input, override=override)
        self.assertEqual(str(e.value), "json does not support non-serializable value provided such as class instances or functions")
           