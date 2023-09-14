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

class TestZion_eventbridge_context(TestCase):
    zion = zion.Zion()
    num = 0


    def test_override_string(self):
        def override(event) :
            event["test"] = "after"
            return event
        event_input = {"test":"before"}
        event = self.zion._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_object(self):
        def override(event) :
            event["test"] = "after"
            return event
        event_input = {"test":{"test1": "before"}}
        event = self.zion._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_default(self):
        def override(event) :
            event["test"] = "after"
            return event
        event_input = {}
        event = self.zion._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_invalid_type_function(self):
        def override(event) :
            def test_func():
                pass
            event["test"] = test_func
            return event
        with pytest.raises(zion.ZionException) as e: 
            event_input = {"test":"before"}
            event = self.zion._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual(str(e.value), "json does not support non-serializable value provided such as class instances or functions")

    def test_override_invalid_type_class(self):
        def override(event) :
            class test_class(object):
                pass
            event["test"] = test_class
            return event
        with pytest.raises(zion.ZionException) as e: 
            event_input = {"test":"before"}
            event = self.zion._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual(str(e.value), "json does not support non-serializable value provided such as class instances or functions")

    def test_override_invalid_type_return(self):
        def override(event) :
            event["test"] = "after"
        with pytest.raises(zion.ZionException) as e: 
            event_input = {"test":"before"}
            event = self.zion._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual(str(e.value), "event is empty, make sure function returns a valid event")
           