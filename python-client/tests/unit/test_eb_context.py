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
from aws_iatk import context_generation
LOG = logging.getLogger(__name__)
LOG.setLevel(logging.DEBUG)

class TestIatk_eventbridge_context(TestCase):
    iatk = aws_iatk.AwsIatk()
    num = 0


    def test_override_string(self):
        def override(event) :
            event["test"] = "after"
            return event
        event_input = {"test":"before"}
        event = self.iatk._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_object(self):
        def override(event) :
            event["test"] = "after"
            return event
        event_input = {"test":{"test1": "before"}}
        event = self.iatk._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_default(self):
        def override(event) :
            event["test"] = "after"
            return event
        event_input = {}
        event = self.iatk._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual("after", event["test"])
        self.assertEqual(1, len(event))

    def test_override_invalid_type_return(self):
        def override(event) :
            event["test"] = "after"
        with pytest.raises(aws_iatk.IatkException) as e: 
            event_input = {"test":"before"}
            event = self.iatk._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual(str(e.value), "event is empty, make sure function returns a valid event")
        self.assertEqual(e.value.error_code, 404)

    def test_override_invalid_type(self):
        def override(event) :
            def test_func():
                return 5
            event["test"] = test_func
            return event
        with pytest.raises(aws_iatk.IatkException) as e: 
            event_input = {"test":"before"}
            event = self.iatk._apply_contexts(generated_event=event_input, callable_contexts=[override])
        self.assertEqual(str(e.value), "context applier override returns a non-JSON-serializable result")
        self.assertEqual(e.value.error_code, 400)
    
    def test_default_eb_context(self):
        event_input = {}
        event = self.iatk._apply_contexts(generated_event=event_input, callable_contexts=[context_generation.eventbridge_event_context])
        context = ["version", "id", "account", "time", "detail-type", "source", "resources", "region"]
        self.assertEqual(len(context), len(event))
        for key in context:
            self.assertIn(key, event)
        self.assertEqual(12, len(event["account"]))

    def test_eb_context_existing(self):
        event_input = {"version": "5", "account" : "123123123123", "region": "us-west-10000"}
        event = self.iatk._apply_contexts(generated_event=event_input, callable_contexts=[context_generation.eventbridge_event_context])
        context = ["version", "id", "account", "time", "detail-type", "source", "resources", "region"]
        self.assertEqual(len(context), len(event))
        for key in context:
            self.assertIn(key, event)
        self.assertEqual("5", event["version"])
        self.assertEqual("123123123123", event["account"])
        self.assertEqual("us-west-10000", event["region"])

    def test_eb_context_doesnt_erase(self):
        event_input = {"testing": 1}

        event = self.iatk._apply_contexts(generated_event=event_input, callable_contexts=[context_generation.eventbridge_event_context])
        context = ["version", "id", "account", "time", "detail-type", "source", "resources", "region"]
        self.assertEqual(len(context) + 1, len(event))
        for key in context:
            self.assertIn(key, event)
        self.assertEqual(1, event["testing"])
           